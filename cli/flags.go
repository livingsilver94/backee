package cli

import (
	"github.com/alecthomas/kong"
)

type KeepassXC struct {
	Path     string `env:"KEEPASSXC_PATH" help:"KeepassXC database path."`
	Password string `env:"KEEPASSXC_PASSWORD" help:"KeepassXC database password."`
}

type Arguments struct {
	Directory string    `short:"C" type:"existingdir" help:"Change the base directory."`
	KeepassXC KeepassXC `embed:"" prefix:"keepassxc."`
	NoColor   bool      `help:"Do not color output (the default when in a non-interactive shell)."`
	Quiet     bool      `short:"q" help:"Do not print anything on the terminal."`
	Variant   string    `help:"Specify the system variant."`

	Services []string `arg:"" help:"Services to install."`
}

func ParseArguments() Arguments {
	var args Arguments
	kong.Parse(&args)
	return args
}
