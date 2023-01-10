package installer_test

import (
	"os"
	"testing"

	"github.com/livingsilver94/backee/installer"
)

func TestReplaceEnvVars(t *testing.T) {
	os.Setenv("MYVAR1", "vAlUe")
	os.Setenv("_mYOtherVar2", "11vAlUe")
	defer func() {
		os.Unsetenv("MYVAR1")
		os.Unsetenv("_mYOtherVar2")
	}()

	expected := "this is vAlUe and this one is 11vAlUe"
	obtained := installer.ReplaceEnvVars("this is ${MYVAR1} and this one is ${_mYOtherVar2}")
	if obtained != expected {
		t.Fatalf("expected %q. Got %q", expected, obtained)
	}
}
