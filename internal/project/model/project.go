package model

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	es_models "github.com/EonsofStupid/tessera/internal/eventstore/v1/models"
)

type Project struct {
	es_models.ObjectRoot

	State                  ProjectState
	Name                   string
	Members                []*ProjectMember
	Roles                  []*ProjectRole
	Applications           []*Application
	Grants                 []*ProjectGrant
	ProjectRoleAssertion   bool
	ProjectRoleCheck       bool
	HasProjectCheck        bool
	PrivateLabelingSetting domain.PrivateLabelingSetting
}

type ProjectState int32

const (
	ProjectStateActive ProjectState = iota
	ProjectStateInactive
	ProjectStateRemoved
)
