package command

import "github.com/shippinAI/nomen/internal/domain"

func userGrantWriteModelToUserGrant(writeModel *UserGrantWriteModel) *domain.UserGrant {
	return &domain.UserGrant{
		ObjectRoot:     writeModelToObjectRoot(writeModel.WriteModel),
		UserID:         writeModel.UserID,
		ProjectID:      writeModel.ProjectID,
		ProjectGrantID: writeModel.ProjectGrantID,
		RoleKeys:       writeModel.RoleKeys,
		State:          writeModel.State,
	}
}
