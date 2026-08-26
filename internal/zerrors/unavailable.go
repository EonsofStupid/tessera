package zerrors

import "fmt"

func ThrowUnavailable(parent error, id, message string) error {
	return CreateNomenError(KindUnavailable, parent, id, message, 1)
}

func ThrowUnavailablef(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindUnavailable, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsUnavailable(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindUnavailable
}
