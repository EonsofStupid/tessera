package zerrors

import "fmt"

func ThrowPreconditionFailed(parent error, id, message string) error {
	return CreateNomenError(KindPreconditionFailed, parent, id, message, 1)
}

func ThrowPreconditionFailedf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindPreconditionFailed, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsPreconditionFailed(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindPreconditionFailed
}
