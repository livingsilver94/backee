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
	"text/template"

	"github.com/go-logr/logr"
	"github.com/livingsilver94/backee/service"
)

type Performer func(logr.Logger, *service.Service) error

var (
	Setup Performer = func(log logr.Logger, srv *service.Service) error {
		if srv.Setup == nil || *srv.Setup == "" {
			return nil
		}
		log.Info("Running setup script")
		return runScript(*srv.Setup)
	}

	PackageInstaller Performer = func(log logr.Logger, srv *service.Service) error {
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

func SymlinkPerformer(repo Repository) Performer {
	return func(log logr.Logger, srv *service.Service) error {
		if len(srv.Links) == 0 {
			return nil
		}
		log.Info("Symlinking files")
		linkDir, err := repo.LinkDir(srv.Name)
		if err != nil {
			return err
		}
		wr := symlinkWriter{log: log}
		return writeFilePaths(srv.Links, linkDir, &wr)
	}
}

func CopyPerformer(repo Repository, vars map[string]string) Performer {
	return func(log logr.Logger, srv *service.Service) error {
		if len(srv.Copies) == 0 {
			return nil
		}
		log.Info("Copying files")
		dataDir, err := repo.DataDir(srv.Name)
		if err != nil {
			return err
		}
		wr := copyWriter{variables: vars}
		return writeFilePaths(srv.Copies, dataDir, &wr)
	}
}

func Finalizer(repo Repository, vars map[string]string) Performer {
	return func(log logr.Logger, srv *service.Service) error {
		if srv.Finalize == nil || *srv.Finalize == "" {
			return nil
		}
		log.Info("Running finalizer script")
		tmpl, err := template.New("finalizer").Parse(*srv.Finalize)
		if err != nil {
			return err
		}
		var script strings.Builder
		err = tmpl.Execute(&script, vars)
		if err != nil {
			return err
		}
		return runScript(script.String())
	}
}

type fileWriter interface {
	readSource(src string) error
	writeDestination(dst string) error
}

func writeFilePaths(paths map[string]service.FilePath, srcBase string, writer fileWriter) error {
	for srcFile, param := range paths {
		dstPath := ReplaceEnvVars(param.Path)
		srcPath := filepath.Join(srcBase, srcFile)

		err := writer.readSource(srcPath)
		if err != nil {
			return err
		}
		owner, err := parentPathOwner(dstPath)
		if err != nil {
			return err
		}
		err = RunAsUnixID(
			func() error { return writeFilePath(dstPath, srcPath, param.Mode, writer) },
			owner,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeFilePath(dst, src string, mode uint16, writer fileWriter) error {
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}
	err = writer.writeDestination(dst)
	if err != nil {
		return err
	}
	if mode != 0 {
		err := os.Chmod(dst, fs.FileMode(mode))
		if err != nil {
			return err
		}
	}
	return nil
}

type symlinkWriter struct {
	log logr.Logger
	src string
}

func (w *symlinkWriter) readSource(src string) error {
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
	variables map[string]string
	tmpl      *template.Template
}

func (w *copyWriter) readSource(src string) error {
	tmpl, err := template.ParseFiles(src)
	if err != nil {
		return err
	}
	w.tmpl = tmpl
	return nil
}

func (w *copyWriter) writeDestination(dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	buff := bufio.NewWriter(dstFile)
	err = w.tmpl.Execute(buff, w.variables)
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
