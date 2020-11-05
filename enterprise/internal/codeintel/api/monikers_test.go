package api

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	bundlemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client/mocks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
)

func TestLookupMoniker(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	setMockBundleManagerClientPackageInformation(t, mockBundleManagerClient, 42, "sub2/main.go", "1234", testPackageInformation)
	setMockStoreGetPackage(t, mockStore, "gomod", "leftpad", "0.1.0", testDump2, true)
	setMockBundleManagerClientMonikerResults(t, mockBundleManagerClient, 50, "definition", "gomod", "pad", 10, 5, []bundles.Location{
		{DumpID: 42, Path: "foo.go", Range: testRange1},
		{DumpID: 42, Path: "bar.go", Range: testRange2},
		{DumpID: 42, Path: "baz.go", Range: testRange3},
		{DumpID: 42, Path: "bar.go", Range: testRange4},
		{DumpID: 42, Path: "baz.go", Range: testRange5},
	}, 15)

	locations, totalCount, err := lookupMoniker(mockStore, mockBundleManagerClient, 42, "sub2/main.go", "definition", testMoniker2, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error querying moniker: %s", err)
	}
	if totalCount != 15 {
		t.Errorf("unexpected total count. want=%d have=%d", 5, totalCount)
	}

	expectedLocations := []ResolvedLocation{
		{Dump: testDump2, Path: "sub2/foo.go", Range: testRange1},
		{Dump: testDump2, Path: "sub2/bar.go", Range: testRange2},
		{Dump: testDump2, Path: "sub2/baz.go", Range: testRange3},
		{Dump: testDump2, Path: "sub2/bar.go", Range: testRange4},
		{Dump: testDump2, Path: "sub2/baz.go", Range: testRange5},
	}
	if diff := cmp.Diff(expectedLocations, locations); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}

func TestLookupMonikerNoPackageInformationID(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	_, totalCount, err := lookupMoniker(mockStore, mockBundleManagerClient, 42, "sub/main.go", "definition", testMoniker3, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error querying moniker: %s", err)
	}
	if totalCount != 0 {
		t.Errorf("unexpected total count. want=%d have=%d", 0, totalCount)
	}
}

func TestLookupMonikerNoPackage(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()

	setMockBundleManagerClientPackageInformation(t, mockBundleManagerClient, 42, "main.go", "1234", testPackageInformation)
	setMockStoreGetPackage(t, mockStore, "gomod", "leftpad", "0.1.0", store.Dump{}, false)

	_, totalCount, err := lookupMoniker(mockStore, mockBundleManagerClient, 42, "main.go", "definition", testMoniker1, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error querying moniker: %s", err)
	}
	if totalCount != 0 {
		t.Errorf("unexpected total count. want=%d have=%d", 0, totalCount)
	}
}
