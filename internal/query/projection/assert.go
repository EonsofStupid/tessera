package projection

import (
	"fmt"

	"github.com/EonsofStupid/tessera/internal/eventstore"
	"github.com/EonsofStupid/tessera/internal/zerrors"
)

func assertEvent[T eventstore.Event](event eventstore.Event) (T, error) {
	e, ok := event.(T)
	if !ok {
		return e, zerrors.CreateZitadelError(zerrors.KindInvalidArgument, nil, "HANDL-1m9fS", fmt.Sprintf("reduce.wrong.event.type %T", event), 1)
	}
	return e, nil
}
