package repo_test

import (
	"testing"

	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

func TestNewDepGraph(t *testing.T) {
	graph := repo.NewDepGraph(10)
	if graph.Depth() != 0 {
		t.Fatalf("expected initial depth zero. Got %d", graph.Depth())
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Level() did not panic with an empty DepGraph")
		}
	}()
	graph.Level(1) // Must panic since depth is zero.
}

func TestInsert1Level(t *testing.T) {
	graph := repo.NewDepGraph(1)
	service := &service.Service{Name: "serv1"}
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
	service1 := &service.Service{Name: "serv1"}
	service2 := &service.Service{Name: "serv2"}
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
