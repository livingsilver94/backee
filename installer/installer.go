package installer

import (
	"log/slog"

	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

type Repository interface {
	DataDir(srvName string) (string, error)
	LinkDir(srvName string) (string, error)
	ResolveDeps(srv *service.Service) (repo.DepGraph, error)
}

type Installer struct {
	repository Repository
	variables  repo.Variables
	list       List
}

func New(repository Repository, options ...Option) Installer {
	i := Installer{
		repository: repository,
		variables:  repo.NewVariables(),
		list:       NewList(),
	}
	for _, option := range options {
		option(&i)
	}
	return i
}

func (inst *Installer) Install(srv *service.Service) error {
	if srv == nil {
		return nil
	}

	depGraph, err := inst.repository.ResolveDeps(srv)
	if err != nil {
		return err
	}
	for level := depGraph.Depth() - 1; level >= 0; level-- {
		for _, dep := range depGraph.Level(level).Slice() {
			err := inst.installSingle(dep)
			if err != nil {
				return err
			}
		}
	}
	return inst.installSingle(srv)
}

func (inst *Installer) installSingle(srv *service.Service) error {
	log := slog.Default().WithGroup(srv.Name)
	if inst.list.Contains(srv.Name) {
		log.Info("Already installed")
		return nil
	}
	err := inst.cacheVars(srv)
	if err != nil {
		return err
	}
	repl := NewReplacer(srv.Name, inst.variables)
	steps := []Step{
		Setup{},
		OSPackages{},
		NewSymlinks(inst.repository, repl),
		NewCopies(inst.repository, repl),
		NewFinalization(repl),
	}
	for _, step := range steps {
		err := step.Run(log, srv)
		if err != nil {
			return err
		}
	}
	inst.list.Insert(srv.Name)
	return nil
}

func (inst *Installer) cacheVars(srv *service.Service) error {
	datadir, err := inst.repository.DataDir(srv.Name)
	if err != nil {
		return err
	}
	err = inst.variables.Insert(
		srv.Name,
		service.VarDatadir,
		service.VarValue{Kind: service.ClearText, Value: datadir})
	if err != nil {
		return err
	}
	return inst.variables.InsertMany(srv.Name, srv.Variables)
}

type Option func(*Installer)

func WithVariables(v repo.Variables) Option {
	return func(i *Installer) {
		i.variables = v
	}
}

func WithList(li List) Option {
	return func(i *Installer) {
		i.list = li
	}
}
