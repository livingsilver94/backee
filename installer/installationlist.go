package installer

import (
	"bufio"
	"fmt"
	"io"

	"github.com/hashicorp/go-set"
)

type InstallationList struct {
	dest      io.Writer
	installed *set.Set[string]
}

func NewInstallationList(dest io.ReadWriter) InstallationList {
	if dest == nil {
		dest = newDiscard()
	}
	installed := set.New[string](10)
	scan := bufio.NewScanner(dest)
	for scan.Scan() {
		installed.Insert(scan.Text())
	}
	return InstallationList{
		dest:      dest,
		installed: installed,
	}
}

func (il *InstallationList) Insert(name string) {
	fmt.Fprintf(il.dest, "\n"+name)
	il.installed.Insert(name)
}

func (il *InstallationList) Contains(name string) bool {
	return il.installed.Contains(name)
}

func (il *InstallationList) Size() int {
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
