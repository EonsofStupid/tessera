package zerrors

import "fmt"

func ThrowInternal(parent error, id, message string) error {
	return CreateNomenError(KindInternal, parent, id, message, 1)
}

func ThrowInternalf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindInternal, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsInternal(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindInternal
}
