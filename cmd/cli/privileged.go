package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/alecthomas/kong"
	priv "github.com/livingsilver94/backee/installer/privileged"
)

type privileged struct {
	// Pipe is the file interface of the pipe used to
	// receive instructions from the parent process.
	Pipe *os.File `arg:"" type:"fd" help:"file descriptor of the pipe from which to receive instructions."`
}

func (p privileged) Run() error {
	cmd, err := priv.ReceiveCommand(p.Pipe)
	if err != nil {
		return err
	}
	p.Pipe.Close()
	return cmd.Execute()
}

func fdMapper(ctx *kong.DecodeContext, target reflect.Value) error {
	if target.Type() != reflect.TypeOf((*os.File)(nil)) {
		return errors.New("the fd type must be applied to *os.File")
	}
	var fdString string
	err := ctx.Scan.PopValueInto("fd", &fdString)
	if err != nil {
		return err
	}
	fd, err := strconv.ParseUint(fdString, 10, strconv.IntSize)
	if err != nil {
		return err
	}
	file := os.NewFile(uintptr(fd), "pipe")
	if file == nil {
		return fmt.Errorf("%d is an invalid file descriptor", fd)
	}
	target.Set(reflect.ValueOf(file))
	return nil
}
