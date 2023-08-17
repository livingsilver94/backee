package installer

import (
	"errors"
	"fmt"

	"github.com/livingsilver94/backee"
)

var (
	ErrNoService  = errors.New("service not found")
	ErrNoVariable = errors.New("variable not found")
)

type Variables struct {
	Common map[string]string

	resolved map[string]value
	stores   map[backee.VarKind]VarStore
}

func NewVariables() Variables {
	return Variables{
		resolved: make(map[string]value),
		stores:   make(map[backee.VarKind]VarStore),
	}
}

// AddParent adds parent to the parents of srv.
// That hints the two services are tied together and it may be useful to
// Get parent's variables as well when Getting srv's variables.
// AddParent returns ErrNoService if srv or parent does not exist.
func (c Variables) AddParent(srv, parent string) error {
	val, ok := c.resolved[srv]
	if !ok {
		return ErrNoService
	}
	if _, ok := c.resolved[parent]; !ok {
		return ErrNoService
	}
	val.parents = append(val.parents, parent)
	c.resolved[srv] = val
	return nil
}

// Insert saves value for a service named srv under key.
// If the value is not clear text, it is resolved immediately and then cached.
// If key is already present for srv, Insert is no-op.
func (c Variables) Insert(srv, key string, value backee.VarValue) error {
	switch _, err := c.Get(srv, key); err {
	case ErrNoService:
		c.resolved[srv] = newValue()
	case ErrNoVariable:
		break
	default:
		return nil
	}
	var v string
	if kind := value.Kind; kind == backee.ClearText {
		v = value.Value
	} else {
		store, ok := c.stores[kind]
		if !ok {
			return fmt.Errorf("no variable store registered for kind %q", kind)
		}
		var err error
		v, err = store.Value(value.Value)
		if err != nil {
			return err
		}
	}
	c.resolved[srv].vars[key] = v
	return nil
}

func (c Variables) InsertMany(srv string, values map[string]backee.VarValue) error {
	for key, value := range values {
		err := c.Insert(srv, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// Parents returns the parent list of srv.
// If srv does not exist, ErrNoService is returned.
func (c Variables) Parents(srv string) ([]string, error) {
	val, ok := c.resolved[srv]
	if !ok {
		return nil, ErrNoService
	}
	return val.parents, nil
}

func (c Variables) Get(service, key string) (string, error) {
	val, ok := c.resolved[service]
	if !ok {
		return "", ErrNoService
	}
	variable, ok := val.vars[key]
	if !ok {
		v, ok := c.Common[key]
		if !ok {
			return "", ErrNoVariable
		}
		variable = v
	}
	return variable, nil
}

func (c Variables) Length() int {
	return len(c.resolved)
}

func (c Variables) RegisterStore(kind backee.VarKind, store VarStore) {
	c.stores[kind] = store
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
