package runner

import (
	"context"
	"fmt"
	"io/fs"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner/testdata"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func makeTestSchemas(t *testing.T) []*schemas.Schema {
	return []*schemas.Schema{
		makeTestSchema(t, "well-formed"),
		makeTestSchema(t, "query-error"),
	}
}

func makeTestSchema(t *testing.T, name string) *schemas.Schema {
	fs, err := fs.Sub(testdata.Content, name)
	if err != nil {
		t.Fatalf("malformed migration definitions %q: %s", name, err)
	}

	definitions, err := definition.ReadDefinitions(fs)
	if err != nil {
		t.Fatalf("malformed migration definitions %q: %s", name, err)
	}

	return &schemas.Schema{
		Name:                name,
		MigrationsTableName: fmt.Sprintf("%s_migrations_table", name),
		FS:                  fs,
		Definitions:         definitions,
	}
}

func overrideSchemas(t *testing.T) {
	liveSchemas := schemas.Schemas
	schemas.Schemas = makeTestSchemas(t)
	t.Cleanup(func() { schemas.Schemas = liveSchemas })
}

func testStoreWithVersion(version int, dirty bool) *MockStore {
	migrationHook := func(ctx context.Context, definition definition.Definition) error {
		version = definition.ID
		return nil
	}

	store := NewMockStore()
	store.LockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)
	store.TryLockFunc.SetDefaultReturn(true, func(err error) error { return err }, nil)
	store.UpFunc.SetDefaultHook(migrationHook)
	store.DownFunc.SetDefaultHook(migrationHook)
	store.VersionFunc.SetDefaultHook(func(ctx context.Context) (int, bool, bool, error) {
		return version, dirty, true, nil
	})

	return store
}

func makeTestRunner(t *testing.T, store Store) *Runner {
	testSchemas := makeTestSchemas(t)
	storeFactories := make(map[string]StoreFactory, len(testSchemas))

	for _, testSchema := range testSchemas {
		storeFactories[testSchema.Name] = func(ctx context.Context) (Store, error) {
			return store, nil
		}
	}

	return NewRunner(storeFactories)
}
