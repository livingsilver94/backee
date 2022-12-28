package installer

import (
	"bufio"
	"fmt"
	"io"

	"github.com/hashicorp/go-set"
)

type installedList struct {
	dest      *bufio.Writer
	installed *set.Set[string]
}

func newInstalledList(dest io.ReadWriter) installedList {
	installed := set.New[string](10)
	scan := bufio.NewScanner(dest)
	for scan.Scan() {
		installed.Insert(scan.Text())
	}
	return installedList{
		dest:      bufio.NewWriter(dest),
		installed: installed,
	}
}

func (il *installedList) insert(name string) {
	fmt.Fprintln(il.dest, name)
	il.installed.Insert(name)
}

func (il *installedList) contains(name string) bool {
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
