package repo

import "github.com/livingsilver94/backee/service"

type Repo interface {
	DataDir(srvName string) (string, error)
	LinkDir(srvName string) (string, error)
	ResolveDeps(srv *service.Service) (DepGraph, error)
}
