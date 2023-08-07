package installer

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/livingsilver94/backee/service"
	"golang.org/x/exp/slog"
)

type Performer func(*slog.Logger, *service.Service) error

var (
	Setup Performer = func(log *slog.Logger, srv *service.Service) error {
		if srv.Setup == nil || *srv.Setup == "" {
			return nil
		}
		log.Info("Running setup script")
		return runScript(*srv.Setup)
	}

	PackageInstaller Performer = func(log *slog.Logger, srv *service.Service) error {
		if len(srv.Packages) == 0 {
			return nil
		}
		log.Info("Installing OS packages")
		args := make([]string, 0, len(srv.PkgManager[1:])+len(srv.Packages))
		args = append(args, srv.PkgManager[1:]...)
		args = append(args, srv.Packages...)
		return runProcess(srv.PkgManager[0], args...)
	}
)

func SymlinkPerformer(repo Repository, tmpl Template) Performer {
	return func(log *slog.Logger, srv *service.Service) error {
		if len(srv.Links) == 0 {
			return nil
		}
		log.Info("Symlinking files")
		linkDir, err := repo.LinkDir(srv.Name)
		if err != nil {
			return err
		}
		return writeFiles(srv.Links, linkDir, tmpl, newSymlinkWriter(log))
	}
}

func CopyPerformer(repo Repository, tmpl Template) Performer {
	return func(log *slog.Logger, srv *service.Service) error {
		if len(srv.Copies) == 0 {
			return nil
		}
		log.Info("Copying files")
		dataDir, err := repo.DataDir(srv.Name)
		if err != nil {
			return err
		}
		return writeFiles(srv.Copies, dataDir, tmpl, newCopyWriter(tmpl))
	}
}

func writeFiles(files map[string]service.FilePath, baseDir string, tmpl Template, wr fileWriter) error {
	var dstBuf strings.Builder
	for src, dst := range files {
		err := tmpl.Execute(dst.Path, &dstBuf)
		if err != nil {
			return err
		}
		err = writeFile(service.FilePath{Path: dstBuf.String(), Mode: dst.Mode}, filepath.Join(baseDir, src), wr)
		if err != nil {
			if !errors.Is(err, fs.ErrPermission) {
				return err
			}
			// TODO: retry with privileged permissions.
		}
		dstBuf.Reset()
	}
	return nil
}

func writeFile(dst service.FilePath, src string, wr fileWriter) error {
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

func Finalizer(tmpl Template) Performer {
	return func(log *slog.Logger, srv *service.Service) error {
		if srv.Finalize == nil || *srv.Finalize == "" {
			return nil
		}
		log.Info("Running finalizer script")
		var script strings.Builder
		err := tmpl.Execute(*srv.Finalize, &script)
		if err != nil {
			return err
		}
		return runScript(script.String())
	}
}

type fileWriter interface {
	loadSource(src string) error
	writeDestination(dst string) error
}

type symlinkWriter struct {
	log *slog.Logger

	src string
}

func newSymlinkWriter(log *slog.Logger) *symlinkWriter {
	return &symlinkWriter{
		log: log,
	}
}

func (w *symlinkWriter) loadSource(src string) error {
	if _, err := os.Stat(src); err != nil { // Check that srcPath exists.
		return err
	}
	w.src = src
	return nil
}

func (w *symlinkWriter) writeDestination(dst string) error {
	err := os.Symlink(w.src, dst)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err
		}
		w.log.Info("Already exists", "path", dst)
	}
	return nil
}

type copyWriter struct {
	tmpl Template

	srcContent string
}

func newCopyWriter(tmpl Template) *copyWriter {
	return &copyWriter{
		tmpl: tmpl,
	}
}

func (w *copyWriter) loadSource(src string) error {
	content, err := os.ReadFile(src)
	w.srcContent = string(content)
	return err
}

func (w *copyWriter) writeDestination(dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	buff := bufio.NewWriter(dstFile)
	err = w.tmpl.Execute(w.srcContent, buff)
	if err != nil {
		return err
	}
	return buff.Flush()
}

func runProcess(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type UnixID struct {
	UID uint32
	GID uint32
}

func PathOwner(path string) (UnixID, error) {
	return PathOwnerFS(os.DirFS(path), ".")
}

func parentPathOwner(path string) (UnixID, error) {
	for {
		if len(path) == 1 {
			return UnixID{}, fmt.Errorf("parent directory of %s: %w", path, fs.ErrNotExist)
		}
		id, err := PathOwner(path)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return UnixID{}, err
			}
			path = filepath.Dir(path)
			continue
		}
		return id, nil
	}
}
