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

func Run(run Runner) error {
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
		err = cmd.Run()
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				continue
			}
			return err
		}
		return SendRunner(pWrite, run)
	}
	return ErrNoElevUtil
}
