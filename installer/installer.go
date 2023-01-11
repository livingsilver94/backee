package installer

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

type Repository interface {
	DataDir(srvName string) (string, error)
	LinkDir(srvName string) (string, error)
	ResolveDeps(srv *service.Service) (repo.DepGraph, error)
}

type VarStore interface {
	Value(storeValue string) (varValue string, err error)
}

const (
	installedListFilename = "installed.txt"
)

type Installer struct {
	repository Repository
	varcache   VarCache
	logger     logr.Logger
}

func New(repository Repository, options ...Option) Installer {
	i := Installer{
		repository: repository,
		varcache:   NewVarCache(),
		logger:     logr.Discard(),
	}
	for _, option := range options {
		option(&i)
	}
	return i
}

func (inst Installer) Install(services []*service.Service) error {
	ilistFile, err := os.OpenFile(installedListFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	var list List
	if err != nil {
		list = NewList(nil)
	} else {
		defer ilistFile.Close()
		list = NewList(ilistFile)
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

func (inst Installer) install(srv *service.Service, ilist *List) error {
	if ilist.Contains(srv.Name) {
		return nil
	}
	performers := []Performer{
		Setup,
		PackageInstaller,
		SymlinkPerformer(inst.repository),
		CopyPerformer(inst.repository),
		Finalizer(inst.repository, inst.varcache),
	}
	log := inst.logger.WithName(srv.Name)
	for _, perf := range performers {
		err := perf(log, srv)
		if err != nil {
			return err
		}
	}
	ilist.Insert(srv.Name)
	return nil
}

type Option func(*Installer)

func WithLogger(lg logr.Logger) Option {
	return func(i *Installer) {
		i.logger = lg
	}
}

func WithStore(kind service.VarKind, store VarStore) Option {
	return func(i *Installer) {
		i.varcache.SetStore(kind, store)
	}
}
