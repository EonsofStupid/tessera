package zerrors

import "fmt"

func ThrowUnauthenticated(parent error, id, message string) error {
	return CreateNomenError(KindUnauthenticated, parent, id, message, 1)
}

func ThrowUnauthenticatedf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindUnauthenticated, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsUnauthenticated(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindUnauthenticated
}
