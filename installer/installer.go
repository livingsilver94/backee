// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package installer

import (
	"log/slog"

	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/repo/solver"
	"github.com/livingsilver94/backee/service"
)

type Installer struct {
	repository repo.Repo
	variables  repo.Variables
	list       List
}

func New(repository repo.Repo, options ...Option) Installer {
	i := Installer{
		repository: repository,
		variables:  repo.NewVariables(),
		list:       NewList(),
	}
	i.variables.RegisterSolver(service.Datadir, solver.NewDatadir(repository))
	for _, option := range options {
		option(&i)
	}
	return i
}

func (inst *Installer) Install(srv *service.Service) error {
	if srv == nil {
		return nil
	}

	depGraph, err := inst.repository.ResolveDeps(srv)
	if err != nil {
		return err
	}
	for level := depGraph.Depth() - 1; level >= 0; level-- {
		depGraph.Level(level).ForEach(func(dep *service.Service) bool {
			err = inst.installSingle(dep)
			return err == nil
		})
		if err != nil {
			return err
		}
	}
	return inst.installSingle(srv)
}

func (inst *Installer) installSingle(srv *service.Service) error {
	log := slog.Default().WithGroup(srv.Name)
	if inst.list.Contains(srv.Name) {
		log.Info("Already installed")
		return nil
	}
	err := inst.variables.InsertMany(srv.Name, srv.Variables)
	if err != nil {
		return err
	}
	repl := NewTemplate(srv.Name, inst.variables)
	steps := []Step{
		Setup{},
		OSPackages{},
		NewSymlinks(inst.repository, repl),
		NewCopies(inst.repository, repl),
		NewFinalization(repl),
	}
	for _, step := range steps {
		err := step.Run(log, srv)
		if err != nil {
			return err
		}
	}
	inst.list.Insert(srv.Name)
	return nil
}

type Option func(*Installer)

func WithCommonVars(vars map[string]string) Option {
	return func(i *Installer) {
		i.variables.Common = vars
	}
}

func WithVarSolvers(solvs map[service.VarKind]repo.VarSolver) Option {
	return func(i *Installer) {
		for kind, solv := range solvs {
			i.variables.RegisterSolver(kind, solv)
		}
	}
}

func WithList(li List) Option {
	return func(i *Installer) {
		i.list = li
	}
}
