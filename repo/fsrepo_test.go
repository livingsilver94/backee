package repo_test

import (
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/livingsilver94/backee"
	"github.com/livingsilver94/backee/repo"
)

func TestService(t *testing.T) {
	fs := fstest.MapFS{"srv/service.yaml": &fstest.MapFile{}}
	rep := repo.NewFSRepo(fs)
	expected := &backee.Service{Name: "srv", PkgManager: backee.DefaultPkgManager}
	obtained, err := rep.Service("srv")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(obtained, expected) {
		t.Fatalf("expected %v. Got %v", expected, obtained)
	}
}

func TestAllServices(t *testing.T) {
	fs := fstest.MapFS{
		"srv1/service.yaml": &fstest.MapFile{},
		"srv2/service.yaml": &fstest.MapFile{},
		"emptydir":          &fstest.MapFile{Mode: fs.ModeDir},
		"garbage.txt":       &fstest.MapFile{Data: []byte("please ignore"), Mode: 0644},
	}
	rep := repo.NewFSRepo(fs)
	expected := []*backee.Service{
		{Name: "srv1", PkgManager: backee.DefaultPkgManager},
		{Name: "srv2", PkgManager: backee.DefaultPkgManager}}
	obtained, err := rep.AllServices()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(obtained, expected) {
		t.Fatalf("expected %v. Got %v", expected, obtained)
	}
}

func TestResolveDeps(t *testing.T) {
	fs := fstest.MapFS{
		"lvl1-1/service.yaml": &fstest.MapFile{Data: []byte(`depends: ["lvl2-1"]`)},
		"lvl1-2/service.yaml": &fstest.MapFile{Data: []byte(`depends: ["lvl2-2"]`)},
		"lvl2-1/service.yaml": &fstest.MapFile{},
		"lvl2-2/service.yaml": &fstest.MapFile{},
	}
	deps := backee.NewDepSet(2)
	for _, name := range []string{"lvl1-1", "lvl1-2"} {
		deps.Insert(name)
	}
	srv := &backee.Service{
		Name:    "srv",
		Depends: &deps,
	}
	rep := repo.NewFSRepo(fs)

	expected := repo.NewDepGraph(2)
	expected.Insert(0, &backee.Service{Name: "lvl1-1"})
	expected.Insert(0, &backee.Service{Name: "lvl1-2"})
	expected.Insert(1, &backee.Service{Name: "lvl2-1"})
	expected.Insert(1, &backee.Service{Name: "lvl2-2"})
	obtained, err := rep.ResolveDeps(srv)
	if err != nil {
		t.Fatal(err)
	}
	if !obtained.Equal(expected) {
		t.Fatalf("expected %v. Got %v", expected, obtained)
	}
}
