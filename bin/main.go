package main

import (
	"errors"
	"os"

	"github.com/livingsilver94/backee/bin/cli"
	"golang.org/x/exp/slog"
)

func main() {
	ctx, globals := cli.Parse()
	logOpt := LogHandlerOptions{
		Level:   slog.LevelInfo,
		Colored: !globals.NoColor,
	}
	if globals.Quiet {
		logOpt.Level = slog.LevelError
	}
	slog.SetDefault(slog.New(NewLogHandler(os.Stdout, &logOpt)))

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
