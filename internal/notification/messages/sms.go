package messages

import (
	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/notification/channels"
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
