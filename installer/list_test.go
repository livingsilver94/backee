// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package installer_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/livingsilver94/backee/installer"
)

func TestNewListCached(t *testing.T) {
	tests := []struct {
		in  io.ReadWriter
		out []string
	}{
		{in: bytes.NewBufferString(""), out: make([]string, 0)},
		{in: bytes.NewBufferString("service1\nservice2\nservice3"), out: []string{"service1", "service2", "service3"}},
	}
	for _, test := range tests {
		list, err := installer.NewListCached(test.in)
		if err != nil {
			t.Fatal(err)
		}
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

func TestListInsertCached(t *testing.T) {
	const serviceName = "testy"
	tests := []struct {
		in *bytes.Buffer
	}{
		{in: bytes.NewBufferString("service1\nservice2\nservice3")},
	}
	for _, test := range tests {
		list, err := installer.NewListCached(test.in)
		if err != nil {
			t.Fatal(err)
		}
		list.Insert(serviceName)
		if !list.Contains(serviceName) {
			t.Fatalf("list doesn't contain %q", serviceName)
		}
		if !strings.HasSuffix(test.in.String(), "\n"+serviceName) {
			t.Fatalf("list's ReadWriter doesn't end with %q", serviceName)
		}
	}
}
