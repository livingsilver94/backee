package installer

import "github.com/livingsilver94/backee/service"

type VarCache map[string]map[string]string

func NewVarCache() VarCache {
	return make(VarCache)
}

func (c VarCache) Insert(srv, key string, value service.VarValue) {
	vars, ok := c[srv]
	if !ok {
		c[srv] = make(map[string]string)
		vars = c[srv]
	}
	switch value.Kind {
	case service.ClearText:
		vars[key] = value.Value
	case service.Secret:
		// TODO
	}
}

func (c VarCache) Get(service, key string) (string, bool) {
	vars, ok := c[service]
	if !ok {
		return "", ok
	}
	return vars[key], ok
}

func (c VarCache) GetAll(service string) map[string]string {
	return c[service]
}
