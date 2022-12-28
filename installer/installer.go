package installer

import (
	"os"
	"os/exec"

	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

type Repository interface {
	DataDir() (string, error)
	LinkDir() (string, error)
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
	// TODO installation process.
	ilist.insert(srv.Name)
	return nil
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
