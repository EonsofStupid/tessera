package zerrors

import "fmt"

func ThrowNotFound(parent error, id, message string) error {
	return CreateNomenError(KindNotFound, parent, id, message, 1)
}

func ThrowNotFoundf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindNotFound, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsNotFound(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindNotFound
}
