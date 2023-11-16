//go:build windows

// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package installer

import (
	"io/fs"
)

func runScript(script string) error {
	return runProcess(
		"powershell",
		"-NoLogo",  // Hide copyright banner.
		"-Command", // Run the following script string.
		script,
	)
}

func PathOwnerFS(sys fs.FS, path string) (UnixID, error) {
	return UnixID{}, nil
}

func RunAsUnixID(f func() error, id UnixID) error {
	return f()
}
