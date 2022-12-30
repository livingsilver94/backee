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
	baseDir string
	variant string
}

func NewFSRepo(baseDir string) FSRepo {
	return FSRepo{
		baseDir: baseDir,
		variant: "",
	}
}

func NewFSRepoVariant(baseDir, variant string) FSRepo {
	return FSRepo{
		baseDir: baseDir,
		variant: variant,
	}
}

func (repo FSRepo) Service(name string) (*service.Service, error) {
	dir := filepath.Join(repo.baseDir, name)
	var fname string
	if repo.variant != "" {
		fname = filepath.Join(dir, name, fsRepoFilenamePrefix+"_"+repo.variant+fsRepoFilenameSuffix)
	} else {
		fname = filepath.Join(dir, name, fsRepoFilenamePrefix+fsRepoFilenameSuffix)
	}
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return service.NewFromYAMLReader(name, file)
}

func (repo FSRepo) AllServices() ([]*service.Service, error) {
	children, err := os.ReadDir(repo.baseDir)
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
	dir := filepath.Join(repo.baseDir, name, fsRepoBaseDataDir)
	return filepath.Abs(dir)
}

func (repo FSRepo) LinkDir(name string) (string, error) {
	dir := filepath.Join(repo.baseDir, name, fsRepoBaseLinkDir)
	return filepath.Abs(dir)
}

func (repo FSRepo) resolveDeps(graph *DepGraph, level int, deps *service.DepSet) error {
	for _, depName := range deps.List() {
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
	level += 1
	subdeps := service.NewDepSet(depSetDefaultCap)
	for _, subdep := range graph.Level(level).List() {
		subdeps.InsertAll(subdep.Depends.List())
	}
	return repo.resolveDeps(graph, level, &subdeps)
}
