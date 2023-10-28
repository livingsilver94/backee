package cli

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

// Version will be injected by ld flags.
var Version string

type flagVersion bool

func (v *flagVersion) BeforeReset() error {
	fmt.Println(Version)
	os.Exit(0)
	return nil
}

type Globals struct {
	NoColor bool        `help:"Do not color output (the default when in a non-interactive shell)."`
	Quiet   bool        `short:"q" help:"Do not print anything on the terminal except errors."`
	Version flagVersion `short:"v" help:"Print the version number and exit."`
}

type arguments struct {
	Globals

	Install install `cmd:"" default:"withargs"`
	// Privilege is  a hidden subcommand, not meant to be called by users.
	// Instead, Backee will call it in a privileged fork of itself
	// to perform filesystem operations where administration rights are required.
	Privilege privilege `cmd:"" hidden:""`
}

func Parse() (*kong.Context, Globals) {
	var args arguments
	ctx := kong.Parse(&args)
	return ctx, args.Globals
}
