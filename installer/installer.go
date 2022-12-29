package installer

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
}

func New(repository Repository) Installer {
	return Installer{
		repository: repository,
	}
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
	performers := []performer{
		inst.perform_setup,
		inst.perform_pkg_installation,
		inst.perform_link_installation,
		inst.perform_finalizer,
	}
	for _, perf := range performers {
		err := perf(srv)
		if err != nil {
			return err
		}
	}
	ilist.insert(srv.Name)
	return nil
}

func (Installer) perform_setup(srv *service.Service) error {
	if srv.Setup == nil {
		return nil
	}
	return runScript(*srv.Setup)
}

func (Installer) perform_pkg_installation(srv *service.Service) error {
	if srv.Packages == nil {
		return nil
	}
	args := make([]string, 0, len(srv.PkgManager[1:])+len(srv.Packages))
	args = append(args, srv.PkgManager[1:]...)
	args = append(args, srv.Packages...)
	return runProcess(srv.PkgManager[0], args...)
}

func (inst Installer) perform_link_installation(srv *service.Service) error {
	if srv.Links == nil {
		return nil
	}
	linkdir, err := inst.repository.LinkDir(srv.Name)
	if err != nil {
		return err
	}
	for srcFile, param := range srv.Links {
		srcPath := filepath.Join(linkdir, srcFile)
		dstPath := ReplaceEnvVars(param.Path)
		err := os.MkdirAll(filepath.Dir(dstPath), 0644)
		if err != nil {
			return err
		}
		err = os.Symlink(srcPath, dstPath)
		if err != nil {
			return err
		}
		if param.Mode != 0 {
			err := os.Chmod(dstPath, fs.FileMode(param.Mode))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (inst Installer) perform_finalizer(srv *service.Service) error {
	if srv.Finalize == nil {
		return nil
	}
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

type performer func(*service.Service) error
