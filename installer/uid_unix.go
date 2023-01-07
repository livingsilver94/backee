//go:build unix

package installer

import (
	"io/fs"
	"os"

	"golang.org/x/sys/unix"
)

func UnixIDsFS(sys fs.FS, path string) (uid int, gid int, err error) {
	info, err := fs.Stat(sys, path)
	if err != nil {
		return -1, -1, err
	}
	stat := info.Sys().(unix.Stat_t)
	return int(stat.Uid), int(stat.Gid), nil
}

func UnixIDs(path string) (uid int, gid int, err error) {
	return UnixIDsFS(os.DirFS(path), ".")
}

func RunAs(f func() error, uid, gid int) error {
	oldUID := unix.Getuid()
	oldGID := unix.Getgid()
	err := unix.Setuid(uid)
	if err != nil {
		return err
	}
	defer unix.Setuid(oldUID)
	err = unix.Setgid(gid)
	if err != nil {
		return err
	}
	defer unix.Setgid(oldGID)
	return f()
}
