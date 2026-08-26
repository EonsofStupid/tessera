package idp

import (
	"github.com/shippinAI/nomen/internal/crypto"
	"github.com/shippinAI/nomen/internal/eventstore"
)

type RolesInfo struct {
	OrganizationID     string `json:"organizationId"`
	OrganizationDomain string `json:"organizationDomain"`
}

type NomenIDPAddedEvent struct {
	eventstore.BaseEvent `json:"-"`

	ID                string              `json:"id"`
	Name              string              `json:"name"`
	Issuer            string              `json:"issuer"`
	ClientID          string              `json:"clientId"`
	ClientSecret      *crypto.CryptoValue `json:"clientSecret"`
	Scopes            []string            `json:"scopes,omitempty"`
	InstanceRolesInfo []RolesInfo         `json:"instanceRolesInfo,omitempty"`
	Options
}

func NewNomenIDPAddedEvent(
	base *eventstore.BaseEvent,
	id,
	name,
	issuer,
	clientID string,
	clientSecret *crypto.CryptoValue,
	scopes []string,
	options Options,
	instanceRolesInfo []RolesInfo,
) *NomenIDPAddedEvent {
	return &NomenIDPAddedEvent{
		BaseEvent:         *base,
		ID:                id,
		Name:              name,
		Issuer:            issuer,
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		Scopes:            scopes,
		Options:           options,
		InstanceRolesInfo: instanceRolesInfo,
	}
}

func (e *NomenIDPAddedEvent) Payload() interface{} {
	return e
}

func (e *NomenIDPAddedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return nil
}

type NomenIDPChangedEvent struct {
	eventstore.BaseEvent `json:"-"`

	ID                string              `json:"id"`
	Name              *string             `json:"name,omitempty"`
	Issuer            *string             `json:"issuer,omitempty"`
	ClientID          *string             `json:"clientId,omitempty"`
	ClientSecret      *crypto.CryptoValue `json:"clientSecret,omitempty"`
	Scopes            *[]string           `json:"scopes,omitempty"`
	InstanceRolesInfo *[]RolesInfo        `json:"instanceRolesInfo,omitempty"`
	OptionChanges
}

func NewNomenIDPChangedEvent(
	base *eventstore.BaseEvent,
	id string,
	changes []NomenIDPChanges,
) *NomenIDPChangedEvent {
	e := &NomenIDPChangedEvent{
		BaseEvent: *base,
		ID:        id,
	}
	for _, change := range changes {
		change(e)
	}
	return e
}

type NomenIDPChanges func(*NomenIDPChangedEvent)

func ChangeNomenIDPName(name string) NomenIDPChanges {
	return func(e *NomenIDPChangedEvent) {
		e.Name = &name
	}
}

func ChangeNomenIDPIssuer(issuer string) NomenIDPChanges {
	return func(e *NomenIDPChangedEvent) {
		e.Issuer = &issuer
	}
}

func ChangeNomenIDPClientID(clientID string) NomenIDPChanges {
	return func(e *NomenIDPChangedEvent) {
		e.ClientID = &clientID
	}
}

func ChangeNomenIDPClientSecret(clientSecret *crypto.CryptoValue) NomenIDPChanges {
	return func(e *NomenIDPChangedEvent) {
		e.ClientSecret = clientSecret
	}
}

func ChangeNomenIDPScopes(scopes []string) NomenIDPChanges {
	return func(e *NomenIDPChangedEvent) {
		// explicitly set them to empty in case the scopes are unset
		if scopes == nil {
			scopes = make([]string, 0)
		}
		e.Scopes = &scopes
	}
}

func ChangeNomenIDPInstanceRolesInfo(instanceRolesInfo []RolesInfo) NomenIDPChanges {
	return func(e *NomenIDPChangedEvent) {
		// explicitly set them to empty in case the instance roles are unset
		if instanceRolesInfo == nil {
			instanceRolesInfo = make([]RolesInfo, 0)
		}
		e.InstanceRolesInfo = &instanceRolesInfo
	}
}

func ChangeNomenIDPOptions(options OptionChanges) NomenIDPChanges {
	return func(e *NomenIDPChangedEvent) {
		e.OptionChanges = options
	}
}

func (e *NomenIDPChangedEvent) Payload() interface{} {
	return e
}

func (e *NomenIDPChangedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return nil
}
