//go:build unix

package installer_test

import (
	"testing"
	"testing/fstest"

	"github.com/livingsilver94/backee/installer"
	"golang.org/x/sys/unix"
)

func TestUnixIDsFS(t *testing.T) {
	const expUID = 123
	const expGID = 456
	fs := fstest.MapFS{
		"file.txt": &fstest.MapFile{Sys: unix.Stat_t{Uid: expUID, Gid: expGID}},
	}

	uid, gid, err := installer.UnixIDsFS(fs, "file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if uid != expUID || gid != expGID {
		t.Fatalf("expected UID %d and GID %d. Got %d and %d", expUID, expGID, uid, gid)
	}
}

func TestRunAs(t *testing.T) {
	f := func() error { return nil }
	uid := unix.Getuid()
	gid := unix.Getgid()
	err := installer.RunAs(f, uid, gid)
	if err != nil {
		t.Fatal(err)
	}
}
