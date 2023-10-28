package cli

import (
	"os"

	"github.com/livingsilver94/backee/installer"
	priv "github.com/livingsilver94/backee/privilege"
)

type privilege struct{}

func (p privilege) Run() error {
	installer.RegisterPrivilegedTypes()
	run, err := priv.ReceiveRunner(os.Stdin)
	if err != nil {
		return err
	}
	return run.Run()
}
