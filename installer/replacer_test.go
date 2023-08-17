package installer_test

import (
	"bytes"
	"errors"
	"strings"
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
	vars.Common = map[string]string{"extra": "extraValue"}
	repl := installer.NewReplacer(serviceName, vars)

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

func TestReplaceReader(t *testing.T) {
	tests := []struct {
		in   string
		vars map[string]string
		out  string
	}{
		{"aaaaa", map[string]string{"var1": "value1"}, "aaaaa"},
		{"aaaaa {{var1}}", map[string]string{"var1": "value1"}, "aaaaa value1"},
	}
	for _, test := range tests {
		s := strings.NewReader(test.in)
		vars := installer.NewVariables()
		for key, val := range test.vars {
			vars.Insert("service", key, backee.VarValue{Kind: backee.ClearText, Value: val})
		}
		rep := installer.NewReplacer("service", vars)
		writer := &bytes.Buffer{}
		err := rep.Replace(s, writer)
		if err != nil {
			t.Fatal(err)
		}
		if writer.String() != test.out {
			t.Fatalf("expected string %q. Got %q", test.out, writer.String())
		}
	}
}
