package secret

import (
	"bytes"
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
	cmd := exec.Command("keepassxc-cli", "show", "-q", "-sa", "password", k.dbPath, key)

	in, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	go func() {
		defer in.Close()
		io.WriteString(in, k.password)
	}()

	value, err := cmd.Output()
	return string(bytes.TrimSuffix(value, []byte{'\n'})), err
}
