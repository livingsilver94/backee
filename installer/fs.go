// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package installer

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/livingsilver94/backee/privilege"
	"github.com/livingsilver94/backee/service"
)

func init() {
	privilege.RegisterInterfaceImpl(&SymlinkWriter{})
	privilege.RegisterInterfaceImpl(&CopyWriter{})
	privilege.RegisterInterfaceImpl(privilegedPathWriter{})
}

type FileWriter interface {
	LoadSource(src string) error
	WriteDestination(dst string) error
}

func WritePath(dst service.FilePath, src string, wr FileWriter) error {
	err := wr.LoadSource(src)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(dst.Path), 0755)
	if err != nil {
		return err
	}
	err = wr.WriteDestination(dst.Path)
	if err != nil {
		return err
	}
	if dst.Mode != 0 {
		return os.Chmod(dst.Path, fs.FileMode(dst.Mode))
	}
	return nil
}

func WritePathPrivileged(dst service.FilePath, src string, wr FileWriter) error {
	var r privilege.Runner = privilegedPathWriter{Dst: dst, Src: src, Wr: wr}
	return privilege.Run(r)
}

type SymlinkWriter struct {
	srcPath string
}

func NewSymlinkWriter() *SymlinkWriter {
	return &SymlinkWriter{}
}

func (w *SymlinkWriter) GobEncode() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(w.srcPath)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (w *SymlinkWriter) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&w.srcPath)
	if err != nil {
		return err
	}
	return nil
}

func (w *SymlinkWriter) LoadSource(src string) error {
	if _, err := os.Stat(src); err != nil { // Check that srcPath exists.
		return err
	}
	w.srcPath = src
	return nil
}

func (w *SymlinkWriter) WriteDestination(dst string) error {
	err := os.Symlink(w.srcPath, dst)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err
		}
		eq, errEq := w.isSymlinkEqual(dst)
		if errEq != nil {
			return errEq
		}
		if !eq {
			return err
		}
	}
	return nil
}

func (w *SymlinkWriter) isSymlinkEqual(dst string) (bool, error) {
	eq, err := filepath.EvalSymlinks(dst)
	if err != nil {
		return false, err
	}
	return eq == w.srcPath, nil
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

func (w *CopyWriter) LoadSource(src string) error {
	content, err := os.ReadFile(src)
	w.srcContent = string(content)
	return err
}

func (w *CopyWriter) WriteDestination(dst string) error {
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
	Wr  FileWriter
}

func (p privilegedPathWriter) RunPrivileged() error {
	return WritePath(p.Dst, p.Src, p.Wr)
}
