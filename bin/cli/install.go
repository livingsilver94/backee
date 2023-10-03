package cli

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/livingsilver94/backee"
	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/secret"
)

type keepassXC struct {
	Path     string `env:"KEEPASSXC_PATH" help:"KeepassXC database path."`
	Password string `env:"KEEPASSXC_PASSWORD" help:"KeepassXC database password."`
}

type install struct {
	Directory  string    `short:"C" type:"existingdir" help:"Change the base directory."`
	KeepassXC  keepassXC `embed:"" prefix:"keepassxc."`
	PkgManager []string  `name:"pkgmanager" help:"Override the package manager command for services."`
	Variant    string    `help:"Specify the system variant."`

	Services []string `arg:"" optional:"" help:"Services to install. Pass none to install all services in the base directory."`
}

const (
	installedListFilename = "installed.txt"
)

func (in *install) Run() error {
	if in.Directory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		in.Directory = cwd
	}
	var fileList *os.File
	defer func() {
		fileList.Close()
	}()

	rep := repo.NewFSRepoVariant(repo.NewOSFS(in.Directory), in.Variant)
	srv, err := in.services(rep)
	if err == nil && len(srv) == 0 {
		err = errors.New("no services found")
	}
	if err != nil {
		return err
	}
	ins := in.installer(rep, &fileList)
	for _, s := range srv {
		err := ins.Install(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (in *install) services(rep repo.FSRepo) ([]*backee.Service, error) {
	if len(in.PkgManager) != 0 {
		backee.DefaultPkgManager = in.PkgManager
	}

	if len(in.Services) == 0 {
		return rep.AllServices()
	}
	services := make([]*backee.Service, 0, len(in.Services))
	for _, name := range in.Services {
		srv, err := rep.Service(name)
		if err != nil {
			return nil, err
		}
		services = append(services, srv)
	}
	return services, nil
}

func (in *install) installer(rep repo.FSRepo, fileList **os.File) installer.Installer {
	var (
		list installer.List
		err  error
	)
	*fileList, err = os.OpenFile(installedListFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		slog.Error(err.Error() + "Failed opening the installation list file. Continuing without populating it")
		list = installer.NewList()
	} else {
		list, err = installer.NewListCached(*fileList)
		if err != nil {
			slog.Error("Failed reading previous installation list")
		}
	}

	vrs := installer.NewVariables()
	vrs.Common = envVars()
	if in.KeepassXC.Path != "" {
		kee := secret.NewKeepassXC(in.KeepassXC.Path, in.KeepassXC.Password)
		vrs.RegisterStore(backee.VarKind("keepassxc"), kee)
	}

	return installer.New(rep, installer.WithVariables(vrs), installer.WithList(list))
}

// envVars returns a map of environment variables.
func envVars() map[string]string {
	env := os.Environ()
	envMap := make(map[string]string, len(env))
	for _, keyVal := range env {
		key, val, _ := strings.Cut(keyVal, "=")
		envMap[key] = val
	}
	return envMap
}
