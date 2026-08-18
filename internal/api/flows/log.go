package flows

import "log/slog"

func logErr(err error) {
	slog.Error("flow execution failed", "err", err)
}
