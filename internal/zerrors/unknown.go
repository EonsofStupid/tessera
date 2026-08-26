package zerrors

import "fmt"

func ThrowUnknown(parent error, id, message string) error {
	return CreateNomenError(KindUnknown, parent, id, message, 1)
}

func ThrowUnknownf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindUnknown, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsUnknown(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindUnknown
}
