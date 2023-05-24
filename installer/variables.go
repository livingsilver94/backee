package installer

import (
	"fmt"

	"github.com/livingsilver94/backee/service"
)

type serviceName = string

type Variables struct {
	vars   map[serviceName]map[string]string
	stores map[service.VarKind]VarStore
}

func NewVariables() Variables {
	return Variables{
		vars:   make(map[serviceName]map[string]string),
		stores: make(map[service.VarKind]VarStore),
	}
}

// Insert saves value for a service named srv under key.
// If key is already present for srv, Insert is no-op.
func (c Variables) Insert(srv, key string, value service.VarValue) error {
	vars, ok := c.vars[srv]
	if !ok {
		c.vars[srv] = make(map[string]string)
		vars = c.vars[srv]
	}
	if _, ok := c.Get(srv, key); ok {
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
	vars[key] = v
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

func (c Variables) Get(service, key string) (string, bool) {
	vars, ok := c.vars[service]
	if !ok {
		return "", false
	}
	val, ok := vars[key]
	return val, ok
}

func (c Variables) Length() int {
	return len(c.vars)
}

func (c Variables) GetAll(service string) map[string]string {
	return c.vars[service]
}

func (c Variables) RegisterStore(kind service.VarKind, store VarStore) {
	c.stores[kind] = store
}
