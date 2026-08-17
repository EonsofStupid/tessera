package auth

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/internal/i18n"
	auth_pb "github.com/EonsofStupid/tessera/pkg/grpc/auth"
)

func (s *Server) GetSupportedLanguages(context.Context, *auth_pb.GetSupportedLanguagesRequest) (*auth_pb.GetSupportedLanguagesResponse, error) {
	return &auth_pb.GetSupportedLanguagesResponse{Languages: domain.LanguagesToStrings(i18n.SupportedLanguages())}, nil
}
