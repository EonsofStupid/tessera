package model

import es_models "github.com/shippinAI/nomen/internal/eventstore/v1/models"

type ProjectMember struct {
	es_models.ObjectRoot

	UserID string
	Roles  []string
}
