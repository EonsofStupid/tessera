package admin

import (
	"context"

	"github.com/shippinAI/nomen/internal/api/authz"
	"github.com/shippinAI/nomen/internal/api/grpc/object"
	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/i18n"
	admin_pb "github.com/shippinAI/nomen/pkg/grpc/admin"
)

func (s *Server) GetSupportedLanguages(ctx context.Context, req *admin_pb.GetSupportedLanguagesRequest) (*admin_pb.GetSupportedLanguagesResponse, error) {
	return &admin_pb.GetSupportedLanguagesResponse{Languages: domain.LanguagesToStrings(i18n.SupportedLanguages())}, nil
}

func (s *Server) SetDefaultLanguage(ctx context.Context, req *admin_pb.SetDefaultLanguageRequest) (*admin_pb.SetDefaultLanguageResponse, error) {
	lang, err := domain.ParseLanguage(req.Language)
	if err != nil {
		return nil, err
	}
	details, err := s.command.SetDefaultLanguage(ctx, lang[0])
	if err != nil {
		return nil, err
	}
	return &admin_pb.SetDefaultLanguageResponse{
		Details: object.DomainToChangeDetailsPb(details),
	}, nil
}

func (s *Server) GetDefaultLanguage(ctx context.Context, _ *admin_pb.GetDefaultLanguageRequest) (*admin_pb.GetDefaultLanguageResponse, error) {
	return &admin_pb.GetDefaultLanguageResponse{Language: authz.GetInstance(ctx).DefaultLanguage().String()}, nil
}

func (s *Server) GetAllowedLanguages(ctx context.Context, _ *admin_pb.GetAllowedLanguagesRequest) (*admin_pb.GetAllowedLanguagesResponse, error) {
	allowed := authz.GetInstance(ctx).AllowedLanguages()
	if len(allowed) == 0 {
		allowed = i18n.SupportedLanguages()
	}
	return &admin_pb.GetAllowedLanguagesResponse{Languages: domain.LanguagesToStrings(allowed)}, nil
}
