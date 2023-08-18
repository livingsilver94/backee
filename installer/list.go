package installer

import (
	"bufio"
	"fmt"
	"io"

	"github.com/hashicorp/go-set"
)

type List struct {
	installed *set.Set[string]
	cache     io.Writer
}

func NewList() List {
	return List{
		installed: set.New[string](10),
	}
}

func NewListCached(cache io.ReadWriter) (List, error) {
	list := NewList()
	list.cache = cache

	scan := bufio.NewScanner(cache)
	for scan.Scan() {
		list.installed.Insert(scan.Text())
	}
	return list, scan.Err()
}

func (il *List) Insert(name string) error {
	var err error
	if il.cache != nil {
		_, err = fmt.Fprintf(il.cache, "\n"+name)
	}
	il.installed.Insert(name)
	return err
}

func (il *List) Contains(name string) bool {
	return il.installed.Contains(name)
}

func (il *List) Size() int {
	return il.installed.Size()
}
