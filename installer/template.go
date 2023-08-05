package installer

import (
	"errors"
	"io"
	"strings"

	"github.com/valyala/fasttemplate"
)

const (
	templateOpenTag  = "{{"
	templateCloseTag = "}}"
)

var (
	ErrNoTag = errors.New("template tag not found")
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
	_, err = tmpl.ExecuteFunc(w, t.varReplacerFunc())
	return err
}

func (t Template) ExecuteString(s string) (string, error) {
	b := strings.Builder{}
	err := t.Execute(s, &b)
	return b.String(), err
}

func (t Template) varReplacerFunc() fasttemplate.TagFunc {
	return func(w io.Writer, tag string) (int, error) {
		if val, err := t.vars.Get(t.srv, tag); err == nil {
			// Matched a variable local to the service.
			return w.Write([]byte(val))
		}
		if val, ok := t.ExtraVars[tag]; ok {
			// Matched an environment variable.
			return w.Write([]byte(val))
		}
		parentName, varName, found := strings.Cut(tag, ".")
		if !found {
			return 0, ErrNoTag
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
		return 0, ErrNoTag
	}
}
