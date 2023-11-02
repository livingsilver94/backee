package privilege

import (
	"encoding/gob"
	"io"
)

type Runner interface {
	Run() error
}

func RegisterInterfaceImpl(impl any) {
	gob.Register(impl)
}

func SendRunner(dst io.Writer, r Runner) error {
	return gob.NewEncoder(dst).Encode(&r)
}

func ReceiveRunner(src io.Reader) (Runner, error) {
	var run Runner
	err := gob.NewDecoder(src).Decode(&run)
	if err != nil {
		return nil, err
	}
	return run, nil
}
