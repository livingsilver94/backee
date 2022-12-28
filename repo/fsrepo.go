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
	fname := filepath.Join(dir, name, fsRepoFilenamePrefix+fsRepoFilenameSuffix)
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

func (repo FSRepo) DataDir() (string, error) {
	dir := filepath.Join(repo.baseDir, fsRepoBaseDataDir)
	return filepath.Abs(dir)
}

func (repo FSRepo) LinkDir() (string, error) {
	dir := filepath.Join(repo.baseDir, fsRepoBaseLinkDir)
	return filepath.Abs(dir)
}
