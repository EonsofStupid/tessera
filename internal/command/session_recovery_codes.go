package command

import (
	"context"

	"github.com/shippinAI/nomen/internal/domain"
	"github.com/shippinAI/nomen/internal/eventstore"
)

func CheckRecoveryCode(code string) SessionCommand {
	return func(ctx context.Context, cmd *SessionCommands) ([]eventstore.Command, error) {
		commands, err := checkRecoveryCode(ctx, cmd.sessionWriteModel.UserID, code, cmd.sessionWriteModel.UserResourceOwner, nil, cmd.eventstore.FilterToQueryReducer, cmd.hasher)
		if err != nil {
			return commands, err
		}

		cmd.eventCommands = append(cmd.eventCommands, commands...)
		cmd.RecoveryCodeChecked(ctx, cmd.now())
		return nil, nil
	}
}

func toHumanRecoveryCode(recoveryCodeWriteModel *HumanRecoveryCodeWriteModel) *domain.HumanRecoveryCodes {
	return &domain.HumanRecoveryCodes{
		ObjectDetails: writeModelToObjectDetails(&recoveryCodeWriteModel.WriteModel),
		Codes:         recoveryCodeWriteModel.Codes(),
	}
}
