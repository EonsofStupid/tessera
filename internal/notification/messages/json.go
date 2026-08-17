package messages

import (
	"encoding/json"

	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/notification/channels"
)

var _ channels.Message = (*JSON)(nil)

type JSON struct {
	Serializable        interface{}
	TriggeringEventType eventstore.EventType
}

func (msg *JSON) GetContent() (string, error) {
	bytes, err := json.Marshal(msg.Serializable)
	return string(bytes), err
}

func (msg *JSON) GetTriggeringEventType() eventstore.EventType {
	return msg.TriggeringEventType
}
