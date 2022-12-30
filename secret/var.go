package secret

import (
	"io"
	"os/exec"
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
		defer in.Close()
		io.WriteString(in, k.password)
	}()

	value, err := cmd.Output()
	return string(value), err
}
