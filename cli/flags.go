package cli

import (
	"os"

	"github.com/alecthomas/kong"
)

type Arguments struct {
	Directory string `short:"C" type:"existingdir" help:"Change the base directory."`
	Quiet     bool   `short:"q" help:"Do not print anything on the terminal."`
	Variant   string `help:"Specify the system variant."`

	Services []string `arg:"" help:"Services to install."`
}

func ParseArguments() (Arguments, error) {
	var args Arguments
	kong.Parse(&args)
	if args.Directory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return args, err
		}
		args.Directory = cwd
	}
	return args, nil
}
