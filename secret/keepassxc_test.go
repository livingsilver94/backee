package secret_test

import (
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/livingsilver94/backee/secret"
)

func TestKeepassXC(t *testing.T) {
	k := secret.NewKeepassXC(filepath.Join("testdata", "keepassxc.kdbx"), "password")
	val, err := k.Value("test")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			t.Skip("KeepassXC not installed")
		}
		t.Fatal(err)
	}
	if val != "testvalue" {
		t.Fatalf("expected value %q. Got %q", "testvalue", val)
	}
}
