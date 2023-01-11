package installer

import (
	"fmt"

	"github.com/livingsilver94/backee/service"
)

type VarCache struct {
	vars   map[string]map[string]string
	stores map[service.VarKind]VarStore
}

func NewVarCache() VarCache {
	return VarCache{
		vars:   make(map[string]map[string]string),
		stores: make(map[service.VarKind]VarStore),
	}
}

func (c VarCache) Insert(srv, key string, value service.VarValue) error {
	vars, ok := c.vars[srv]
	if !ok {
		c.vars[srv] = make(map[string]string)
		vars = c.vars[srv]
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

func (c VarCache) Get(service, key string) (string, bool) {
	vars, ok := c.vars[service]
	if !ok {
		return "", ok
	}
	return vars[key], ok
}

func (c VarCache) Length() int {
	return len(c.vars)
}

func (c VarCache) GetAll(service string) map[string]string {
	return c.vars[service]
}

func (c VarCache) SetStore(kind service.VarKind, store VarStore) {
	c.stores[kind] = store
}
