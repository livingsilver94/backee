// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"log/slog"
	"os"

	"github.com/livingsilver94/backee/cli"
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
		slog.Error(err.Error())
		os.Exit(1)
	}
}
