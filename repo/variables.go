package repo

import (
	"errors"
	"fmt"

	"github.com/livingsilver94/backee/service"
)

var (
	ErrNoService  = errors.New("service not found")
	ErrNoVariable = errors.New("variable not found")
)

type VarSolver interface {
	Value(varName string) (varValue string, err error)
}

// Variables resolves and caches services' variables.
type Variables struct {
	Common map[string]string

	resolved map[string]value
	solvers  map[service.VarKind]VarSolver
}

func NewVariables() Variables {
	return Variables{
		resolved: make(map[string]value),
		solvers:  make(map[service.VarKind]VarSolver),
	}
}

// AddParent adds parent to the parents of srv.
// That hints the two services are tied together and it may be useful to
// Get parent's variables as well when Getting srv's variables.
// AddParent returns ErrNoService if srv or parent does not exist.
func (vars Variables) AddParent(srv, parent string) error {
	val, ok := vars.resolved[srv]
	if !ok {
		return ErrNoService
	}
	if _, ok := vars.resolved[parent]; !ok {
		return ErrNoService
	}
	val.parents = append(val.parents, parent)
	vars.resolved[srv] = val
	return nil
}

// Insert saves value for a service named srv under key.
// If the value is not clear text, it is resolved immediately and then cached.
// If key is already present for srv, Insert is no-op.
func (vars Variables) Insert(srv, key string, value service.VarValue) error {
	switch _, err := vars.Get(srv, key); err {
	case ErrNoService:
		vars.resolved[srv] = newValue()
	case ErrNoVariable:
		break
	default:
		return nil
	}
	var v string
	if kind := value.Kind; kind == service.ClearText {
		v = value.Value
	} else {
		solv, ok := vars.solvers[kind]
		if !ok {
			return fmt.Errorf("no variable store registered for kind %q", kind)
		}
		var err error
		v, err = solv.Value(value.Value)
		if err != nil {
			return err
		}
	}
	vars.resolved[srv].vars[key] = v
	return nil
}

func (vars Variables) InsertMany(srv string, values map[string]service.VarValue) error {
	for key, value := range values {
		err := vars.Insert(srv, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// Parents returns the parent list of srv.
// If srv does not exist, ErrNoService is returned.
func (vars Variables) Parents(srv string) ([]string, error) {
	val, ok := vars.resolved[srv]
	if !ok {
		return nil, ErrNoService
	}
	return val.parents, nil
}

func (vars Variables) Get(service, key string) (string, error) {
	val, ok := vars.resolved[service]
	if !ok {
		return "", ErrNoService
	}
	variable, ok := val.vars[key]
	if !ok {
		v, ok := vars.Common[key]
		if !ok {
			return "", ErrNoVariable
		}
		variable = v
	}
	return variable, nil
}

func (vars Variables) Length() int {
	return len(vars.resolved)
}

func (vars Variables) RegisterSolver(kind service.VarKind, solv VarSolver) {
	vars.solvers[kind] = solv
}

type value struct {
	parents []string
	vars    map[string]string
}

func newValue() value {
	return value{
		parents: make([]string, 0),
		vars:    make(map[string]string),
	}
}
