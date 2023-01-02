package repo

import (
	"fmt"

	"github.com/hashicorp/go-set"
	"github.com/livingsilver94/backee/service"
)

type DepGraph struct {
	graph []DepSet
}

func NewDepGraph(capacity int) DepGraph {
	return DepGraph{
		graph: make([]DepSet, 0, capacity),
	}
}

const depSetDefaultCap = 10

func (dg *DepGraph) Insert(level int, srv *service.Service) {
	if level > dg.Depth() {
		panic(fmt.Sprintf("inserting dep on level %d but level %d does not exist", level, level-1))
	}
	if level == dg.Depth() {
		dg.graph = append(dg.graph, NewDepSet(depSetDefaultCap))
	}
	dg.Level(level).Insert(srv)
}

func (dg *DepGraph) Depth() int {
	return len(dg.graph)
}

func (dg *DepGraph) Level(index int) *DepSet {
	return &dg.graph[index]
}

func (dg *DepGraph) Equal(dg2 DepGraph) bool {
	if dg.Depth() != dg2.Depth() {
		return false
	}
	for level := 0; level < dg.Depth(); level++ {
		if !dg.Level(level).Equal(*dg2.Level(level)) {
			return false
		}
	}
	return true
}

type DepSet struct {
	*set.HashSet[*service.Service, string]
}

func NewDepSet(capacity int) DepSet {
	return DepSet{set.NewHashSet[*service.Service, string](capacity)}
}

func (ds DepSet) Equal(ds2 DepSet) bool {
	return ds.HashSet.Equal(ds2.HashSet)
}
