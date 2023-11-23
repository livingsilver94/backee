// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package stepwriter

import (
	"fmt"
	"strings"

	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/service"
)

type DryRun struct{}

func (DryRun) Setup(script string) error {
	_, err := fmt.Print(script)
	return err
}

func (DryRun) InstallPackages(fullCmd []string) error {
	_, err := fmt.Printf("Will run %q", strings.Join(fullCmd, " "))
	return err
}

func (DryRun) SymlinkFile(dst service.FilePath, src string) error {
	_, err := fmt.Printf("%s\t➜ %s", src, dst.Path)
	if err != nil {
		return err
	}
	if dst.Mode != 0 {
		_, err = fmt.Printf(" with permission %o", dst.Mode)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Println()
	return err
}

func (DryRun) CopyFile(dst service.FilePath, src installer.FileCopy) error {
	_, err := fmt.Printf("Will write %q", dst.Path)
	if err != nil {
		return err
	}
	if dst.Mode != 0 {
		_, err = fmt.Printf(" with permission %o", dst.Mode)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Println(" with the following content:")
	if err != nil {
		return err
	}
	_, err = fmt.Println(src)
	return err
}

func (DryRun) Finalize(script string) error {
	_, err := fmt.Print(script)
	return err
}
