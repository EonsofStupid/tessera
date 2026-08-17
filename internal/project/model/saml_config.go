package model

import (
	es_models "github.com/EonsofStupid/tessera/internal/eventstore/v1/models"
)

type SAMLConfig struct {
	es_models.ObjectRoot
	AppID       string
	Metadata    []byte
	MetadataURL string
}
