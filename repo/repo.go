// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package repo

import "github.com/livingsilver94/backee/service"

// Repo is capable of fetching services from a source.
type Repo interface {
	// DataDir returns the data directory path of a service.
	DataDir(srvName string) (string, error)
	// Link returns the symlink directory path of a service.
	LinkDir(srvName string) (string, error)
	// ResolveDeps creates the dependency tree of a service.
	ResolveDeps(srv *service.Service) (DepGraph, error)
}
