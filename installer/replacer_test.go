package installer_test

import (
	"errors"
	"testing"

	"github.com/livingsilver94/backee"
	"github.com/livingsilver94/backee/installer"
)

func TestReplaceServiceVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	repl := installer.NewReplacer(serviceName, vars)

	s := "this test prints {{var2}}"
	const expected = "this test prints value2"
	obtained, err := repl.ReplaceStringToString(s)
	if err != nil {
		t.Fatalf("expected nil error. Got %v", err)
	}
	if obtained != expected {
		t.Fatalf("expected string %q. Got %q", expected, obtained)
	}
}

func TestReplaceNoVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	repl := installer.NewReplacer(serviceName, vars)

	_, err := repl.ReplaceStringToString("{{thisKey}} is not among variables")
	if !errors.Is(err, installer.ErrNoVariable) {
		t.Fatalf("expected %v. Got %v", installer.ErrNoVariable, err)
	}
}

func TestReplaceExtraVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	repl := installer.NewReplacer(serviceName, vars)
	repl.ExtraVars = map[string]string{"extra": "extraValue"}

	s := "this test prints {{extra}}"
	const expected = "this test prints extraValue"
	obtained, err := repl.ReplaceStringToString(s)
	if err != nil {
		t.Fatalf("expected nil error. Got %v", err)
	}
	if obtained != expected {
		t.Fatalf("expected string %q. Got %q", expected, obtained)
	}
}

func TestReplaceParentVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	vars.Insert("parent", "var1", backee.VarValue{Kind: backee.ClearText, Value: "parentValue1"})
	vars.AddParent(serviceName, "parent")
	repl := installer.NewReplacer(serviceName, vars)

	s := "this test prints {{parent.var1}}"
	const expected = "this test prints parentValue1"
	obtained, err := repl.ReplaceStringToString(s)
	if err != nil {
		t.Fatalf("expected nil error. Got %v", err)
	}
	if obtained != expected {
		t.Fatalf("expected string %q. Got %q", expected, obtained)
	}
}
