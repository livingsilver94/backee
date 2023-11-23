//go:build unix

// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package stepwriter_test

import (
	"syscall"
	"testing"
	"testing/fstest"

	"github.com/livingsilver94/backee/installer/stepwriter"
)

func TestUnixIDsFS(t *testing.T) {
	const expUID = 123
	const expGID = 456
	fs := fstest.MapFS{
		"file.txt": &fstest.MapFile{Sys: &syscall.Stat_t{Uid: expUID, Gid: expGID}},
	}

	id, err := stepwriter.PathOwnerFS(fs, "file.txt")
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
	err := stepwriter.RunAsUnixID(f, stepwriter.UnixID{UID: uint32(uid), GID: uint32(gid)})
	if err != nil {
		t.Fatal(err)
	}
}
