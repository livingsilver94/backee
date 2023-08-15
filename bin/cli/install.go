package cli

import (
	"errors"
	"os"

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

func (in *install) Run() error {
	if in.Directory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		in.Directory = cwd
	}
	rep := repo.NewFSRepoVariant(repo.NewOSFS(in.Directory), in.Variant)
	srv, err := in.services(rep)
	if err == nil && len(srv) == 0 {
		err = errors.New("no services found")
	}
	if err != nil {
		return err
	}
	opts := make([]installer.Option, 0, 1)
	if in.KeepassXC.Path != "" {
		store := secret.NewKeepassXC(in.KeepassXC.Path, in.KeepassXC.Password)
		opts = append(opts, installer.WithStore("keepassxc", store))
	}
	ins := installer.New(rep, opts...)
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