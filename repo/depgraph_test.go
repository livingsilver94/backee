package repo_test

import (
	"testing"

	"github.com/livingsilver94/backee"
	"github.com/livingsilver94/backee/repo"
)

func TestInsert1Level(t *testing.T) {
	graph := repo.NewDepGraph(1)
	service := &backee.Service{Name: "serv1"}
	graph.Insert(0, service)
	if graph.Depth() != 1 {
		t.Fatalf("graph depth (%d) went deeper than expected (%d)", graph.Depth(), 1)
	}
	if !graph.Level(0).Contains(service) {
		t.Fatalf("graph doesn't contain service with name %q", service.Name)
	}
}

func TestInsert2Level(t *testing.T) {
	graph := repo.NewDepGraph(1)
	service1 := &backee.Service{Name: "serv1"}
	service2 := &backee.Service{Name: "serv2"}
	graph.Insert(0, service1)
	graph.Insert(1, service2)
	if graph.Depth() != 2 {
		t.Fatalf("graph depth (%d) went deeper than expected (%d)", graph.Depth(), 2)
	}
	if !graph.Level(0).Contains(service1) {
		t.Fatalf("graph doesn't contain service with name %q", service1.Name)
	}
	if !graph.Level(1).Contains(service2) {
		t.Fatalf("graph doesn't contain service with name %q", service2.Name)
	}
}
