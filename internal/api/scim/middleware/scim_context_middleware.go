package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/shippinAI/nomen/logging"

	"github.com/shippinAI/nomen/internal/api/authz"
	zhttp "github.com/shippinAI/nomen/internal/api/http/middleware"
	smetadata "github.com/shippinAI/nomen/internal/api/scim/metadata"
	sresources "github.com/shippinAI/nomen/internal/api/scim/resources"
	"github.com/shippinAI/nomen/internal/query"
	"github.com/shippinAI/nomen/internal/zerrors"
)

func ScimContextMiddleware(q *query.Queries) func(next zhttp.HandlerFuncWithError) zhttp.HandlerFuncWithError {
	return func(next zhttp.HandlerFuncWithError) zhttp.HandlerFuncWithError {
		return func(w http.ResponseWriter, r *http.Request) error {
			ctx, err := initScimContext(r.Context(), q)
			if err != nil {
				return err
			}

			return next(w, r.WithContext(ctx))
		}
	}
}

func initScimContext(ctx context.Context, q *query.Queries) (context.Context, error) {
	data := smetadata.NewScimContextData()
	ctx = smetadata.SetScimContextData(ctx, data)

	userID := authz.GetCtxData(ctx).UserID

	// get the provisioningDomain and ignorePasswordOnCreate metadata keys associated with the service account
	metadataKeys := []smetadata.Key{
		smetadata.KeyProvisioningDomain,
		smetadata.KeyIgnorePasswordOnCreate,
	}
	queries := sresources.BuildMetadataQueries(ctx, metadataKeys)

	metadataList, err := q.SearchUserMetadata(ctx, false, userID, queries, nil)
	if err != nil {
		if zerrors.IsNotFound(err) {
			return ctx, nil
		}
		return ctx, err
	}

	if metadataList == nil || len(metadataList.Metadata) == 0 {
		return ctx, nil
	}

	for _, metadata := range metadataList.Metadata {
		switch metadata.Key {
		case string(smetadata.KeyProvisioningDomain):
			data.ProvisioningDomain = string(metadata.Value)
			if data.ProvisioningDomain != "" {
				data.ExternalIDScopedMetadataKey = smetadata.ScopeExternalIdKey(data.ProvisioningDomain)
			}
		case string(smetadata.KeyIgnorePasswordOnCreate):
			ignorePasswordOnCreate, parseErr := strconv.ParseBool(strings.TrimSpace(string(metadata.Value)))
			if parseErr != nil {
				return ctx,
					zerrors.ThrowInvalidArgumentf(nil, "SMCM-yvw2rt", "Invalid value for metadata key %s: %s", smetadata.KeyIgnorePasswordOnCreate, metadata.Value)
			}
			data.IgnorePasswordOnCreate = ignorePasswordOnCreate
		default:
			logging.WithFields("user_metadata_key", metadata.Key).Warn("unexpected metadata key")
		}
	}
	return smetadata.SetScimContextData(ctx, data), nil
}
