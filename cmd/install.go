package cmd

import (
	"errors"
	"os"

	"github.com/go-logr/logr"
	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/secret"
	"github.com/livingsilver94/backee/service"
)

type KeepassXC struct {
	Path     string `env:"KEEPASSXC_PATH" help:"KeepassXC database path."`
	Password string `env:"KEEPASSXC_PASSWORD" help:"KeepassXC database password."`
}

type Install struct {
	Directory string    `short:"C" type:"existingdir" help:"Change the base directory."`
	KeepassXC KeepassXC `embed:"" prefix:"keepassxc."`
	Variant   string    `help:"Specify the system variant."`

	Services []string `arg:"" optional:"" help:"Services to install. Pass none to install all services in the base directory."`
}

func (in *Install) Run(logger *logr.Logger) error {
	if in.Directory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		in.Directory = cwd
	}

	rep := repo.NewFSRepoVariant(repo.NewOSFS(in.Directory), in.Variant)
	srv, err := in.services(rep, in.Services)
	if err != nil {
		return err
	}
	opts := make([]installer.Option, 0, 2)
	if in.KeepassXC.Path != "" {
		store := secret.NewKeepassXC(in.KeepassXC.Path, in.KeepassXC.Password)
		opts = append(opts, installer.WithStore("keepassxc", store))
	}
	opts = append(opts, installer.WithLogger(*logger))
	ins := installer.New(rep, opts...)
	if !ins.Install(srv) {
		// The installer already logged the error.
		return errors.New("")
	}
	return nil
}

func (Install) services(rep repo.FSRepo, names []string) ([]*service.Service, error) {
	if len(names) == 0 {
		return rep.AllServices()
	}
	services := make([]*service.Service, 0, len(names))
	for _, name := range names {
		srv, err := rep.Service(name)
		if err != nil {
			return nil, err
		}
		services = append(services, srv)
	}
	return services, nil
}
