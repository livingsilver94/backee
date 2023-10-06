//go:build unix

package installer

import (
	"fmt"
	"io/fs"
	"syscall"
)

func runScript(script string) error {
	return runProcess(
		"sh",
		"-e", // Stop script on first error.
		"-c", // Run the following script string.
		script,
	)
}

func PathOwnerFS(sys fs.FS, path string) (UnixID, error) {
	info, err := fs.Stat(sys, path)
	if err != nil {
		return UnixID{}, err
	}
	stat := info.Sys().(*syscall.Stat_t)
	return UnixID{UID: stat.Uid, GID: stat.Gid}, nil
}

func RunAsUnixID(f func() error, id UnixID) error {
	oldUID := syscall.Geteuid()
	oldGID := syscall.Getegid()
	err := syscall.Setegid(int(id.GID))
	if err != nil {
		return fmt.Errorf("setting GID %d: %w", id.GID, err)
	}
	defer syscall.Setegid(oldGID)
	err = syscall.Seteuid(int(id.UID))
	if err != nil {
		return fmt.Errorf("setting UID %d: %w", id.UID, err)
	}
	defer syscall.Seteuid(oldUID)
	return f()
}
