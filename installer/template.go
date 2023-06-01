package installer

import (
	"io"
	"strings"

	"github.com/livingsilver94/backee/service"
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

	srv  *service.Service
	vars Variables
}

func NewTemplate(srv *service.Service, vars Variables) Template {
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

func (t Template) varReplacerFunc() fasttemplate.TagFunc {
	return func(w io.Writer, tag string) (int, error) {
		if val, ok := t.vars.Get(t.srv.Name, tag); ok == nil {
			// Matched a variable local to the service.
			return w.Write([]byte(val))
		}
		if val, ok := t.ExtraVars[tag]; ok {
			// Matched an environment variable.
			return w.Write([]byte(val))
		}
		parentName, varName, found := strings.Cut(tag, ".")
		if !found {
			return 0, nil
		}
		for _, parent := range t.srv.Depends.Slice() {
			if parent != parentName {
				continue
			}
			if val, ok := t.vars.Get(parent, varName); ok == nil {
				// Matched a parent service variable.
				return w.Write([]byte(val))
			}
			break
		}
		return 0, nil
	}
}
