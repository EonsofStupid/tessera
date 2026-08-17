package admin

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	admin_pb "github.com/EonsofStupid/tessera/pkg/grpc/admin"
)

func UpdatePrivacyPolicyToDomain(req *admin_pb.UpdatePrivacyPolicyRequest) *domain.PrivacyPolicy {
	return &domain.PrivacyPolicy{
		TOSLink:        req.TosLink,
		PrivacyLink:    req.PrivacyLink,
		HelpLink:       req.HelpLink,
		SupportEmail:   domain.EmailAddress(req.SupportEmail),
		DocsLink:       req.DocsLink,
		CustomLink:     req.CustomLink,
		CustomLinkText: req.CustomLinkText,
	}
}
