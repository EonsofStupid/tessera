package convert

import (
	"encoding/json"

	session_grpc "github.com/shippinAI/nomen/pkg/grpc/session/v2"
)

func CheckPasskeyGRPCToDomain(checkPasskey *session_grpc.CheckWebAuthN) ([]byte, error) {
	if checkPasskey == nil || checkPasskey.GetCredentialAssertionData() == nil {
		return nil, nil
	}

	return json.Marshal(checkPasskey.GetCredentialAssertionData())
}
