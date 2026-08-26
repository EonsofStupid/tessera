package connect_middleware

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/shippinAI/nomen/internal/api/grpc/gerrors"
	_ "github.com/shippinAI/nomen/internal/statik"
	"github.com/shippinAI/nomen/internal/zerrors"
)

func ErrorHandler() connect.UnaryInterceptorFunc {
	return func(handler connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			return toConnectError(ctx, req, handler)
		}
	}
}

func toConnectError(ctx context.Context, req connect.AnyRequest, handler connect.UnaryFunc) (_ connect.AnyResponse, err error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer func() {
		if rec := recover(); rec != nil {
			recErr, ok := rec.(error)
			if !ok {
				recErr = fmt.Errorf("%v", rec)
			}
			if recErr != nil {
				err = zerrors.ThrowInternal(recErr, zerrors.IDRecover, "Errors.Internal")
			}
		}
		cause := err // avoid passing the transport error as cancel cause.
		err = gerrors.NOMENToConnectError(ctx, err)
		cancel(cause)
	}()
	return handler(ctx, req)
}
