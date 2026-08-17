package model

import (
	"github.com/EonsofStupid/tessera/internal/eventstore/v1/models"
)

type DomainPolicy struct {
	models.ObjectRoot

	State                 PolicyState
	UserLoginMustBeDomain bool
	Default               bool
}
