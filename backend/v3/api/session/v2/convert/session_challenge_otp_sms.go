package convert

import (
	"github.com/EonsofStupid/tessera/backend/v3/domain"
	session_grpc "github.com/EonsofStupid/tessera/pkg/grpc/session/v2"
)

func ChallengeOTPSMSGRPCToDomain(otpSMSChallenge *session_grpc.RequestChallenges_OTPSMS) *domain.ChallengeTypeOTPSMS {
	if otpSMSChallenge == nil {
		return nil
	}
	return &domain.ChallengeTypeOTPSMS{
		ReturnCode: otpSMSChallenge.GetReturnCode(),
	}
}
