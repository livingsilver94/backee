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
	cache  map[serviceName]value
	stores map[service.VarKind]VarStore
}

func NewVariables() Variables {
	return Variables{
		cache:  make(map[serviceName]value),
		stores: make(map[service.VarKind]VarStore),
	}
}

// Insert saves value for a service named srv under key.
// If the value is not clear text, it is resolved immediately and then cached.
// If key is already present for srv, Insert is no-op.
func (c Variables) Insert(srv, key string, value service.VarValue) error {
	switch _, err := c.Get(srv, key); err {
	case ErrNoService:
		c.cache[srv] = newValue()
	case ErrNoVariable:
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
	c.cache[srv].vars[key] = v
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

func (c Variables) Get(service, key string) (string, error) {
	val, ok := c.cache[service]
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
	return len(c.cache)
}

func (c Variables) GetAll(service string) map[string]string {
	return c.cache[service].vars
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
