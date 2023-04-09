package cmd

import (
	"strconv"

	"github.com/alecthomas/kong"
)

type octalInt32 int32

func (o *octalInt32) UnmarshalText(text []byte) error {
	i, err := strconv.ParseInt(string(text), 8, 32)
	if err != nil {
		return err
	}
	*o = octalInt32(i)
	return nil
}

type copy struct {
	Source      string `arg:""`
	Destination string `arg:""`
	Symlink     bool
	Permission  octalInt32 `default:"644"`
}

func (c *copy) Run(ctx *kong.Context) error {
	return nil
}
