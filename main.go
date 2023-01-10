package main

import (
	"log"

	"github.com/livingsilver94/backee/cli"
	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/repo"
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

	var ins installer.Installer
	if args.Quiet {
		ins = installer.New(rep)
	} else {
		ins = installer.New(rep, installer.WithLogger(cli.Logger()))
	}
	return ins.Install(srv)
}

func main() {
	err := run()
	if err != nil {
		log.Println(err)
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
