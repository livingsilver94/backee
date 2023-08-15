package privileged

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
)

var (
	possibleUtils = []string{"sudo", "doas"}
)

func Run(pipe *os.File) (*exec.Cmd, error) {
	path, err := os.Executable()
	if err != nil {
		return nil, err
	}
	for _, util := range possibleUtils {
		cmd := exec.Command(util, path, "privileged", strconv.FormatUint(uint64(pipe.Fd()), 10))
		err = cmd.Run()
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				continue
			}
			return nil, err
		}
		return cmd, nil
	}
	return nil, errors.New("no superuser utility found")
}
