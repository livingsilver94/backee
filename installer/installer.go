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
	err        error
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

func (inst *Installer) Install(services []*service.Service) bool {
	if len(services) == 0 {
		inst.logger.Info("The service list is empty")
		return true
	}

	ilistFile, err := os.OpenFile(installedListFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	var list List
	if err != nil {
		inst.logger.Error(
			err,
			"Continuing without populating the installation list")
		list = NewList(nil)
	} else {
		defer ilistFile.Close()
		list = NewList(ilistFile)
	}

	for _, srv := range services {
		if !inst.installHierarchy(srv, &list) {
			return false
		}
	}
	return true
}

// Error returns the first error encountered while Installing.
func (inst *Installer) Error() error {
	return inst.err
}

func (inst *Installer) installHierarchy(srv *service.Service, list *List) bool {
	depGraph, err := inst.repository.ResolveDeps(srv)
	if err != nil {
		return inst.setError(err)
	}
	for level := depGraph.Depth() - 1; level >= 0; level-- {
		for _, dep := range depGraph.Level(level).Slice() {
			if !inst.installSingle(dep, list) {
				return false
			}
		}
	}
	return inst.installSingle(srv, list)
}

func (inst *Installer) installSingle(srv *service.Service, ilist *List) bool {
	log := inst.logger.WithName(srv.Name)
	if ilist.Contains(srv.Name) {
		log.Info("Already installed")
		return inst.setError(nil)
	}
	performers := []Performer{
		Setup,
		PackageInstaller,
		SymlinkPerformer(inst.repository),
		CopyPerformer(inst.repository, inst.varcache),
		Finalizer(inst.repository, inst.varcache),
	}
	for _, perf := range performers {
		err := perf(log, srv)
		if err != nil {
			log.Error(err, "")
			return inst.setError(err)
		}
	}
	ilist.Insert(srv.Name)
	return inst.setError(nil)
}

func (inst *Installer) setError(err error) bool {
	inst.err = err
	return err == nil
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
