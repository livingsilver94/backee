//go:build unix

package installer_test

import (
	"syscall"
	"testing"
	"testing/fstest"

	"github.com/livingsilver94/backee/installer"
)

func TestUnixIDsFS(t *testing.T) {
	const expUID = 123
	const expGID = 456
	fs := fstest.MapFS{
		"file.txt": &fstest.MapFile{Sys: &syscall.Stat_t{Uid: expUID, Gid: expGID}},
	}

	id, err := installer.PathOwnerFS(fs, "file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if id.UID != expUID || id.GID != expGID {
		t.Fatalf("expected UID %d and GID %d. Got %d and %d", expUID, expGID, id.UID, id.GID)
	}
}

func TestRunAs(t *testing.T) {
	f := func() error { return nil }
	uid := syscall.Getuid()
	gid := syscall.Getgid()
	err := installer.RunAsUnixID(f, installer.UnixID{UID: uint32(uid), GID: uint32(gid)})
	if err != nil {
		t.Fatal(err)
	}
}
