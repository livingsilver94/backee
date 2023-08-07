package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/livingsilver94/backee/log"
	"golang.org/x/exp/slog"
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

	logOpt := log.Options{
		Level:   slog.LevelInfo,
		Colored: !args.Globals.NoColor,
	}
	if args.Globals.Quiet {
		logOpt.Level = slog.LevelError
	}
	slog.SetDefault(slog.New(log.NewHandler(os.Stdout, &logOpt)))

	err := ctx.Run()
	if err != nil {
		if unwrap := errors.Unwrap(err); unwrap != nil {
			// kong stupidly wraps the original error with the name
			// of the command that generated it. I want errors to be
			// readable to users, not developers.
			err = unwrap
		}
		slog.Error(err.Error())
	}
}
