package auth

import (
	"context"

	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/i18n"
	auth_pb "github.com/shippinAI/nomen/pkg/grpc/auth"
)

func (s *Server) GetSupportedLanguages(context.Context, *auth_pb.GetSupportedLanguagesRequest) (*auth_pb.GetSupportedLanguagesResponse, error) {
	return &auth_pb.GetSupportedLanguagesResponse{Languages: domain.LanguagesToStrings(i18n.SupportedLanguages())}, nil
}
