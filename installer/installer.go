package installer

import (
	"os"

	"github.com/livingsilver94/backee"
	"github.com/livingsilver94/backee/repo"
	"golang.org/x/exp/slog"
)

type Repository interface {
	DataDir(srvName string) (string, error)
	LinkDir(srvName string) (string, error)
	ResolveDeps(srv *backee.Service) (repo.DepGraph, error)
}

type VarStore interface {
	Value(varName string) (varValue string, err error)
}

const (
	installedListFilename = "installed.txt"
)

type Installer struct {
	repository Repository
	variables  Variables
}

func New(repository Repository, options ...Option) Installer {
	i := Installer{
		repository: repository,
		variables:  NewVariables(),
	}
	for _, option := range options {
		option(&i)
	}
	return i
}

func (inst *Installer) Install(srv *backee.Service) error {
	ilistFile, err := os.OpenFile(installedListFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	var list InstallList
	if err != nil {
		slog.Error(err.Error() + ". Continuing without populating the installation list")
		list = NewInstallList(nil)
	} else {
		defer ilistFile.Close()
		list = NewInstallList(ilistFile)
	}
	return inst.installHierarchy(srv, &list)
}

func (inst *Installer) installHierarchy(srv *backee.Service, list *InstallList) error {
	if srv == nil {
		return nil
	}

	depGraph, err := inst.repository.ResolveDeps(srv)
	if err != nil {
		return err
	}
	for level := depGraph.Depth() - 1; level >= 0; level-- {
		for _, dep := range depGraph.Level(level).Slice() {
			err := inst.installSingle(dep, list)
			if err != nil {
				return err
			}
		}
	}
	return inst.installSingle(srv, list)
}

func (inst *Installer) installSingle(srv *backee.Service, ilist *InstallList) error {
	log := slog.Default().WithGroup(srv.Name)
	if ilist.Contains(srv.Name) {
		log.Info("Already installed")
		return nil
	}
	err := inst.cacheVars(srv)
	if err != nil {
		return err
	}
	repl := NewReplacer(srv.Name, inst.variables)
	performers := []Performer{
		Setup,
		PackageInstaller,
		SymlinkPerformer(inst.repository, repl),
		CopyPerformer(inst.repository, repl),
		Finalizer(repl),
	}
	for _, perf := range performers {
		err := perf(log, srv)
		if err != nil {
			return err
		}
	}
	ilist.Insert(srv.Name)
	return nil
}

func (inst *Installer) cacheVars(srv *backee.Service) error {
	datadir, err := inst.repository.DataDir(srv.Name)
	if err != nil {
		return err
	}
	err = inst.variables.Insert(
		srv.Name,
		backee.VarDatadir,
		backee.VarValue{Kind: backee.ClearText, Value: datadir})
	if err != nil {
		return err
	}
	return inst.variables.InsertMany(srv.Name, srv.Variables)
}

type Option func(*Installer)

func WithVariables(v Variables) Option {
	return func(i *Installer) {
		i.variables = v
	}
}
