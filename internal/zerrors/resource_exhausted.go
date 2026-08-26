package zerrors

import "fmt"

func ThrowResourceExhausted(parent error, id, message string) error {
	return CreateNomenError(KindResourceExhausted, parent, id, message, 1)
}

func ThrowResourceExhaustedf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindResourceExhausted, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsResourceExhausted(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindResourceExhausted
}
