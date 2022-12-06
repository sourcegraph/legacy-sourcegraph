package codeintel

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"

	stores "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSCIPMigrator(t *testing.T) {
	logger := logtest.Scoped(t)
	rawDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, rawDB)
	codeIntelDB := stores.NewCodeIntelDB(logger, rawDB)
	store := basestore.NewWithHandle(db.Handle())
	codeIntelStore := basestore.NewWithHandle(codeIntelDB.Handle())
	migrator := NewSCIPMigrator(store, codeIntelStore)
	ctx := context.Background()

	contents, err := os.ReadFile("./testdata/lsif.sql")
	if err != nil {
		t.Fatalf("unexpected error reading file: %s", err)
	}
	if _, err := codeIntelDB.ExecContext(ctx, string(contents)); err != nil {
		t.Fatalf("unexpected error executing test file: %s", err)
	}

	assertProgress := func(expectedProgress float64, applyReverse bool) {
		if progress, err := migrator.Progress(context.Background(), applyReverse); err != nil {
			t.Fatalf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. want=%.2f have=%.2f", expectedProgress, progress)
		}
	}

	// assertCounts := func(expectedCounts []int) {
	// 	query := sqlf.Sprintf(`SELECT num_locations FROM lsif_data_definitions ORDER BY scheme, identifier`)

	// 	if counts, err := basestore.ScanInts(store.Query(context.Background(), query)); err != nil {
	// 		t.Fatalf("unexpected error querying num diagnostics: %s", err)
	// 	} else if diff := cmp.Diff(expectedCounts, counts); diff != "" {
	// 		t.Errorf("unexpected counts (-want +got):\n%s", diff)
	// 	}
	// }

	// n := 500
	// expectedCounts := make([]int, 0, n)
	// locations := make([]LocationData, 0, n)

	// for i := 0; i < n; i++ {
	// 	expectedCounts = append(expectedCounts, i+1)
	// 	locations = append(locations, LocationData{URI: fmt.Sprintf("file://%d", i)})

	// 	data, err := serializer.MarshalLocations(locations)
	// 	if err != nil {
	// 		t.Fatalf("unexpected error serializing locations: %s", err)
	// 	}

	// 	if err := store.Exec(context.Background(), sqlf.Sprintf(
	// 		"INSERT INTO lsif_data_definitions (dump_id, scheme, identifier, data, schema_version, num_locations) VALUES (%s, %s, %s, %s, 1, 0)",
	// 		42+i/(n/2), // 50% id=42, 50% id=43
	// 		fmt.Sprintf("s%04d", i),
	// 		fmt.Sprintf("i%04d", i),
	// 		data,
	// 	)); err != nil {
	// 		t.Fatalf("unexpected error inserting row: %s", err)
	// 	}
	// }

	fmt.Printf("YUP!!\n") // TODO

	assertProgress(0, false)

	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(0.5, false)

	// if err := migrator.Up(context.Background()); err != nil {
	// 	t.Fatalf("unexpected error performing up migration: %s", err)
	// }
	// assertProgress(1, false)

	// assertCounts(expectedCounts)

	// if err := migrator.Down(context.Background()); err != nil {
	// 	t.Fatalf("unexpected error performing down migration: %s", err)
	// }
	// assertProgress(0.5, true)

	// if err := migrator.Down(context.Background()); err != nil {
	// 	t.Fatalf("unexpected error performing down migration: %s", err)
	// }
	// assertProgress(0, true)
}
