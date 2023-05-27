package installer

import (
	"bufio"
	"fmt"
	"io"

	"github.com/hashicorp/go-set"
)

type InstallList struct {
	dest      io.Writer
	installed *set.Set[string]
}

func NewInstallList(dest io.ReadWriter) InstallList {
	if dest == nil {
		dest = newDiscard()
	}
	installed := set.New[string](10)
	scan := bufio.NewScanner(dest)
	for scan.Scan() {
		installed.Insert(scan.Text())
	}
	return InstallList{
		dest:      dest,
		installed: installed,
	}
}

func (il *InstallList) Insert(name string) {
	fmt.Fprintf(il.dest, "\n"+name)
	il.installed.Insert(name)
}

func (il *InstallList) Contains(name string) bool {
	return il.installed.Contains(name)
}

func (il *InstallList) Size() int {
	return il.installed.Size()
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
