package zerrors

import "fmt"

func ThrowAlreadyExists(parent error, id, message string) error {
	return CreateNomenError(KindAlreadyExists, parent, id, message, 1)
}

func ThrowAlreadyExistsf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindAlreadyExists, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsErrorAlreadyExists(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindAlreadyExists
}
