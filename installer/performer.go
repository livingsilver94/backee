package installer

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/livingsilver94/backee"
)

type Performer func(*slog.Logger, *backee.Service) error

var (
	Setup Performer = func(log *slog.Logger, srv *backee.Service) error {
		if srv.Setup == nil || *srv.Setup == "" {
			return nil
		}
		log.Info("Running setup script")
		return runScript(*srv.Setup)
	}

	PackageInstaller Performer = func(log *slog.Logger, srv *backee.Service) error {
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

func SymlinkPerformer(repo Repository, repl Replacer) Performer {
	return func(log *slog.Logger, srv *backee.Service) error {
		if len(srv.Links) == 0 {
			return nil
		}
		log.Info("Symlinking files")
		linkDir, err := repo.LinkDir(srv.Name)
		if err != nil {
			return err
		}
		return writeFiles(srv.Links, linkDir, repl, NewSymlinkWriter(log))
	}
}

func CopyPerformer(repo Repository, repl Replacer) Performer {
	return func(log *slog.Logger, srv *backee.Service) error {
		if len(srv.Copies) == 0 {
			return nil
		}
		log.Info("Copying files")
		dataDir, err := repo.DataDir(srv.Name)
		if err != nil {
			return err
		}
		return writeFiles(srv.Copies, dataDir, repl, NewCopyWriter(repl))
	}
}

func Finalizer(repl Replacer) Performer {
	return func(log *slog.Logger, srv *backee.Service) error {
		if srv.Finalize == nil || *srv.Finalize == "" {
			return nil
		}
		log.Info("Running finalizer script")
		var script strings.Builder
		err := repl.ReplaceString(*srv.Finalize, &script)
		if err != nil {
			return err
		}
		return runScript(script.String())
	}
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
