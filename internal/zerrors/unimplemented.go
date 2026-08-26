package zerrors

import "fmt"

func ThrowUnimplemented(parent error, id, message string) error {
	return CreateNomenError(KindUnimplemented, parent, id, message, 1)
}

func ThrowUnimplementedf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindUnimplemented, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsUnimplemented(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindUnimplemented
}
