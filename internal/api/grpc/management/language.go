package management

import (
	"context"

	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/i18n"
	mgmt_pb "github.com/shippinAI/nomen/pkg/grpc/management"
)

func (s *Server) GetSupportedLanguages(context.Context, *mgmt_pb.GetSupportedLanguagesRequest) (*mgmt_pb.GetSupportedLanguagesResponse, error) {
	return &mgmt_pb.GetSupportedLanguagesResponse{Languages: domain.LanguagesToStrings(i18n.SupportedLanguages())}, nil
}
