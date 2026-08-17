package repository

import (
	"context"

	usr_model "github.com/EonsofStupid/tessera/internal/user/model"
)

type TokenRepository interface {
	TokenByIDs(ctx context.Context, userID, tokenID string) (*usr_model.TokenView, error)
}
