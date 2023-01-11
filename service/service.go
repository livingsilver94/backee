package service

import (
	"errors"
	"io"

	"github.com/hashicorp/go-set"
	"gopkg.in/yaml.v3"
)

const (
	// VarDelimiter delimits a variable name from a simple string.
	// A variable has VarDelimiter before and after its name.
	VarDelimiter = "%"

	// VarDatadir is the variable name that idenfies a Service's
	// data directory path.
	VarDatadir = "datadir"
)

type Service struct {
	Name       string              `yaml:"-"`
	Depends    *DepSet             `yaml:"depends"`
	Setup      *string             `yaml:"setup"`
	PkgManager []string            `yaml:"pkgmanager"`
	Packages   []string            `yaml:"packages"`
	Links      map[string]FilePath `yaml:"links"`
	Variables  map[string]VarValue `yaml:"variables"`
	Copies     map[string]FilePath `yaml:"copies"`
	Finalize   *string             `yaml:"finalize"`
}

func NewFromYAML(name string, yml []byte) (*Service, error) {
	var srv Service
	err := yaml.Unmarshal(yml, &srv)
	srv.Name = name
	return &srv, err
}

func NewFromYAMLReader(name string, rd io.Reader) (*Service, error) {
	var srv Service
	err := yaml.NewDecoder(rd).Decode(&srv)
	srv.Name = name
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return &srv, err
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

func (ds *DepSet) UnmarshalYAML(node *yaml.Node) error {
	var deps []string
	err := node.Decode(&deps)
	if err != nil {
		return err
	}
	*ds = NewDepSetFrom(deps)
	return nil
}

type FilePath struct {
	Path string `yaml:"path"`
	Mode uint16 `yaml:"mode"`
}

func (lp *FilePath) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var path string
		err := node.Decode(&path)
		if err != nil {
			return err
		}
		lp.Path = path
		lp.Mode = 0644
	default:
		type noRecursion FilePath
		var noRec noRecursion
		err := node.Decode(&noRec)
		if err != nil {
			return err
		}
		*lp = FilePath(noRec)
	}
	return nil
}

type VarKind string

const (
	ClearText VarKind = "cleartext"
)

type VarValue struct {
	Kind  VarKind `yaml:"kind"`
	Value string  `yaml:"value"`
}

func (val *VarValue) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var value string
		err := node.Decode(&value)
		if err != nil {
			return err
		}
		val.Kind = ClearText
		val.Value = value
	default:
		type noRecursion VarValue
		var noRec noRecursion
		err := node.Decode(&noRec)
		if err != nil {
			return err
		}
		if noRec.Kind == "" {
			noRec.Kind = ClearText
		}
		*val = VarValue(noRec)
	}
	return nil
}
