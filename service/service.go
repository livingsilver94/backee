package service

import (
	"github.com/goccy/go-yaml"
	"github.com/zyedidia/generic/btree"
	"github.com/zyedidia/generic/hashset"
)

type name = string

type Service struct {
	Name       string             `yaml:"-"`
	Depends    *DepSet            `yaml:"depends"`
	Setup      *string            `yaml:"setup"`
	PkgManager []string           `yaml:"pkgmanager"`
	Packages   []string           `yaml:"packages"`
	Links      *LinkMap           `yaml:"links"`
	Variables  map[string]VarKind `yaml:"variables"`
	Finalize   *string            `yaml:"finalize"`
}

func NewFromYaml(name string, yml []byte) (Service, error) {
	var srv Service
	err := yaml.Unmarshal(yml, &srv)
	if err != nil {
		return Service{}, err
	}
	srv.Name = name
	return srv, nil
}

type DepSet struct {
	hashset.Set[string]
}

func (ds *DepSet) UnmarshalYAML(data []byte) error {
	var deps []string
	err := yaml.Unmarshal(data, &deps)
	if err != nil {
		return err
	}
	for _, dep := range deps {
		ds.Put(dep)
	}
	return nil
}

type LinkMap struct {
	btree.Tree[string, LinkParams]
}

func (lm *LinkMap) UnmarshalYAML(data []byte) error {
	m := make(map[string]LinkParams)
	err := yaml.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	for path, params := range m {
		lm.Put(path, params)
	}
	return nil
}

type LinkParams struct {
	Path string
	Mode uint16
}

type VarKind struct{}
