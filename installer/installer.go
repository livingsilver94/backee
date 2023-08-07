package installer

import (
	"os"
	"strings"

	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
	"golang.org/x/exp/slog"
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
}

func New(repository Repository, options ...Option) Installer {
	i := Installer{
		repository: repository,
		varcache:   NewVariables(),
	}
	for _, option := range options {
		option(&i)
	}
	return i
}

func (inst *Installer) Install(srv *service.Service) error {
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

func (inst *Installer) installHierarchy(srv *service.Service, list *InstallList) error {
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

func (inst *Installer) installSingle(srv *service.Service, ilist *InstallList) error {
	log := slog.Default().WithGroup(srv.Name)
	if ilist.Contains(srv.Name) {
		log.Info("Already installed")
		return nil
	}
	err := inst.cacheVars(srv)
	if err != nil {
		return err
	}
	tmpl := NewTemplate(srv.Name, inst.varcache)
	tmpl.ExtraVars = environMap()
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
			return err
		}
	}
	ilist.Insert(srv.Name)
	return nil
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

type Option func(*Installer)

func WithStore(kind service.VarKind, store VarStore) Option {
	return func(i *Installer) {
		i.varcache.RegisterStore(kind, store)
	}
}

// environMap returns a map of environment variables.
func environMap() map[string]string {
	env := os.Environ()
	envMap := make(map[string]string, len(env))
	for _, keyVal := range env {
		key, val, _ := strings.Cut(keyVal, "=")
		envMap[key] = val
	}
	return envMap
}
