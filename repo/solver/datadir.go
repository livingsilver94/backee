// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package solver

import "github.com/livingsilver94/backee/repo"

type Datadir struct {
	repo repo.Repo
}

func NewDatadir(repo repo.Repo) Datadir {
	return Datadir{
		repo: repo,
	}
}

func (d Datadir) Value(varName string) (varValue string, err error) {
	return d.repo.DataDir(varName)
}
