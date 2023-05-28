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
	varcache   Variables
	logger     logr.Logger
	err        error
}

func New(repository Repository, options ...Option) Installer {
	i := Installer{
		repository: repository,
		varcache:   NewVariables(),
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
	var list InstallList
	if err != nil {
		inst.logger.Error(
			err,
			"Continuing without populating the installation list")
		list = NewInstallList(nil)
	} else {
		defer ilistFile.Close()
		list = NewInstallList(ilistFile)
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

func (inst *Installer) installHierarchy(srv *service.Service, list *InstallList) bool {
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

func (inst *Installer) installSingle(srv *service.Service, ilist *InstallList) bool {
	log := inst.logger.WithName(srv.Name)
	if ilist.Contains(srv.Name) {
		log.Info("Already installed")
		return inst.setError(nil)
	}
	err := inst.cacheVars(srv)
	if err != nil {
		return inst.setError(err)
	}
	tmpl := NewTemplate(srv, inst.varcache)
	performers := []Performer{
		Setup,
		PackageInstaller,
		SymlinkPerformer(inst.repository, tmpl),
		CopyPerformer(inst.repository, tmpl),
		Finalizer(tmpl),
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

func (inst *Installer) cacheVars(srv *service.Service) error {
	datadir, err := inst.repository.DataDir(srv.Name)
	if err != nil {
		return err
	}
	err = inst.varcache.Insert(
		srv.Name,
		service.VarDatadir,
		service.VarValue{Kind: service.ClearText, Value: datadir})
	if err != nil {
		return err
	}
	return inst.varcache.InsertMany(srv.Name, srv.Variables)
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
		i.varcache.RegisterStore(kind, store)
	}
}
