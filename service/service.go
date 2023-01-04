package service

import (
	"errors"
	"io"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-set"
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

func (ds *DepSet) UnmarshalYAML(data []byte) error {
	var deps []string
	err := yaml.Unmarshal(data, &deps)
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

func (lp *FilePath) UnmarshalYAML(data []byte) error {
	var path string
	err := yaml.Unmarshal(data, &path)
	if err != nil {
		// FIXME: https://github.com/goccy/go-yaml/issues/338
		if !strings.Contains(err.Error(), "of type") {
			return err
		}
		type noRecursion FilePath
		var noRec noRecursion
		err := yaml.Unmarshal(data, &noRec)
		*lp = FilePath(noRec)
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
		if noRec.Kind == "" {
			noRec.Kind = ClearText
		}
		*val = VarValue(noRec)
		return err
	}
	val.Kind = ClearText
	val.Value = value
	return nil
}
