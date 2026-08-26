package zerrors

import (
	"fmt"
)

func ThrowPermissionDenied(parent error, id, message string) error {
	return CreateNomenError(KindPermissionDenied, parent, id, message, 1)
}

func ThrowPermissionDeniedf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindPermissionDenied, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsPermissionDenied(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindPermissionDenied
}
