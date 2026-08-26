package convert

import (
	"github.com/shippinAI/nomen/backend/v3/domain"
	session_grpc "github.com/shippinAI/nomen/pkg/grpc/session/v2"
)

func ChallengeOTPSMSGRPCToDomain(otpSMSChallenge *session_grpc.RequestChallenges_OTPSMS) *domain.ChallengeTypeOTPSMS {
	if otpSMSChallenge == nil {
		return nil
	}
	return &domain.ChallengeTypeOTPSMS{
		ReturnCode: otpSMSChallenge.GetReturnCode(),
	}
}
