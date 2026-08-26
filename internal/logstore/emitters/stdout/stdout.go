package stdout

import (
	"context"
	"encoding/json"

	"github.com/shippinAI/nomen/logging"

	"github.com/shippinAI/nomen/internal/logstore"
)

func NewStdoutEmitter[T logstore.LogRecord[T]]() logstore.LogEmitter[T] {
	return logstore.LogEmitterFunc[T](func(ctx context.Context, bulk []T) error {
		for idx := range bulk {
			bytes, err := json.Marshal(bulk[idx])
			if err != nil {
				return err
			}
			logging.WithFields("record", string(bytes)).Info("log record emitted")
		}
		return nil
	})
}
