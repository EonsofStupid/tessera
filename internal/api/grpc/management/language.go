package management

import (
	"context"

	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/internal/i18n"
	mgmt_pb "github.com/EonsofStupid/tessera/pkg/grpc/management"
)

func (s *Server) GetSupportedLanguages(context.Context, *mgmt_pb.GetSupportedLanguagesRequest) (*mgmt_pb.GetSupportedLanguagesResponse, error) {
	return &mgmt_pb.GetSupportedLanguagesResponse{Languages: domain.LanguagesToStrings(i18n.SupportedLanguages())}, nil
}
