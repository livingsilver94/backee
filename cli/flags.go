package cli

import (
	"github.com/alecthomas/kong"
)

type Arguments struct {
	Quiet   bool   `short:"q" help:"Do not print anything on the terminal."`
	Variant string `help:"Specify the system variant."`

	Services []string `arg:"" type:"existingdir" help:"Services to install."`
}

func ParseArguments() Arguments {
	var args Arguments
	kong.Parse(&args)
	return args
}
