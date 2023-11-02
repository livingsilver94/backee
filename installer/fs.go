package installer

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/livingsilver94/backee/privilege"
	"github.com/livingsilver94/backee/service"
)

type FileWriter interface {
	loadSource(src string) error
	writeDestination(dst string) error
}

type PrivilegedFileWriter interface {
	FileWriter
	gob.GobEncoder
}

func WritePath(dst service.FilePath, src string, wr FileWriter) error {
	err := wr.loadSource(src)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(dst.Path), 0755)
	if err != nil {
		return err
	}
	err = wr.writeDestination(dst.Path)
	if err != nil {
		return err
	}
	if dst.Mode != 0 {
		return os.Chmod(dst.Path, fs.FileMode(dst.Mode))
	}
	return nil
}

func WritePathPrivileged(dst service.FilePath, src string, wr PrivilegedFileWriter) error {
	return privilege.Run(privilegedPathWriter{Dst: dst, Src: src, Wr: wr})
}

func RegisterPrivilegedTypes() {
	privilege.RegisterInterfaceImpl(CopyWriter{})
	privilege.RegisterInterfaceImpl(privilegedPathWriter{})
}

type SymlinkWriter struct {
	log     *slog.Logger
	srcPath string
}

func NewSymlinkWriter(log *slog.Logger) *SymlinkWriter {
	return &SymlinkWriter{
		log: log,
	}
}

func (w *SymlinkWriter) loadSource(src string) error {
	if _, err := os.Stat(src); err != nil { // Check that srcPath exists.
		return err
	}
	w.srcPath = src
	return nil
}

func (w *SymlinkWriter) writeDestination(dst string) error {
	err := os.Symlink(w.srcPath, dst)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err
		}
		w.log.Info("Already exists", "path", dst)
	}
	return nil
}

type CopyWriter struct {
	repl       Template
	srcContent string
}

func NewCopyWriter(repl Template) *CopyWriter {
	return &CopyWriter{
		repl: repl,
	}
}

func (w *CopyWriter) GobEncode() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(w.repl)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(w.srcContent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (w *CopyWriter) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&w.repl)
	if err != nil {
		return err
	}
	err = dec.Decode(&w.srcContent)
	if err != nil {
		return err
	}
	return nil
}

func (w *CopyWriter) loadSource(src string) error {
	content, err := os.ReadFile(src)
	w.srcContent = string(content)
	return err
}

func (w *CopyWriter) writeDestination(dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	buff := bufio.NewWriter(dstFile)
	err = w.repl.ReplaceString(w.srcContent, buff)
	if err != nil {
		return err
	}
	return buff.Flush()
}

type privilegedPathWriter struct {
	Dst service.FilePath
	Src string
	Wr  PrivilegedFileWriter
}

func (p privilegedPathWriter) Run() error {
	return WritePath(p.Dst, p.Src, p.Wr)
}
