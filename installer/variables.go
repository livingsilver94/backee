package installer

import (
	"errors"
	"fmt"

	"github.com/livingsilver94/backee/service"
)

var (
	ErrNoService  = errors.New("service not found")
	ErrNoVariable = errors.New("variable not found")
)

type serviceName = string

type Variables struct {
	resolved map[serviceName]value
	stores   map[service.VarKind]VarStore
}

func NewVariables() Variables {
	return Variables{
		resolved: make(map[serviceName]value),
		stores:   make(map[service.VarKind]VarStore),
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
func (c Variables) Insert(srv, key string, value service.VarValue) error {
	switch _, err := c.Get(srv, key); err {
	case ErrNoService:
		c.resolved[srv] = newValue()
	case ErrNoVariable:
		break
	default:
		return nil
	}
	var v string
	if kind := value.Kind; kind == service.ClearText {
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

func (c Variables) InsertMany(srv string, values map[string]service.VarValue) error {
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
		return "", ErrNoVariable
	}
	return variable, nil
}

func (c Variables) Length() int {
	return len(c.resolved)
}

func (c Variables) GetAll(service string) map[string]string {
	return c.resolved[service].vars
}

func (c Variables) RegisterStore(kind service.VarKind, store VarStore) {
	c.stores[kind] = store
}

type value struct {
	parents []serviceName
	vars    map[string]string
}

func newValue() value {
	return value{
		parents: make([]string, 0),
		vars:    make(map[string]string),
	}
}
