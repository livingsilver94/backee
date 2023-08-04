package repo

import (
	"fmt"

	"github.com/hashicorp/go-set"
	"github.com/livingsilver94/backee/service"
)

// DepGraph is the dependency graph of a service, akin to a family tree.
// Levels contain dependencies of dependencies: level 0 contains dependencies
// of the main service, while level 1 contains dependencies of all services in level 1, an so on.
// Levels are 0-indexed.
type DepGraph struct {
	graph []DepSet
}

// NewDepGraph returns an empty DepGraph with an initial depth capacity.
func NewDepGraph(initalDepthCap int) DepGraph {
	return DepGraph{
		graph: make([]DepSet, 0, initalDepthCap),
	}
}

const depSetDefaultCap = 10

// Insert inserts srv at level in the dependency graph.
// Insert panics if level is greater than graph's current depth.
func (dg *DepGraph) Insert(level int, srv *service.Service) {
	if level > dg.Depth() {
		panic(fmt.Sprintf("inserting dep on level %d but level %d does not exist", level, level-1))
	}
	if level == dg.Depth() {
		dg.graph = append(dg.graph, NewDepSet(depSetDefaultCap))
	}
	dg.Level(level).Insert(srv)
}

// Depth returns graph's current depth.
func (dg *DepGraph) Depth() int {
	return len(dg.graph)
}

// Level returns the dependency set for the level.
// Level panics if the level is greater than graph's current depth.
func (dg *DepGraph) Level(index int) *DepSet {
	return &dg.graph[index]
}

// Equal compares the graph with another by inspecting all the levels.
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

// DepSet is a unique set of services.
type DepSet struct {
	*set.HashSet[*service.Service, string]
}

// NewDepSet returns an empty set of services with an initial capacity.
func NewDepSet(initialCap int) DepSet {
	return DepSet{set.NewHashSet[*service.Service, string](initialCap)}
}

// Equal compares the set with another.
func (ds DepSet) Equal(ds2 DepSet) bool {
	return ds.HashSet.Equal(ds2.HashSet)
}
