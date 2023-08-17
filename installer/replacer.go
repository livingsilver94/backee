package installer

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/livingsilver94/backee"
	"github.com/valyala/fasttemplate"
)

type Replacer struct {
	ExtraVars map[string]string

	serviceName string
	variables   Variables
}

func NewReplacer(serviceName string, vars Variables) Replacer {
	return Replacer{
		serviceName: serviceName,
		variables:   vars,
	}
}

func (t Replacer) Replace(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(greedyTagSplitter)
	for scanner.Scan() {
		_, err := fasttemplate.ExecuteFunc(
			scanner.Text(),
			backee.VarOpenTag, backee.VarCloseTag,
			w,
			t.replaceTag,
		)
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (t Replacer) ReplaceString(s string, w io.Writer) error {
	_, err := fasttemplate.ExecuteFunc(
		s,
		backee.VarOpenTag, backee.VarCloseTag,
		w,
		t.replaceTag,
	)
	return err
}

func (t Replacer) ReplaceStringToString(s string) (string, error) {
	return fasttemplate.ExecuteFuncStringWithErr(
		s,
		backee.VarOpenTag, backee.VarCloseTag,
		t.replaceTag,
	)
}

func (t Replacer) replaceTag(w io.Writer, varName string) (int, error) {
	if val, err := t.variables.Get(t.serviceName, varName); err == nil {
		// Matched a variable local to the backee.
		return w.Write([]byte(val))
	}
	if val, ok := t.ExtraVars[varName]; ok {
		// Matched an extra variable.
		return w.Write([]byte(val))
	}
	parentName, parentVar, found := strings.Cut(varName, backee.VarParentSep)
	if !found {
		return 0, ErrNoVariable
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
	return 0, ErrNoVariable
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

	iTag := bytes.LastIndex(data, []byte(backee.VarOpenTag))
	if iTag < 0 {
		// No tags in this string.
		return len(data), data, nil
	}
	if bytes.Contains(data[iTag+len(backee.VarOpenTag):], []byte(backee.VarCloseTag)) {
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
