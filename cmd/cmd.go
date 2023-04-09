package cmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/livingsilver94/backee/logger"
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

type Arguments struct {
	Globals

	Install Install `cmd:"" default:"withargs"`
	// Copy is hidden and it is not meant to be called by users,
	// instead Backee will call it in a privileged fork of itself
	// to perform filesystem operations where admninistration rights are required.
	Copy copy `cmd:"" hidden:""`
}

func Run() {
	var args Arguments
	ctx := kong.Parse(&args)

	level := logger.LevelInfo
	if args.Globals.Quiet {
		level = logger.LevelError
	}
	logger := logger.New(level, !args.Globals.NoColor)

	ctx.Run(&logger)
}
