package service

import (
	"io"
	"strings"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-set"
)

type Service struct {
	Name       string              `yaml:"-"`
	Depends    *DepSet             `yaml:"depends"`
	Setup      *string             `yaml:"setup"`
	PkgManager []string            `yaml:"pkgmanager"`
	Packages   []string            `yaml:"packages"`
	Links      *LinkMap            `yaml:"links"`
	Variables  map[string]VarValue `yaml:"variables"`
	Finalize   *string             `yaml:"finalize"`
}

func NewFromYAML(name string, yml []byte) (*Service, error) {
	var srv Service
	err := yaml.Unmarshal(yml, &srv)
	if err != nil {
		return nil, err
	}
	srv.Name = name
	return &srv, nil
}

func NewFromYAMLReader(name string, rd io.Reader) (*Service, error) {
	var srv Service
	err := yaml.NewDecoder(rd).Decode(&srv)
	if err != nil {
		return nil, err
	}
	srv.Name = name
	return &srv, nil
}

func (srv *Service) Hash() string {
	return srv.Name
}

type DepSet struct {
	*set.Set[string]
}

func NewDepSet(capacity int) DepSet {
	return DepSet{set.New[string](capacity)}
}

func NewDepSetFrom(items []string) DepSet {
	return DepSet{set.From(items)}
}

func (ds DepSet) Equal(ds2 DepSet) bool {
	return ds.Set.Equal(ds2.Set)
}

func (ds *DepSet) UnmarshalYAML(data []byte) error {
	var deps []string
	err := yaml.Unmarshal(data, &deps)
	if err != nil {
		return err
	}
	*ds = NewDepSetFrom(deps)
	return nil
}

type LinkMap struct {
	*orderedmap.OrderedMap[string, LinkParams]
}

func NewLinkMap() LinkMap {
	return LinkMap{orderedmap.NewOrderedMap[string, LinkParams]()}
}

func (lm *LinkMap) Equal(lm2 LinkMap) bool {
	if lm.Len() != lm2.Len() {
		return false
	}
	el1 := lm.Front()
	el2 := lm2.Front()
	for el1 != nil {
		if el1.Key != el2.Key || el1.Value != el2.Value {
			return false
		}
		el1 = el1.Next()
		el2 = el2.Next()
	}
	return true
}

func (lm *LinkMap) UnmarshalYAML(data []byte) error {
	m := make(map[string]LinkParams)
	err := yaml.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	*lm = NewLinkMap()
	for path, params := range m {
		lm.Set(path, params)
	}
	return nil
}

type LinkParams struct {
	Path string `yaml:"path"`
	Mode uint16 `yaml:"mode"`
}

func (lp *LinkParams) UnmarshalYAML(data []byte) error {
	var path string
	err := yaml.Unmarshal(data, &path)
	if err != nil {
		// FIXME: https://github.com/goccy/go-yaml/issues/338
		if !strings.Contains(err.Error(), "of type") {
			return err
		}
		type noRecursion LinkParams
		var noRec noRecursion
		err := yaml.Unmarshal(data, &noRec)
		*lp = LinkParams(noRec)
		return err
	}
	lp.Path = path
	lp.Mode = 0644
	return nil
}

type VarKind string

const (
	ClearText VarKind = "cleartext"
	Secret    VarKind = "secret"
)

type VarValue struct {
	Kind  VarKind `yaml:"kind"`
	Value string  `yaml:"value"`
}

func (val *VarValue) UnmarshalYAML(data []byte) error {
	var value string
	err := yaml.Unmarshal(data, &value)
	if err != nil {
		// FIXME: https://github.com/goccy/go-yaml/issues/338
		if !strings.Contains(err.Error(), "of type") {
			return err
		}
		type noRecursion VarValue
		var noRec noRecursion
		err := yaml.Unmarshal(data, &noRec)
		*val = VarValue(noRec)
		return err
	}
	val.Kind = ClearText
	val.Value = value
	return nil
}

const (
	VariableDatadir = "datadir"
)

func VarPlaceholder(varname string) string {
	return "%" + varname + "%"
}
