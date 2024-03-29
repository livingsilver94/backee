// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package repo

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/livingsilver94/backee/service"
)

var (
	// ErrNoService is returned when no Service could be found.
	ErrNoService = errors.New("service not found")
	// ErrNoVariable is returned when no variable name could be found.
	ErrNoVariable = errors.New("variable not found")
)

// VarSolver resolves the intermediate value of variables.
type VarSolver interface {
	// Value resolves the intermediate value
	// of varName or returns an error.
	Value(varName string) (varValue string, err error)
}

// Variables resolves and caches services' variables.
type Variables struct {
	// Common is an optional collection of variables that services
	// might have in common. It is initially nil.
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
		return errNoService(srv)
	}
	if _, ok := vars.resolved[parent]; !ok {
		return errNoService(parent)
	}
	val.Parents = append(val.Parents, parent)
	vars.resolved[srv] = val
	return nil
}

// Insert saves value for a service named srv under key.
// If the value is not clear text, it is resolved immediately and then cached.
// If key is already present for srv, Insert is no-op.
func (vars Variables) Insert(srv, key string, val service.VarValue) error {
	_, err := vars.Get(srv, key)
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNoService) {
		vars.resolved[srv] = value{Vars: make(map[string]string)}
	}

	var v string
	if kind := val.Kind; kind == service.ClearText {
		v = val.Value
	} else {
		solv, ok := vars.solvers[kind]
		if !ok {
			return fmt.Errorf("no variable store registered for kind %q", kind)
		}
		var err error
		v, err = solv.Value(val.Value)
		if err != nil {
			return err
		}
	}
	vars.resolved[srv].Vars[key] = v
	return nil
}

// InsertMany is a convenience method to Insert multiple values.
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
		return nil, errNoService(srv)
	}
	return val.Parents, nil
}

// Get returns the value of a Service's variable, previously cached via Insert.
// If srv does not exist, ErrNoService is returned.
// If key does not exist, ErrNoVariable is returned.
func (vars Variables) Get(srv, key string) (string, error) {
	val, ok := vars.resolved[srv]
	if !ok {
		return "", errNoService(srv)
	}
	variable, ok := val.Vars[key]
	if !ok {
		v, ok := vars.Common[key]
		if !ok {
			return "", errNoVariable(key)
		}
		variable = v
	}
	return variable, nil
}

// Length returns how many variables were cached.
func (vars Variables) Length() int {
	return len(vars.resolved)
}

// RegisterSolver registers a VarSolver for a VarKind.
func (vars Variables) RegisterSolver(kind service.VarKind, solv VarSolver) {
	vars.solvers[kind] = solv
}

// GobEncode implements the gob.GobEncoder interface.
func (vars Variables) GobEncode() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(vars.Common)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(vars.resolved)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GobDecode implements the gob.GobDecoder interface.
func (vars *Variables) GobDecode(data []byte) error {
	*vars = NewVariables()
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&vars.Common)
	if err != nil {
		return err
	}
	if len(vars.Common) == 0 {
		// gob allocates an empty map even if we encoded a nil map.
		// We want to replicate the behavior of NewVariables, so if
		// gob allocated an empty map, we reset it to nil.
		vars.Common = nil
	}
	err = dec.Decode(&vars.resolved)
	if err != nil {
		return err
	}
	return nil
}

type value struct {
	Parents []string
	Vars    map[string]string
}

func errNoService(name string) error {
	return fmt.Errorf("%q %w", name, ErrNoService)
}

func errNoVariable(name string) error {
	return fmt.Errorf("%q %w", name, ErrNoVariable)
}
