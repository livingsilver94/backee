// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package cli

import (
	"os"

	_ "github.com/livingsilver94/backee/installer"
	priv "github.com/livingsilver94/backee/privilege"
)

type privilege struct{}

func (p privilege) Run() error {
	run, err := priv.ReceiveRunner(os.Stdin)
	if err != nil {
		return err
	}
	return run.RunPrivileged()
}
