package installer

import (
	"bufio"
	"fmt"
	"io"

	"github.com/hashicorp/go-set"
)

type List struct {
	dest      *bufio.Writer
	installed *set.Set[string]
}

func NewList(dest io.ReadWriter) List {
	if dest == nil {
		dest = newDiscard()
	}
	installed := set.New[string](10)
	scan := bufio.NewScanner(dest)
	for scan.Scan() {
		installed.Insert(scan.Text())
	}
	return List{
		dest:      bufio.NewWriter(dest),
		installed: installed,
	}
}

func (il *List) Insert(name string) {
	fmt.Fprintln(il.dest, name)
	il.installed.Insert(name)
}

func (il *List) Contains(name string) bool {
	return il.installed.Contains(name)
}

type discard struct {
	io.Writer
}

func newDiscard() discard {
	return discard{io.Discard}
}

func (d discard) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}
