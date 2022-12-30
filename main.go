package main

import (
	"fmt"
	"log"
	"os"

	"github.com/livingsilver94/backee/cli"
	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/repo"
)

func run() error {
	flags := cli.ParseFlags()

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	rep := repo.NewFSRepoVariant(cwd, flags.Variant)
	ins := installer.New(rep)
	fmt.Println(ins)
	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Println(err)
	}
}
