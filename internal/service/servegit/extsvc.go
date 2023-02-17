package servegit

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// extSVCID is the external service ID used by sourcegraph's local code
// syncing. We use a hardcoded ID to simplify finding and mutating the
// external service.
const extSVCID = 0xC0DE

func ensureExtSVC(observationCtx *observation.Context, url string) error {
	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, conf.Get().ServiceConnections().PostgresDSN, "servegit")
	if err != nil {
		return errors.Wrap(err, "servegit failed to connect to frontend DB")
	}
	db := database.NewDB(observationCtx.Logger, sqlDB)

	return doEnsureExtSVC(context.Background(), db.ExternalServices(), url)
}

func doEnsureExtSVC(ctx context.Context, store database.ExternalServiceStore, url string) error {
	config, err := json.Marshal(schema.OtherExternalServiceConnection{
		Url:   url,
		Repos: []string{"src-serve"},
	})
	if err != nil {
		return errors.Wrap(err, "failed to marshal external service configuration")
	}

	return store.Upsert(ctx, &types.ExternalService{
		ID:          extSVCID,
		Kind:        extsvc.KindOther,
		DisplayName: "Your local repositories (autogenerated)",
		Config:      extsvc.NewUnencryptedConfig(string(config)),
	})
}
