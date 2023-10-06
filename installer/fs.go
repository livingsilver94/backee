package installer

import (
	"bufio"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/livingsilver94/backee"
)

type FileWriter interface {
	loadSource(src string) error
	writeDestination(dst string) error
}

func WritePath(dst backee.FilePath, src string, wr FileWriter) error {
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
	repl       Replacer
	srcContent string
}

func NewCopyWriter(repl Replacer) *CopyWriter {
	return &CopyWriter{
		repl: repl,
	}
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

func writeFiles(files map[string]backee.FilePath, baseDir string, repl Replacer, wr FileWriter) error {
	var resolvedDst strings.Builder
	for srcFile, dstFile := range files {
		err := repl.ReplaceString(dstFile.Path, &resolvedDst)
		if err != nil {
			return err
		}
		err = WritePath(
			backee.FilePath{Path: resolvedDst.String(), Mode: dstFile.Mode},
			filepath.Join(baseDir, srcFile),
			wr,
		)
		if err != nil {
			if !errors.Is(err, fs.ErrPermission) {
				return err
			}
			// TODO: retry with privileged permissions.
		}
		resolvedDst.Reset()
	}
	return nil
}