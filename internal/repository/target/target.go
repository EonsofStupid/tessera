package target

import (
	"context"
	"time"

	"github.com/EonsofStupid/tessera/internal/crypto"
	"github.com/EonsofStupid/tessera/internal/eventstore"
	targetdomain "github.com/EonsofStupid/tessera/internal/execution/target"
)

const (
	eventTypePrefix  eventstore.EventType = "target."
	AddedEventType                        = eventTypePrefix + "added"
	ChangedEventType                      = eventTypePrefix + "changed"
	RemovedEventType                      = eventTypePrefix + "removed"
)

type AddedEvent struct {
	eventstore.BaseEvent `json:"-"`

	Name             string                   `json:"name"`
	TargetType       targetdomain.TargetType  `json:"targetType"`
	Endpoint         string                   `json:"endpoint"`
	Timeout          time.Duration            `json:"timeout"`
	InterruptOnError bool                     `json:"interruptOnError"`
	SigningKey       *crypto.CryptoValue      `json:"signingKey"`
	PayloadType      targetdomain.PayloadType `json:"payloadType"`
}

func (e *AddedEvent) SetBaseEvent(base *eventstore.BaseEvent) { e.BaseEvent = *base }
func (e *AddedEvent) Payload() any                            { return e }
func (e *AddedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return []*eventstore.UniqueConstraint{NewAddUniqueConstraint(e.Name)}
}

func NewAddedEvent(ctx context.Context, aggregate *eventstore.Aggregate, name string, targetType targetdomain.TargetType, endpoint string, timeout time.Duration, interruptOnError bool, signingKey *crypto.CryptoValue, payloadType targetdomain.PayloadType) *AddedEvent {
	return &AddedEvent{
		BaseEvent:        *eventstore.NewBaseEventForPush(ctx, aggregate, AddedEventType),
		Name:             name,
		TargetType:       targetType,
		Endpoint:         endpoint,
		Timeout:          timeout,
		InterruptOnError: interruptOnError,
		SigningKey:       signingKey,
		PayloadType:      payloadType,
	}
}

type ChangedEvent struct {
	eventstore.BaseEvent `json:"-"`

	Name             *string                  `json:"name,omitempty"`
	TargetType       *targetdomain.TargetType `json:"targetType,omitempty"`
	Endpoint         *string                  `json:"endpoint,omitempty"`
	Timeout          *time.Duration           `json:"timeout,omitempty"`
	InterruptOnError *bool                    `json:"interruptOnError,omitempty"`
	SigningKey       *crypto.CryptoValue      `json:"signingKey,omitempty"`
	PayloadType      targetdomain.PayloadType `json:"payloadType,omitempty"`

	oldName string
}

func (e *ChangedEvent) SetBaseEvent(base *eventstore.BaseEvent) { e.BaseEvent = *base }
func (e *ChangedEvent) Payload() any                            { return e }
func (e *ChangedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	if e.oldName == "" {
		return nil
	}
	return []*eventstore.UniqueConstraint{
		NewRemoveUniqueConstraint(e.oldName),
		NewAddUniqueConstraint(*e.Name),
	}
}

type Changes func(event *ChangedEvent)

func NewChangedEvent(ctx context.Context, aggregate *eventstore.Aggregate, changes []Changes) *ChangedEvent {
	event := &ChangedEvent{BaseEvent: *eventstore.NewBaseEventForPush(ctx, aggregate, ChangedEventType)}
	for _, change := range changes {
		change(event)
	}
	return event
}

func ChangeName(oldName, name string) Changes {
	return func(event *ChangedEvent) {
		event.Name = &name
		event.oldName = oldName
	}
}

func ChangeTargetType(targetType targetdomain.TargetType) Changes {
	return func(event *ChangedEvent) { event.TargetType = &targetType }
}

func ChangeEndpoint(endpoint string) Changes {
	return func(event *ChangedEvent) { event.Endpoint = &endpoint }
}

func ChangeTimeout(timeout time.Duration) Changes {
	return func(event *ChangedEvent) { event.Timeout = &timeout }
}

func ChangeInterruptOnError(interruptOnError bool) Changes {
	return func(event *ChangedEvent) { event.InterruptOnError = &interruptOnError }
}

func ChangeSigningKey(signingKey *crypto.CryptoValue) Changes {
	return func(event *ChangedEvent) { event.SigningKey = signingKey }
}

func ChangePayloadType(payloadType targetdomain.PayloadType) Changes {
	return func(event *ChangedEvent) { event.PayloadType = payloadType }
}

type RemovedEvent struct {
	eventstore.BaseEvent `json:"-"`
	name                 string
}

func (e *RemovedEvent) SetBaseEvent(base *eventstore.BaseEvent) { e.BaseEvent = *base }
func (e *RemovedEvent) Payload() any                            { return e }
func (e *RemovedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return []*eventstore.UniqueConstraint{NewRemoveUniqueConstraint(e.name)}
}

func NewRemovedEvent(ctx context.Context, aggregate *eventstore.Aggregate, name string) *RemovedEvent {
	return &RemovedEvent{*eventstore.NewBaseEventForPush(ctx, aggregate, RemovedEventType), name}
}
