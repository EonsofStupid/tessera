package management

import (
	"github.com/shippinAI/nomen/internal/api/grpc/server/middleware"
	"github.com/shippinAI/nomen/pkg/grpc/change"
)

func (c *ListUserChangesResponse) Localizers() []middleware.Localizer {
	if c == nil {
		return nil
	}
	return changesLocalizers(c.Result)
}
func (c *ListOrgChangesResponse) Localizers() []middleware.Localizer {
	if c == nil {
		return nil
	}
	return changesLocalizers(c.Result)
}
func (c *ListProjectChangesResponse) Localizers() []middleware.Localizer {
	if c == nil {
		return nil
	}
	return changesLocalizers(c.Result)
}
func (c *ListProjectGrantChangesResponse) Localizers() []middleware.Localizer {
	if c == nil {
		return nil
	}
	return changesLocalizers(c.Result)
}
func (c *ListAppChangesResponse) Localizers() []middleware.Localizer {
	if c == nil {
		return nil
	}
	return changesLocalizers(c.Result)
}

func changesLocalizers(changes []*change.Change) []middleware.Localizer {
	localizers := make([]middleware.Localizer, len(changes))
	for i, item := range changes {
		localizers[i] = item.EventType
	}
	return localizers
}
