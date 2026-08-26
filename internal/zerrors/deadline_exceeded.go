package zerrors

import "fmt"

func ThrowDeadlineExceeded(parent error, id, message string) error {
	return CreateNomenError(KindDeadlineExceeded, parent, id, message, 1)
}

func ThrowDeadlineExceededf(parent error, id, format string, a ...any) error {
	return CreateNomenError(KindDeadlineExceeded, parent, id, fmt.Sprintf(format, a...), 1)
}

func IsDeadlineExceeded(err error) bool {
	nomenErr, ok := AsNomenError(err)
	return ok && nomenErr.Kind == KindDeadlineExceeded
}
