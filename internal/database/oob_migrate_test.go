package database

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestExternalServiceConfigMigrator(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	t.Run("Up/Down/Progress", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalServiceConfigMigratorWithDB(db, testKey{})
		migrator.BatchSize = 2

		requireProgressEqual := func(want float64) {
			t.Helper()

			got, err := migrator.Progress(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%.3f", want) != fmt.Sprintf("%.3f", got) {
				t.Fatalf("invalid progress: want %f, got %f", want, got)
			}
		}

		// progress on empty table should be 1
		requireProgressEqual(1)

		// Create 10 external services
		svcs := types.GenerateExternalServices(10, types.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := ExternalServices(db).Create(ctx, confGet, svc); err != nil {
				t.Fatal(err)
			}
		}

		// progress on non-migrated table should be 0
		requireProgressEqual(0)

		// Up should migrate two configs
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}
		// services: 10, migrated: 2, progress: 20%
		requireProgressEqual(0.2)

		// Let's migrate the other services
		for i := 2; i <= 5; i++ {
			if err := migrator.Up(ctx); err != nil {
				t.Fatal(err)
			}
			requireProgressEqual(float64(i) * 0.2)
		}
		requireProgressEqual(1)

		// Down should revert the migration for 2 services
		if err := migrator.Down(ctx); err != nil {
			t.Fatal(err)
		}
		// services: 10, migrated: 8, progress: 80%
		requireProgressEqual(0.8)

		// Let's revert the other services
		for i := 3; i >= 0; i-- {
			if err := migrator.Down(ctx); err != nil {
				t.Fatal(err)
			}
			requireProgressEqual(float64(i) * 0.2)
		}
		requireProgressEqual(0)
	})

	t.Run("Up/Encryption", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalServiceConfigMigratorWithDB(db, testKey{})
		migrator.BatchSize = 10

		// Create 10 external services
		svcs := types.GenerateExternalServices(10, types.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := ExternalServices(db).Create(ctx, confGet, svc); err != nil {
				t.Fatal(err)
			}
		}

		// migrate the services
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}

		// was the config actually encrypted?
		rows, err := db.Query("SELECT config, encryption_key_id FROM external_services ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		key := testKey{}

		var i int
		for rows.Next() {
			var config, keyID string

			err = rows.Scan(&config, &keyID)
			if err != nil {
				t.Fatal(err)
			}

			if config == svcs[i].Config {
				t.Fatalf("stored config is the same as before migration")
			}

			secret, err := key.Decrypt(ctx, []byte(config))
			if err != nil {
				t.Fatal(err)
			}

			if secret.Secret() != svcs[i].Config {
				t.Fatalf("decrypted config is different from the original one")
			}

			if id, _ := key.ID(ctx); keyID != id {
				t.Fatalf("wrong encryption_key_id, want %s, got %s", id, keyID)
			}

			i++
		}
		if rows.Err() != nil {
			t.Fatal(err)
		}
	})

	t.Run("Down/Decryption", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		migrator := NewExternalServiceConfigMigratorWithDB(db, testKey{})
		migrator.BatchSize = 10

		// Create 10 external services
		svcs := types.GenerateExternalServices(10, types.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := ExternalServices(db).Create(ctx, confGet, svc); err != nil {
				t.Fatal(err)
			}
		}

		// migrate the services
		if err := migrator.Up(ctx); err != nil {
			t.Fatal(err)
		}

		// revert the migration
		if err := migrator.Down(ctx); err != nil {
			t.Fatal(err)
		}

		// was the config actually reverted?
		rows, err := db.Query("SELECT config, encryption_key_id FROM external_services ORDER BY id")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()

		var i int
		for rows.Next() {
			var config, keyID string

			err = rows.Scan(&config, &keyID)
			if err != nil {
				t.Fatal(err)
			}

			if keyID != "" {
				t.Fatalf("encryption_key_id is still stored in the table")
			}

			if config != svcs[i].Config {
				t.Fatalf("stored config is still encrypted")
			}

			i++
		}
		if rows.Err() != nil {
			t.Fatal(err)
		}
	})

	t.Run("Up/InvalidKey", func(t *testing.T) {
		db := dbtesting.GetDB(t)

		// Create 10 external services
		svcs := types.GenerateExternalServices(10, types.MakeExternalServices()...)
		confGet := func() *conf.Unified {
			return &conf.Unified{}
		}
		for _, svc := range svcs {
			if err := ExternalServices(db).Create(ctx, confGet, svc); err != nil {
				t.Fatal(err)
			}
		}

		// setup invalid key after storing the services
		migrator := NewExternalServiceConfigMigratorWithDB(db, invalidKey{})
		migrator.BatchSize = 10

		// migrate the services, should fail
		err := migrator.Up(ctx)
		if err == nil {
			t.Fatal("migration the service with an invalid key should fail")
		}
		if err.Error() != "invalid encryption round-trip" {
			t.Fatal(err)
		}
	})
}

// invalidKey is an encryption.Key that just base64 encodes the plaintext,
// but silently fails to decrypt the secret.
type invalidKey struct{}

func (k invalidKey) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(plaintext)), nil
}

func (k invalidKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	s := encryption.NewSecret(string(ciphertext))
	return &s, nil
}

func (k invalidKey) ID(ctx context.Context) (string, error) {
	return "invalidKey", nil
}
