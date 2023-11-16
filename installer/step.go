// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

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

	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

type Step interface {
	Run(*slog.Logger, *service.Service) error
}

type Setup struct{}

func (Setup) Run(log *slog.Logger, srv *service.Service) error {
	if srv.Setup == nil || *srv.Setup == "" {
		return nil
	}
	log.Info("Running setup script")
	return runScript(*srv.Setup)
}

type OSPackages struct{}

func (OSPackages) Run(log *slog.Logger, srv *service.Service) error {
	if len(srv.Packages) == 0 {
		return nil
	}
	log.Info("Installing OS packages")
	args := make([]string, 0, len(srv.PkgManager[1:])+len(srv.Packages))
	args = append(args, srv.PkgManager[1:]...)
	args = append(args, srv.Packages...)
	return runProcess(srv.PkgManager[0], args...)
}

type Symlinks struct {
	repo repo.Repo
	repl Template
}

func NewSymlinks(repo repo.Repo, repl Template) Symlinks {
	return Symlinks{
		repo: repo,
		repl: repl,
	}
}

func (s Symlinks) Run(log *slog.Logger, srv *service.Service) error {
	if len(srv.Links) == 0 {
		return nil
	}
	log.Info("Symlinking files")
	linkDir, err := s.repo.LinkDir(srv.Name)
	if err != nil {
		return err
	}
	return writeFiles(srv.Links, linkDir, s.repl, NewSymlinkWriter())
}

type Copies struct {
	repo repo.Repo
	repl Template
}

func NewCopies(repo repo.Repo, repl Template) Copies {
	return Copies{
		repo: repo,
		repl: repl,
	}
}

func (c Copies) Run(log *slog.Logger, srv *service.Service) error {
	if len(srv.Copies) == 0 {
		return nil
	}
	log.Info("Copying files")
	dataDir, err := c.repo.DataDir(srv.Name)
	if err != nil {
		return err
	}
	return writeFiles(srv.Copies, dataDir, c.repl, NewCopyWriter(c.repl))
}

type Finalization struct {
	repl Template
}

func NewFinalization(repl Template) Finalization {
	return Finalization{
		repl: repl,
	}
}

func (f Finalization) Run(log *slog.Logger, srv *service.Service) error {
	if srv.Finalize == nil || *srv.Finalize == "" {
		return nil
	}
	log.Info("Running finalizer script")
	var script strings.Builder
	err := f.repl.ReplaceString(*srv.Finalize, &script)
	if err != nil {
		return err
	}
	return runScript(script.String())
}

func runProcess(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writeFiles(files map[string]service.FilePath, baseDir string, repl Template, wr FileWriter) error {
	var resolvedDst strings.Builder
	for srcFile, dstFile := range files {
		err := repl.ReplaceString(dstFile.Path, &resolvedDst)
		if err != nil {
			return err
		}
		err = writePath(
			service.FilePath{Path: resolvedDst.String(), Mode: dstFile.Mode},
			filepath.Join(baseDir, srcFile),
			wr,
		)
		if err != nil {
			return err
		}
		resolvedDst.Reset()
	}
	return nil
}

func writePath(dst service.FilePath, src string, wr FileWriter) error {
	err := WritePath(dst, src, wr)
	if err != nil {
		if !errors.Is(err, fs.ErrPermission) {
			return err
		}
		err = WritePathPrivileged(dst, src, wr)
		if err != nil {
			return err
		}
	}
	return nil
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
