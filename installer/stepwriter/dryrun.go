// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package stepwriter

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/service"
)

type DryRun struct {
	// Dest is where messages are written to.
	// When nil, it defaults to [os.Stdout].
	Dest io.Writer
}

func (d DryRun) Setup(script string) error {
	_, err := d.print(script)
	return err
}

func (d DryRun) InstallPackages(fullCmd []string) error {
	_, err := d.printf("Will run %q", strings.Join(fullCmd, " "))
	return err
}

func (d DryRun) SymlinkFile(dst service.FilePath, src string) error {
	_, err := d.printf("%s\t➜ %s", src, dst.Path)
	if err != nil {
		return err
	}
	if dst.Mode != 0 {
		_, err = d.printf(" with permission %o", dst.Mode)
		if err != nil {
			return err
		}
	}
	_, err = d.println()
	return err
}

func (d DryRun) CopyFile(dst service.FilePath, src installer.FileCopy) error {
	_, err := d.printf("Will write %q", dst.Path)
	if err != nil {
		return err
	}
	if dst.Mode != 0 {
		_, err = d.printf(" with permission %o", dst.Mode)
		if err != nil {
			return err
		}
	}
	_, err = d.println(" with the following content:")
	if err != nil {
		return err
	}
	_, err = d.println(src)
	return err
}

func (d DryRun) Finalize(script string) error {
	_, err := fmt.Print(script)
	return err
}

func (d DryRun) print(a ...any) (n int, err error) {
	dest := d.Dest
	if dest == nil {
		dest = os.Stdout
	}
	return fmt.Fprint(d.Dest, a...)
}

func (d DryRun) printf(format string, a ...any) (n int, err error) {
	dest := d.Dest
	if dest == nil {
		dest = os.Stdout
	}
	return fmt.Fprintf(d.Dest, format, a...)
}

func (d DryRun) println(a ...any) (n int, err error) {
	dest := d.Dest
	if dest == nil {
		dest = os.Stdout
	}
	return fmt.Fprintln(d.Dest, a...)
}
