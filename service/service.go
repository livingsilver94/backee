package service

import (
	"errors"
	"io"

	"github.com/hashicorp/go-set"
	"gopkg.in/yaml.v3"
)

const (
	// VarDatadir is the variable name that
	// idenfies a Service's data directory path.
	VarDatadir = "datadir"

	// VarOpenTag is the string opening a variable name
	// to be replaced with its real value.
	VarOpenTag = "{{"
	// VarParentSep separates parent service's name with
	// parent service's variable name, when addressing a parent's variable.
	VarParentSep = "."
	// TemplateOpenTag is the string closing a variable name
	// to be replaced with its real value.
	VarCloseTag = "}}"
)

var (
	DefaultPkgManager = []string{"pkcon", "install", "-y"}
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

func New(name string) *Service {
	return &Service{
		Name: name,
		Variables: map[string]VarValue{
			VarDatadir: {Kind: Datadir, Value: name},
		},
		PkgManager: DefaultPkgManager,
	}
}

func NewFromYAML(name string, yml []byte) (*Service, error) {
	srv := New(name)
	return srv, yaml.Unmarshal(yml, srv)
}

func NewFromYAMLReader(name string, rd io.Reader) (*Service, error) {
	srv := New(name)
	err := yaml.NewDecoder(rd).Decode(srv)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return srv, err
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
	Datadir   VarKind = "datadir"
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
