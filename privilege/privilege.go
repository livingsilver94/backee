package privilege

import (
	"errors"
	"os"
	"os/exec"
)

const (
	CLICommand = "privilege"
)

var (
	ErrNoElevUtil = errors.New("no privilege elevation utility found")
)

var (
	elevationUtils = []string{"sudo", "doas"}
)

func Run(run Runner) (err error) {
	path, err := os.Executable()
	if err != nil {
		return err
	}
	pRead, pWrite, err := os.Pipe()
	if err != nil {
		return err
	}
	for _, util := range elevationUtils {
		cmd := exec.Command(util, path, CLICommand)
		cmd.Stdin = pRead
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				continue
			}
			return err
		}
		defer func() {
			err = anyOf(err, pWrite.Close())
			if err != nil {
				cmd.Process.Kill()
			} else {
				err = cmd.Wait()
			}
		}()
		return SendRunner(pWrite, run)
	}
	return ErrNoElevUtil
}

func anyOf(err1, err2 error) error {
	if err1 != nil {
		return err1
	}
	return err2
}
