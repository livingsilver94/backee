package main

import (
	"github.com/livingsilver94/backee/cli"
	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/secret"
	"github.com/livingsilver94/backee/service"
)

func run() error {
	args, err := cli.ParseArguments()
	if err != nil {
		return err
	}

	rep := repo.NewFSRepoVariant(repo.NewOSFS(args.Directory), args.Variant)
	srv, err := services(rep, args.Services)
	if err != nil {
		return err
	}

	opts := make([]installer.Option, 0, 2)
	if !args.Quiet {
		opts = append(opts, installer.WithLogger(cli.Logger))
	}
	if args.KeepassXC.Path != "" {
		store := secret.NewKeepassXC(args.KeepassXC.Path, args.KeepassXC.Password)
		opts = append(opts, installer.WithStore("keepassxc", store))
	}
	ins := installer.New(rep, opts...)
	return ins.Install(srv)
}

func main() {
	err := run()
	if err != nil {
		cli.Logger.Error(err, "")
	}
}

func services(rep repo.FSRepo, names []string) ([]*service.Service, error) {
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
