//go:build unix

package installer

import (
	"io/fs"

	"golang.org/x/sys/unix"
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
	stat := info.Sys().(unix.Stat_t)
	return UnixID{UID: stat.Uid, GID: stat.Gid}, nil
}

func RunAsUnixID(f func() error, id UnixID) error {
	oldUID := unix.Getuid()
	oldGID := unix.Getgid()
	err := unix.Setuid(int(id.UID))
	if err != nil {
		return err
	}
	defer unix.Setuid(oldUID)
	err = unix.Setgid(int(id.GID))
	if err != nil {
		return err
	}
	defer unix.Setgid(oldGID)
	return f()
}
