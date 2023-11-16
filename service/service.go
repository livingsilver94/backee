// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

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
	// DefaultPkgManager is the default package manager command to install OS packages.
	DefaultPkgManager = []string{"pkcon", "install", "-y"}
)

// Service is a collection of resources to reinstall/restore on an operating system.
// All fields except Name are optional.
type Service struct {
	// Name uniquely identifies a Service.
	Name string `yaml:"-"`

	// Depends is a set of Service names upon which this Service depends.
	Depends *DepSet `yaml:"depends"`

	// Setup is a script (UNIX Shell or Powershell, depending on the operating system)
	// to run before reinstalling and/or restoring any resources.
	Setup *string `yaml:"setup"`

	// PkgManager is combination of command name
	// and arguments to reinstall operating system packages.
	// It must accept a list of package names appended to it.
	PkgManager []string `yaml:"pkgmanager"`

	// Packages is a list of operating system packages to reinstall.
	// It will be appended as-is to PkgManager.
	Packages []string `yaml:"packages"`

	// Links is a collection of symlinks to restore. Their source path
	// is relative to Service's linkdir.
	Links map[string]FilePath `yaml:"links"`

	// Variables is a collection of local variables associated to this Service.
	// A template engine may use them to edit files in place before copying them,
	// or to customize Setup and Finalize scripts.
	// Variables will always contain at least VarDatadir of Datadir kind.
	Variables map[string]VarValue `yaml:"variables"`

	// Copies is a collection of files to copy. Their source path
	// is relative to Service's datadir. A template engine may use Variables
	// to customize the content.
	Copies map[string]FilePath `yaml:"copies"`

	// Setup is a script (UNIX Shell or Powershell, depending on the operating system)
	// to run after reinstalling and/or restoring any resources.
	Finalize *string `yaml:"finalize"`
}

// New creates a Service with a given name. Variables will contain VarDatadir
// and PkgManager will be DefaultPkgManager.
func New(name string) *Service {
	return &Service{
		Name: name,
		Variables: map[string]VarValue{
			VarDatadir: {Kind: Datadir, Value: name},
		},
		PkgManager: DefaultPkgManager,
	}
}

// NewFromYAML creates a Service with a given name whose fields are defined by
// a buffered YAML document.
func NewFromYAML(name string, yml []byte) (*Service, error) {
	srv := New(name)
	return srv, yaml.Unmarshal(yml, srv)
}

// NewFromYAML creates a Service with a given name whose fields are defined by
// a streaming YAML document.
func NewFromYAMLReader(name string, rd io.Reader) (*Service, error) {
	srv := New(name)
	err := yaml.NewDecoder(rd).Decode(srv)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return srv, err
}

// Hash returns a string that uniquely identifies this Service.
// It currently returns Name.
func (srv *Service) Hash() string {
	return srv.Name
}

// DepSet is an unordered set of Service names.
type DepSet struct {
	*set.Set[string]
}

// NewDepSet crates an empty DepSet with an initial capacity.
func NewDepSet(capacity int) DepSet {
	return DepSet{set.New[string](capacity)}
}

// NewDepSetFrom creates a DepSet from a list of Service names.
func NewDepSetFrom(items []string) DepSet {
	return DepSet{set.From(items)}
}

// Equal returns true if ds is equal to ds2.
func (ds DepSet) Equal(ds2 DepSet) bool {
	return ds.Set.Equal(ds2.Set)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (ds *DepSet) UnmarshalYAML(node *yaml.Node) error {
	var deps []string
	err := node.Decode(&deps)
	if err != nil {
		return err
	}
	*ds = NewDepSetFrom(deps)
	return nil
}

// FilePath is a filesystem file path with its file mode.
type FilePath struct {
	Path string `yaml:"path"`
	Mode uint16 `yaml:"mode"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (lp *FilePath) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		var path string
		err := node.Decode(&path)
		if err != nil {
			return err
		}
		lp.Path = path
		lp.Mode = 0
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

// VarKind is a kind of variable.
// Any non-ClearText kind will require a solver to extract the value.
type VarKind string

const (
	// ClearText is the simplest VarKind. Any variable of type ClearText
	// has its value immediately accessible.
	ClearText VarKind = "cleartext"

	// Datadir is the path of a Service's data directory.
	Datadir VarKind = "datadir"
)

// VarValue is a variable's value.
type VarValue struct {
	// Kind is the variable kind. Unless it's ClearText, a solver is required
	// to get the real value.
	Kind VarKind `yaml:"kind"`
	// Value is either the final value, if Kind is ClearText, or an intermediate value.
	Value string `yaml:"value"`
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
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
