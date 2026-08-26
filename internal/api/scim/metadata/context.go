package metadata

import (
	"context"
	"strings"

	"github.com/shippinAI/nomen/internal/zerrors"
)

const bulkIDPrefix = "bulkid:"

type scimContextKeyType struct{}

var scimContextKey scimContextKeyType

type ScimContextData struct {
	ProvisioningDomain          string
	IgnorePasswordOnCreate      bool
	ExternalIDScopedMetadataKey ScopedKey
	bulkIDMapping               map[string]string
}

func NewScimContextData() ScimContextData {
	return ScimContextData{
		ExternalIDScopedMetadataKey: ScopedKey(KeyExternalId),
		bulkIDMapping:               make(map[string]string),
	}
}

func SetScimContextData(ctx context.Context, data ScimContextData) context.Context {
	return context.WithValue(ctx, scimContextKey, data)
}

func GetScimContextData(ctx context.Context) ScimContextData {
	data, _ := ctx.Value(scimContextKey).(ScimContextData)
	return data
}

func SetScimBulkIDMapping(ctx context.Context, bulkID, nomenID string) context.Context {
	data := GetScimContextData(ctx)
	data.bulkIDMapping[bulkID] = nomenID
	return ctx
}

func ResolveScimBulkIDIfNeeded(ctx context.Context, resourceID string) (string, error) {
	lowerResourceID := strings.ToLower(resourceID)
	if !strings.HasPrefix(lowerResourceID, bulkIDPrefix) {
		return resourceID, nil
	}

	bulkID := strings.TrimPrefix(lowerResourceID, bulkIDPrefix)
	data := GetScimContextData(ctx)
	nomenID, ok := data.bulkIDMapping[bulkID]
	if !ok {
		return bulkID, zerrors.ThrowInvalidArgumentf(nil, "SCIM-BLK4", "Could not resolve bulkID %v to created ID", bulkID)
	}

	return nomenID, nil
}
