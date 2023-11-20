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
	writer     StepWriter
}

func New(repository repo.Repo, options ...Option) Installer {
	i := Installer{
		repository: repository,
		variables:  repo.NewVariables(),
		list:       NewList(),
		writer:     nil, // TODO: pass a real writer.
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
			err = inst.InstallSingle(dep)
			return err == nil
		})
		if err != nil {
			return err
		}
	}
	return inst.InstallSingle(srv)
}

func (inst *Installer) InstallSingle(srv *service.Service) error {
	if inst.list.Contains(srv.Name) {
		slog.Default().WithGroup(srv.Name).Info("Already installed")
		return nil
	}
	err := inst.variables.InsertMany(srv.Name, srv.Variables)
	if err != nil {
		return err
	}
	err = inst.runAllSteps(srv)
	if err != nil {
		return err
	}
	inst.list.Insert(srv.Name)
	return nil
}

func (inst *Installer) Steps(srv *service.Service) Steps {
	return NewSteps(srv, inst.writer)
}

func (inst *Installer) runAllSteps(srv *service.Service) error {
	steps := inst.Steps(srv)
	list := []func() error{
		func() error { return steps.Setup() },
		func() error { return steps.InstallPackages() },
		func() error { return steps.LinkFiles(inst.repository, inst.variables) },
		func() error { return steps.CopyFiles(inst.repository, inst.variables) },
		func() error { return steps.Finalize(inst.variables) },
	}
	for _, step := range list {
		err := step()
		if err != nil {
			return err
		}
	}
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

func WithStepWriter(sw StepWriter) Option {
	return func(i *Installer) {
		i.writer = sw
	}
}
