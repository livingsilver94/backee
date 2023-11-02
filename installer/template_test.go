package installer_test

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

func TestReplaceServiceVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	repl := installer.NewTemplate(serviceName, vars)

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
	repl := installer.NewTemplate(serviceName, vars)

	_, err := repl.ReplaceStringToString("{{thisKey}} is not among variables")
	if !errors.Is(err, repo.ErrNoVariable) {
		t.Fatalf("expected %v. Got %v", repo.ErrNoVariable, err)
	}
}

func TestReplaceExtraVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	vars.Common = map[string]string{"extra": "extraValue"}
	repl := installer.NewTemplate(serviceName, vars)

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
	vars.Insert("parent", "var1", service.VarValue{Kind: service.ClearText, Value: "parentValue1"})
	vars.AddParent(serviceName, "parent")
	repl := installer.NewTemplate(serviceName, vars)

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
		vars := repo.NewVariables()
		for key, val := range test.vars {
			vars.Insert("service", key, service.VarValue{Kind: service.ClearText, Value: val})
		}
		rep := installer.NewTemplate("service", vars)
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

func TestGobCodec(t *testing.T) {
	expected := installer.NewTemplate(serviceName, createVariables("var1", "value1"))

	data, err := expected.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	var result installer.Template
	err = result.GobDecode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected decoded template %#v. Got %#v", expected, result)
	}
}

const serviceName = "service1"

func createVariables(keyVal ...string) repo.Variables {
	if len(keyVal)%2 != 0 {
		panic("keys and values must be pairs")
	}
	v := repo.NewVariables()
	for i := 0; i < len(keyVal)-1; i += 2 {
		v.Insert(serviceName, keyVal[i], service.VarValue{Kind: service.ClearText, Value: keyVal[i+1]})
	}
	return v
}
