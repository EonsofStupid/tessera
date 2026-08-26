package zerrors

import "fmt"

func ThrowInvalidArgument(parent error, id, message string) error {
	return CreateNomenError(KindInvalidArgument, parent, id, message, 1)
}

func ThrowInvalidArgumentf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindInvalidArgument, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsErrorInvalidArgument(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindInvalidArgument
}
