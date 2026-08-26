package model

import es_models "github.com/shippinAI/nomen/internal/eventstore/v1/models"

type ProjectRole struct {
	es_models.ObjectRoot

	Key         string
	DisplayName string
	Group       string
}
