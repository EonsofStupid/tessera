package messages

import (
	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/notification/channels"
)

var _ channels.Message = (*SMS)(nil)

type SMS struct {
	SenderPhoneNumber    string
	RecipientPhoneNumber string
	Content              string
	TriggeringEventType  eventstore.EventType

	// VerificationID is set by the sender
	VerificationID *string
	InstanceID     string
	JobID          string
	UserID         string
}

func (msg *SMS) GetContent() (string, error) {
	return msg.Content, nil
}

func (msg *SMS) GetTriggeringEventType() eventstore.EventType {
	return msg.TriggeringEventType
}
