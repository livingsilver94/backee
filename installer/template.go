package installer

import (
	"io"
	"strings"

	"github.com/valyala/fasttemplate"
)

const (
	templateOpenTag  = "{{"
	templateCloseTag = "}}"
)

var (
	tmpl fasttemplate.Template
)

type Template struct {
	ExtraVars map[string]string

	srv  string
	vars Variables
}

func NewTemplate(srv string, vars Variables) Template {
	return Template{
		srv:  srv,
		vars: vars,
	}
}

func (t Template) Execute(s string, w io.Writer) error {
	err := tmpl.Reset(s, templateOpenTag, templateCloseTag)
	if err != nil {
		return err
	}
	_, err = tmpl.ExecuteFunc(w, t.replaceTag)
	return err
}

func (t Template) ExecuteString(s string) (string, error) {
	builder := strings.Builder{}
	err := t.Execute(s, &builder)
	return builder.String(), err
}

func (t Template) replaceTag(w io.Writer, tag string) (int, error) {
	if val, err := t.vars.Get(t.srv, tag); err == nil {
		// Matched a variable local to the service.
		return w.Write([]byte(val))
	}
	if val, ok := t.ExtraVars[tag]; ok {
		// Matched an extra variable.
		return w.Write([]byte(val))
	}
	parentName, varName, found := strings.Cut(tag, ".")
	if !found {
		return 0, ErrNoVariable
	}
	parents, _ := t.vars.Parents(t.srv)
	for _, parent := range parents {
		if parent != parentName {
			continue
		}
		if val, ok := t.vars.Get(parent, varName); ok == nil {
			// Matched a parent service variable.
			return w.Write([]byte(val))
		}
		break
	}
	return 0, ErrNoVariable
}
