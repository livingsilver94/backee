package installer_test

import (
	"errors"
	"testing"

	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/service"
)

func TestTemplateServiceVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	tmpl := installer.NewTemplate(serviceName, vars)

	s := "this test prints {{var2}}"
	const expected = "this test prints value2"
	obtained, err := tmpl.ExecuteString(s)
	if err != nil {
		t.Fatalf("expected nil error. Got %v", err)
	}
	if obtained != expected {
		t.Fatalf("expected string %q. Got %q", expected, obtained)
	}
}

func TestTemplateNoVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	tmpl := installer.NewTemplate(serviceName, vars)

	_, err := tmpl.ExecuteString("{{thisKey}} is not among variables")
	if !errors.Is(err, installer.ErrNoVariable) {
		t.Fatalf("expected %v. Got %v", installer.ErrNoVariable, err)
	}
}

func TestTemplateExtraVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	tmpl := installer.NewTemplate(serviceName, vars)
	tmpl.ExtraVars = map[string]string{"extra": "extraValue"}

	s := "this test prints {{extra}}"
	const expected = "this test prints extraValue"
	obtained, err := tmpl.ExecuteString(s)
	if err != nil {
		t.Fatalf("expected nil error. Got %v", err)
	}
	if obtained != expected {
		t.Fatalf("expected string %q. Got %q", expected, obtained)
	}
}

func TestTemplateParentVar(t *testing.T) {
	vars := createVariables("var1", "value1", "var2", "value2")
	vars.Insert("parent", "var1", service.VarValue{Kind: service.ClearText, Value: "parentValue1"})
	vars.AddParent(serviceName, "parent")
	tmpl := installer.NewTemplate(serviceName, vars)

	s := "this test prints {{parent.var1}}"
	const expected = "this test prints parentValue1"
	obtained, err := tmpl.ExecuteString(s)
	if err != nil {
		t.Fatalf("expected nil error. Got %v", err)
	}
	if obtained != expected {
		t.Fatalf("expected string %q. Got %q", expected, obtained)
	}
}
