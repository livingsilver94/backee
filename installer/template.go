package installer

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
	"github.com/valyala/fasttemplate"
)

type Template struct {
	serviceName string
	variables   repo.Variables
}

func NewTemplate(serviceName string, vars repo.Variables) Template {
	return Template{
		serviceName: serviceName,
		variables:   vars,
	}
}

func (t Template) Replace(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(greedyTagSplitter)
	for scanner.Scan() {
		_, err := fasttemplate.ExecuteFunc(
			scanner.Text(),
			service.VarOpenTag, service.VarCloseTag,
			w,
			t.replaceTag,
		)
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (t Template) ReplaceString(s string, w io.Writer) error {
	_, err := fasttemplate.ExecuteFunc(
		s,
		service.VarOpenTag, service.VarCloseTag,
		w,
		t.replaceTag,
	)
	return err
}

func (t Template) ReplaceStringToString(s string) (string, error) {
	return fasttemplate.ExecuteFuncStringWithErr(
		s,
		service.VarOpenTag, service.VarCloseTag,
		t.replaceTag,
	)
}

func (t Template) replaceTag(w io.Writer, varName string) (int, error) {
	if val, err := t.variables.Get(t.serviceName, varName); err == nil {
		// Matched a variable local to the service.
		return w.Write([]byte(val))
	}

	parentName, parentVar, found := strings.Cut(varName, service.VarParentSep)
	if !found {
		return 0, repo.ErrNoVariable
	}
	parents, _ := t.variables.Parents(t.serviceName)
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
	return 0, repo.ErrNoVariable
}

// greedyTagSplitter is a bufio.SplitFunc that reads
// the longest string possible without unclosed tags.
func greedyTagSplitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if !atEOF && len(data) < bufio.MaxScanTokenSize {
		// Read more: we're greedy!
		return 0, nil, nil
	}
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	iTag := bytes.LastIndex(data, []byte(service.VarOpenTag))
	if iTag < 0 {
		// No tags in this string.
		return len(data), data, nil
	}
	if bytes.Contains(data[iTag+len(service.VarOpenTag):], []byte(service.VarCloseTag)) {
		// Tag is correctly closed. Return whole string.
		return len(data), data, nil
	}
	if !atEOF {
		// We don't want to return the full string if there's an unclosed tag.
		// Return string right before the tag opening.
		return len(data[:iTag]), data[:iTag], nil
	}
	// Tag is not closed but there's nothing we can do.
	return len(data), data, nil
}
