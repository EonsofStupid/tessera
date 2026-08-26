package projection

import (
	"fmt"

	"github.com/shippinAI/nomen/internal/eventstore"
	"github.com/shippinAI/nomen/internal/zerrors"
)

func assertEvent[T eventstore.Event](event eventstore.Event) (T, error) {
	e, ok := event.(T)
	if !ok {
		return e, zerrors.CreateNomenError(zerrors.KindInvalidArgument, nil, "HANDL-1m9fS", fmt.Sprintf("reduce.wrong.event.type %T", event), 1)
	}
	return e, nil
}
