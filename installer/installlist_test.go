package installer_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/livingsilver94/backee/installer"
)

func TestNewList(t *testing.T) {
	emptyRead := bytes.NewBufferString("")
	someRead := bytes.NewBufferString("service1\nservice2\nservice3")

	tests := []struct {
		in  io.ReadWriter
		out []string
	}{
		{in: nil, out: make([]string, 0)},
		{in: emptyRead, out: make([]string, 0)},
		{in: someRead, out: []string{"service1", "service2", "service3"}},
	}

	for _, test := range tests {
		list := installer.NewInstallList(test.in)
		if list.Size() != len(test.out) {
			t.Fatalf("Expected list length %d. Got %d", len(test.out), list.Size())
		}
		for _, s := range test.out {
			if !list.Contains(s) {
				t.Fatalf("list should contain %s but it doesn't", s)
			}
		}
	}
}

func TestListInsert(t *testing.T) {
	someRead := bytes.NewBufferString("service1\nservice2\nservice3")
	list := installer.NewInstallList(someRead)
	list.Insert("testy")
	if !list.Contains("testy") {
		t.Fatalf("list doesn't contain %q", "testy")
	}
	if !strings.HasSuffix(someRead.String(), "\ntesty") {
		t.Fatalf("list's ReadWriter doesn't end with %q", "testy")
	}
}
