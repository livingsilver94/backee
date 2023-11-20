// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package installer

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

type StepWriter interface {
	Setup(script string) error
	InstallPackages(fullCmd []string) error
	SymlinkFile(dst service.FilePath, src string) error
	CopyFile(dst service.FilePath, src FileCopy) error
	Finalize(script string) error
}

type Steps struct {
	srv *service.Service
	log *slog.Logger
	wri StepWriter
}

func NewSteps(srv *service.Service, wri StepWriter) Steps {
	return Steps{
		srv: srv,
		log: slog.Default().WithGroup(srv.Name),
		wri: wri,
	}
}

func (s Steps) Setup() error {
	if s.srv.Setup == nil || *s.srv.Setup == "" {
		return nil
	}
	s.log.Info("Running setup script")
	return s.wri.Setup(*s.srv.Setup)
}

func (s Steps) InstallPackages() error {
	if len(s.srv.Packages) == 0 {
		return nil
	}
	s.log.Info("Installing OS packages")
	return s.wri.InstallPackages(append(s.srv.PkgManager, s.srv.Packages...))
}

func (s Steps) LinkFiles(repo repo.Repo, vars repo.Variables) error {
	if len(s.srv.Links) == 0 {
		return nil
	}
	s.log.Info("Symlinking files")

	lnDir, err := repo.LinkDir(s.srv.Name)
	if err != nil {
		return err
	}
	tmpl := NewTemplate(s.srv.Name, vars)
	dest := &strings.Builder{}
	for srcFile, dstFile := range s.srv.Links {
		_, err := tmpl.ReplaceString(dstFile.Path, dest)
		if err != nil {
			return err
		}
		err = s.wri.SymlinkFile(
			service.FilePath{Path: dest.String(), Mode: dstFile.Mode},
			filepath.Join(lnDir, srcFile))
		if err != nil {
			return err
		}
		dest.Reset()
	}
	return nil
}

func (s Steps) CopyFiles(repo repo.Repo, vars repo.Variables) error {
	if len(s.srv.Copies) == 0 {
		return nil
	}
	s.log.Info("Copying files")

	dataDir, err := repo.DataDir(s.srv.Name)
	if err != nil {
		return err
	}
	tmpl := NewTemplate(s.srv.Name, vars)
	dest := &strings.Builder{}
	for srcFile, dstFile := range s.srv.Copies {
		_, err := tmpl.ReplaceString(dstFile.Path, dest)
		if err != nil {
			return err
		}
		err = s.wri.CopyFile(
			service.FilePath{Path: dest.String(), Mode: dstFile.Mode},
			FileCopy{Src: filepath.Join(dataDir, srcFile), Templ: tmpl})
		if err != nil {
			return err
		}
		dest.Reset()
	}
	return nil
}

func (s Steps) Finalize(vars repo.Variables) error {
	if s.srv.Finalize == nil || *s.srv.Finalize == "" {
		return nil
	}
	s.log.Info("Running finalizer script")
	tmpl := NewTemplate(s.srv.Name, vars)
	script := &strings.Builder{}
	_, err := tmpl.ReplaceString(*s.srv.Finalize, script)
	if err != nil {
		return err
	}
	return s.wri.Finalize(script.String())
}

type FileCopy struct {
	Src   string
	Templ Template
}

func (fc FileCopy) WriteTo(w io.Writer) (n int64, err error) {
	file, err := os.Open(fc.Src)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	cont, err := io.ReadAll(file) // TODO: reuse buffer.
	if err != nil {
		return 0, err
	}
	if isBinary(cont) {
		n, err := w.Write(cont)
		return int64(n), err
	}
	return fc.Templ.ReplaceString(unsafe.String(&cont[0], len(cont)), w)
}

func (fc FileCopy) String() string {
	buf := &bytes.Buffer{}
	_, err := fc.WriteTo(buf)
	if err != nil {
		return ""
	}
	if isBinary(buf.Bytes()) {
		return "*binary*"
	}
	return buf.String()
}

func isBinary(cont []byte) bool {
	return bytes.ContainsRune(cont, 0x0)
}
