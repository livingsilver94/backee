package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/livingsilver94/backee/cli"
	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

func run() error {
	args := cli.ParseArguments()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	rep := repo.NewFSRepoVariant(cwd, args.Variant)
	srv, err := services(rep, args.Services)
	if err != nil {
		return err
	}

	var ins installer.Installer
	if args.Quiet {
		ins = installer.New(rep)
	} else {
		ins = installer.New(rep, installer.WithLogger(logger()))
	}

	return ins.Install(srv)
}

func main() {
	err := run()
	if err != nil {
		log.Println(err)
	}
}

func logger() logr.Logger {
	f := func(prefix, args string) {
		fmt.Printf(
			"[%s] %s — %s",
			time.Now().Format("15:04:05"), strings.ToUpper(prefix), args,
		)
	}
	return funcr.New(f, funcr.Options{})
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
