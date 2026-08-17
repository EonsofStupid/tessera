package model

import es_models "github.com/EonsofStupid/tessera/internal/eventstore/v1/models"

type ProjectRole struct {
	es_models.ObjectRoot

	Key         string
	DisplayName string
	Group       string
}
