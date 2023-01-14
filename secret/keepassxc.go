package secret

import (
	"errors"
	"io"
	"os/exec"
	"strings"
)

type KeepassXC struct {
	dbPath   string
	password string
}

func NewKeepassXC(dbPath, password string) KeepassXC {
	return KeepassXC{
		dbPath:   dbPath,
		password: password,
	}
}

func (k KeepassXC) Value(key string) (string, error) {
	cmd := exec.Command("keepassxc-cli", "show", "-sa", "password", k.dbPath, key)

	in, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	go func() {
		io.WriteString(in, k.password)
	}()

	value, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			err = errors.New(strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return strings.TrimSuffix(string(value), "\n"), nil
}
