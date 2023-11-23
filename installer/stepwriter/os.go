// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package stepwriter

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/privilege"
	"github.com/livingsilver94/backee/service"
)

func init() {
	privilege.RegisterInterfaceImpl(symlinkWriter{})
	privilege.RegisterInterfaceImpl(fileCopyWriter{})
	privilege.RegisterInterfaceImpl(privilegedPathWriter{})
}

type OS struct{}

func (OS) Setup(script string) error {
	return runScript(script)
}

func (OS) InstallPackages(fullCmd []string) error {
	return runProcess(fullCmd[0], fullCmd[1:]...)
}

func (OS) SymlinkFile(dst service.FilePath, src string) error {
	return writePossiblyPrivilegedPath(dst, &symlinkWriter{SrcPath: src})
}

func (OS) CopyFile(dst service.FilePath, src installer.FileCopy) error {
	return writePossiblyPrivilegedPath(dst, &fileCopyWriter{FileCopy: src})
}

func (OS) Finalize(script string) error {
	return runScript(script)
}

type fileWriter interface {
	writeFile(dst string) error
}

func writePath(dst service.FilePath, wr fileWriter) error {
	err := os.MkdirAll(filepath.Dir(dst.Path), 0755)
	if err != nil {
		return err
	}
	err = wr.writeFile(dst.Path)
	if err != nil {
		return err
	}
	if dst.Mode != 0 {
		return os.Chmod(dst.Path, fs.FileMode(dst.Mode))
	}
	return nil
}

func writePathPrivileged(dst service.FilePath, wr fileWriter) error {
	var r privilege.Runner = privilegedPathWriter{Dst: dst, Wr: wr}
	return privilege.Run(r)
}

func writePossiblyPrivilegedPath(dst service.FilePath, wr fileWriter) error {
	err := writePath(dst, wr)
	if err != nil {
		if !errors.Is(err, fs.ErrPermission) {
			return err
		}
		err = writePathPrivileged(dst, wr)
		if err != nil {
			return err
		}
	}
	return nil
}

type symlinkWriter struct {
	SrcPath string
}

func (w symlinkWriter) writeFile(dst string) error {
	err := os.Symlink(w.SrcPath, dst)
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

func (w *symlinkWriter) isSymlinkEqual(dst string) (bool, error) {
	eq, err := filepath.EvalSymlinks(dst)
	if err != nil {
		return false, err
	}
	return eq == w.SrcPath, nil
}

type fileCopyWriter struct {
	FileCopy installer.FileCopy
}

func (w fileCopyWriter) writeFile(dst string) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = w.FileCopy.WriteTo(file)
	return err
}

type privilegedPathWriter struct {
	Dst service.FilePath
	Wr  fileWriter
}

func (p privilegedPathWriter) RunPrivileged() error {
	return writePath(p.Dst, p.Wr)
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
