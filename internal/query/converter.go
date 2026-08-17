package query

import (
	"github.com/EonsofStupid/tessera/internal/domain"
	"github.com/EonsofStupid/tessera/internal/eventstore"
)

func readModelToObjectDetails(model *eventstore.ReadModel) *domain.ObjectDetails {
	return &domain.ObjectDetails{
		Sequence:      model.ProcessedSequence,
		ResourceOwner: model.ResourceOwner,
		EventDate:     model.ChangeDate,
	}
}
