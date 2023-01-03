package installer

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

type Repository interface {
	DataDir(string) (string, error)
	LinkDir(string) (string, error)
	ResolveDeps(srv *service.Service) (repo.DepGraph, error)
}

const (
	installedListFilename = "installed.txt"
)

type Installer struct {
	repository Repository
	logger     logr.Logger
}

func New(repository Repository, options ...Option) Installer {
	i := Installer{
		repository: repository,
		logger:     logr.Discard(),
	}
	for _, option := range options {
		option(&i)
	}
	return i
}

func (inst Installer) Install(services []*service.Service) error {
	ilistFile, err := os.OpenFile(installedListFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	var list installedList
	if err != nil {
		list = newInstalledList(newDiscard())
	} else {
		defer ilistFile.Close()
		list = newInstalledList(ilistFile)
	}

	for _, srv := range services {
		depGraph, err := inst.repository.ResolveDeps(srv)
		if err != nil {
			return err
		}
		for level := depGraph.Depth() - 1; level >= 0; level-- {
			for _, dep := range depGraph.Level(level).List() {
				err := inst.install(dep, &list)
				if err != nil {
					return err
				}
			}
		}
		err = inst.install(srv, &list)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst Installer) install(srv *service.Service, ilist *installedList) error {
	if ilist.contains(srv.Name) {
		return nil
	}
	type performer func(logr.Logger, *service.Service) error
	performers := []performer{
		inst.perform_setup,
		inst.perform_pkg_installation,
		inst.perform_link_installation,
		inst.perform_finalizer,
	}
	log := inst.logger.WithName(srv.Name)
	for _, perf := range performers {
		err := perf(log, srv)
		if err != nil {
			return err
		}
	}
	ilist.insert(srv.Name)
	return nil
}

func (Installer) perform_setup(log logr.Logger, srv *service.Service) error {
	if srv.Setup == nil {
		return nil
	}
	log.Info("% Running setup script")
	return runScript(*srv.Setup)
}

func (Installer) perform_pkg_installation(log logr.Logger, srv *service.Service) error {
	if srv.Packages == nil {
		return nil
	}
	log.Info("% Installing OS packages")
	args := make([]string, 0, len(srv.PkgManager[1:])+len(srv.Packages))
	args = append(args, srv.PkgManager[1:]...)
	args = append(args, srv.Packages...)
	return runProcess(srv.PkgManager[0], args...)
}

func (inst Installer) perform_link_installation(log logr.Logger, srv *service.Service) error {
	if srv.Links == nil {
		return nil
	}
	log.Info("% Installing symbolic links")
	linkdir, err := inst.repository.LinkDir(srv.Name)
	if err != nil {
		return err
	}
	for srcFile, dstRawPath := range srv.Links {
		srcPath := filepath.Join(linkdir, srcFile)
		dstPath := ReplaceEnvVars(dstRawPath)

		dstDir := filepath.Dir(dstPath)
		err := inst.runAsOwner(dstDir, func() error {
			err := os.MkdirAll(dstDir, 0755)
			if err != nil {
				return err
			}
			err = os.Symlink(srcPath, dstPath)
			if err != nil {
				if !errors.Is(err, fs.ErrExist) {
					return err
				}
				log.Info("% already exists", dstPath)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst Installer) perform_finalizer(log logr.Logger, srv *service.Service) error {
	if srv.Finalize == nil {
		return nil
	}
	log.Info("% Running finalizer script")
	datadir, err := inst.repository.DataDir(srv.Name)
	if err != nil {
		return err
	}
	replacements := make([]string, 0, len(srv.Variables)*2+2)
	replacements = append(replacements, service.VarPlaceholder(service.VariableDatadir))
	replacements = append(replacements, datadir)
	for key, val := range srv.Variables {
		switch val.Kind {
		case service.ClearText:
			replacements = append(replacements, service.VarPlaceholder(key))
			replacements = append(replacements, val.Value)
		case service.Secret:
			// TODO
		}
	}

	script := strings.NewReplacer(replacements...).Replace(*srv.Finalize)
	return runScript(script)
}

func (inst Installer) runAsOwner(path string, f func() error) error {
	for {
		if len(path) == 1 {
			return fmt.Errorf("parent directory of %s: %w", path, fs.ErrNotExist)
		}
		uid, gid, err := ID(path)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			path = filepath.Dir(path)
			continue
		}
		return RunAs(f, uid, gid)
	}
}

func runProcess(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = nil
	return cmd.Run()
}

func runScript(script string) error {
	return runProcess(
		"sh",
		"-e", // Stop script on first error.
		"-c", // Run the following script string.
		script,
	)
}

type Option func(*Installer)

func WithLogger(lg logr.Logger) Option {
	return func(i *Installer) {
		i.logger = lg
	}
}
