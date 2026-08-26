package oidc

import (
	"context"
	"errors"
	"net/http"

	"github.com/shippinAI/nomen/oidc/v3/pkg/oidc"
	"github.com/shippinAI/nomen/oidc/v3/pkg/op"

	"github.com/shippinAI/nomen/backend/v3/instrumentation/logging"
	http_util "github.com/shippinAI/nomen/internal/api/http"
	"github.com/shippinAI/nomen/internal/zerrors"
)

// oidcError ensures [*oidc.Error] and [op.StatusError] types for err.
// It must be used when an error passes the package boundary towards oidc.
// When err is already of the correct type is passed as-is.
// If the err is a Nomen error, it is transformed with a proper HTTP status code.
// Unknown errors are treated as internal server errors.
func oidcError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, op.ErrInvalidRefreshToken) {
		err = zerrors.ThrowInvalidArgument(err, "OIDCS-ef2Gi", "Errors.User.RefreshToken.Invalid")
	}
	var (
		sError op.StatusError
		oError *oidc.Error
		zError *zerrors.NomenError
	)
	if errors.As(err, &sError) || errors.As(err, &oError) {
		return err
	}

	// here we are encountering an error type that is completely unknown to us.
	if !errors.As(err, &zError) {
		err = zerrors.ThrowInternal(err, "OIDC-AhX2u", "Errors.Internal")
		errors.As(err, &zError)
	}

	statusCode, _ := http_util.NomenErrorToHTTPStatusCode(ctx, err)
	newOidcErr := oidc.ErrServerError
	if statusCode < 500 {
		newOidcErr = oidc.ErrInvalidRequest
	}
	oidcErr := newOidcErr().WithParent(err)
	oidcErr.Description = zError.GetMessage()
	return op.NewStatusError(oidcErr, statusCode)
}

func writeRecoverError(w http.ResponseWriter, r *http.Request, err error) {
	op.WriteError(w, r, err, logging.FromCtx(r.Context()))
}
