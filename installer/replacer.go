package installer

import (
	"io"
	"strings"

	"github.com/livingsilver94/backee/service"
	"github.com/valyala/fasttemplate"
)

var (
	repl fasttemplate.Template
)

type Replacer struct {
	ExtraVars map[string]string

	srvName   string
	variables Variables
}

func NewReplacer(srvName string, vars Variables) Replacer {
	return Replacer{
		srvName:   srvName,
		variables: vars,
	}
}

func (t Replacer) Replace(s string, w io.Writer) error {
	err := repl.Reset(s, service.VarOpenTag, service.VarCloseTag)
	if err != nil {
		return err
	}
	_, err = repl.ExecuteFunc(w, t.replaceTag)
	return err
}

func (t Replacer) ReplaceToString(s string) (string, error) {
	builder := strings.Builder{}
	err := t.Replace(s, &builder)
	return builder.String(), err
}

func (t Replacer) replaceTag(w io.Writer, varName string) (int, error) {
	if val, err := t.variables.Get(t.srvName, varName); err == nil {
		// Matched a variable local to the service.
		return w.Write([]byte(val))
	}
	if val, ok := t.ExtraVars[varName]; ok {
		// Matched an extra variable.
		return w.Write([]byte(val))
	}
	parentName, parentVar, found := strings.Cut(varName, service.VarParentSep)
	if !found {
		return 0, ErrNoVariable
	}
	parents, _ := t.variables.Parents(t.srvName)
	for _, parent := range parents {
		if parent != parentName {
			continue
		}
		if val, ok := t.variables.Get(parent, parentVar); ok == nil {
			// Matched a parent service variable.
			return w.Write([]byte(val))
		}
		break
	}
	return 0, ErrNoVariable
}
