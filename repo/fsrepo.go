package repo

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/livingsilver94/backee/service"
)

const (
	fsRepoBaseDataDir = "data"
	fsRepoBaseLinkDir = "links"

	fsRepoFilenamePrefix = "service"
	fsRepoFilenameSuffix = ".yaml"
)

type FSRepo struct {
	baseFS  fs.FS
	variant string
}

func NewFSRepo(baseFS fs.FS) FSRepo {
	return FSRepo{
		baseFS:  baseFS,
		variant: "",
	}
}

func NewFSRepoVariant(baseFS fs.FS, variant string) FSRepo {
	return FSRepo{
		baseFS:  baseFS,
		variant: variant,
	}
}

func (repo FSRepo) Service(name string) (*service.Service, error) {
	var fname string
	if repo.variant != "" {
		fname = name + "/" + (fsRepoFilenamePrefix + "_" + repo.variant + fsRepoFilenameSuffix)
	} else {
		fname = name + "/" + (fsRepoFilenamePrefix + fsRepoFilenameSuffix)
	}
	file, err := repo.baseFS.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return service.NewFromYAMLReader(name, file)
}

func (repo FSRepo) AllServices() ([]*service.Service, error) {
	children, err := fs.ReadDir(repo.baseFS, ".")
	if err != nil {
		return nil, err
	}
	services := make([]*service.Service, 0, len(children))
	for _, child := range children {
		if !child.IsDir() {
			continue
		}
		srv, err := repo.Service(child.Name())
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		services = append(services, srv)
	}
	return services, nil
}

const depGraphDefaultDepth = 4

func (repo FSRepo) ResolveDeps(srv *service.Service) (DepGraph, error) {
	graph := NewDepGraph(depGraphDefaultDepth)
	if srv.Depends == nil {
		return graph, nil
	}
	return graph, repo.resolveDeps(&graph, 0, srv.Depends)
}

func (repo FSRepo) DataDir(name string) (string, error) {
	switch typ := repo.baseFS.(type) {
	case OSFS:
		dir := filepath.Join(typ.path, name, fsRepoBaseDataDir)
		return filepath.Abs(dir)
	default:
		panic("unimplemented")
	}
}

func (repo FSRepo) LinkDir(name string) (string, error) {
	switch typ := repo.baseFS.(type) {
	case OSFS:
		dir := filepath.Join(typ.path, name, fsRepoBaseLinkDir)
		return filepath.Abs(dir)
	default:
		panic("unimplemented")
	}
}

func (repo FSRepo) resolveDeps(graph *DepGraph, level int, deps *service.DepSet) error {
	for _, depName := range deps.Slice() {
		srv, err := repo.Service(depName)
		if err != nil {
			return err
		}
		graph.Insert(level, srv)
	}
	if level == graph.Depth() {
		return nil
	}

	// Gather dependencies of dependencies.
	subdeps := service.NewDepSet(depSetDefaultCap)
	for _, subdep := range graph.Level(level).Slice() {
		if subdep.Depends == nil {
			continue
		}
		subdeps.InsertAll(subdep.Depends.Slice())
	}
	return repo.resolveDeps(graph, level+1, &subdeps)
}

// OSFS circumvents the inability to check whether fs.FS
// is a real operating system path. FSRepo will use OSFS
// to discriminate a real OS path from a network, or a
// testing, file system.
//
// Ideally, Go should provide the inverse operation of os.DirFS().
type OSFS struct {
	fs.StatFS
	path string
}

func NewOSFS(path string) OSFS {
	return OSFS{
		StatFS: os.DirFS(path).(fs.StatFS),
		path:   path,
	}
}
