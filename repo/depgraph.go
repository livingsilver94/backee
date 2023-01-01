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

type DepSet struct {
	*set.HashSet[*service.Service, string]
}

func NewDepSet(capacity int) DepSet {
	return DepSet{set.NewHashSet[*service.Service, string](capacity)}
}
