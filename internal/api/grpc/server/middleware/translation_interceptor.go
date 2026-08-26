package middleware

import (
	"context"

	"google.golang.org/grpc"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/i18n"
	_ "github.com/shippinAI/nomen/internal/statik"
	"github.com/shippinAI/nomen/internal/telemetry/tracing"
)

func TranslationHandler() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		ctx, span := tracing.NewSpan(ctx)
		defer func() { span.EndWithError(err) }()

		if loc, ok := resp.(localizers); ok && resp != nil {
			translator := getTranslator(ctx)
			translateFields(ctx, loc, translator)
		}
		if err != nil {
			translator := getTranslator(ctx)
			err = translateError(ctx, err, translator)
		}
		return resp, err
	}
}

func getTranslator(ctx context.Context) *i18n.Translator {
	return i18n.NewNomenTranslator(authz.GetInstance(ctx).DefaultLanguage())
}
