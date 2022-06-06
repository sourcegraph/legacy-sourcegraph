package database

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSavedSearchesIsEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	isEmpty, err := db.SavedSearches().IsEmpty(ctx)
	if err != nil {
		t.Fatal()
	}
	want := true
	if want != isEmpty {
		t.Errorf("want %v, got %v", want, isEmpty)
	}

	_, err = db.Users().Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	_, err = db.SavedSearches().Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	isEmpty, err = db.SavedSearches().IsEmpty(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want = false
	if want != isEmpty {
		t.Errorf("want %v, got %v", want, isEmpty)
	}
}

func TestSavedSearchesCreate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	_, err := db.Users().Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := db.SavedSearches().Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}
	if ss == nil {
		t.Fatalf("no saved search returned, create failed")
	}

	want := &types.SavedSearch{
		ID:          1,
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	if !reflect.DeepEqual(ss, want) {
		t.Errorf("query is '%v', want '%v'", ss, want)
	}
}

func TestSavedSearchesUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	_, err := db.Users().Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	_, err = db.SavedSearches().Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	updated := &types.SavedSearch{
		ID:          1,
		Query:       "test2",
		Description: "test2",
		UserID:      &userID,
		OrgID:       nil,
	}

	updatedSearch, err := db.SavedSearches().Update(ctx, updated)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(updatedSearch, updated) {
		t.Errorf("updatedSearch is %v, want %v", updatedSearch, updated)
	}
}

func TestSavedSearchesDelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	_, err := db.Users().Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	_, err = db.SavedSearches().Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	err = db.SavedSearches().Delete(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	allQueries, err := db.SavedSearches().ListAll(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(allQueries) > 0 {
		t.Error("expected no queries in saved_searches table")
	}
}

func TestSavedSearchesGetByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	_, err := db.Users().Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := db.SavedSearches().Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	if ss == nil {
		t.Fatalf("no saved search returned, create failed")
	}
	savedSearch, err := db.SavedSearches().ListSavedSearchesByUserID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := []*types.SavedSearch{{
		ID:          1,
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}}
	if !reflect.DeepEqual(savedSearch, want) {
		t.Errorf("query is '%v+', want '%v+'", savedSearch, want)
	}
}

func TestSavedSearchesGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	_, err := db.Users().Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := db.SavedSearches().Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	if ss == nil {
		t.Fatalf("no saved search returned, create failed")
	}
	savedSearch, err := db.SavedSearches().GetByID(ctx, ss.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := &api.SavedQuerySpecAndConfig{Spec: api.SavedQueryIDSpec{Subject: api.SettingsSubject{User: &userID}, Key: "1"}, Config: api.ConfigSavedQuery{
		Key:         "1",
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}}

	if diff := cmp.Diff(want, savedSearch); diff != "" {
		t.Fatalf("Mismatch (-want +got):\n%s", diff)
	}
}

func TestListSavedSearchesByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	_, err := db.Users().Create(ctx, NewUser{DisplayName: "test", Email: "test@test.com", Username: "test", Password: "test", EmailVerificationCode: "c2"})
	if err != nil {
		t.Fatal("can't create user", err)
	}
	userID := int32(1)
	fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := db.SavedSearches().Create(ctx, fake)
	if err != nil {
		t.Fatal(err)
	}

	if ss == nil {
		t.Fatalf("no saved search returned, create failed")
	}

	org1, err := db.Orgs().Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal("can't create org1", err)
	}
	org2, err := db.Orgs().Create(ctx, "org2", nil)
	if err != nil {
		t.Fatal("can't create org2", err)
	}

	orgFake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		UserID:      nil,
		OrgID:       &org1.ID,
	}
	orgSearch, err := db.SavedSearches().Create(ctx, orgFake)
	if err != nil {
		t.Fatal(err)
	}
	if orgSearch == nil {
		t.Fatalf("no saved search returned, org saved search create failed")
	}

	org2Fake := &types.SavedSearch{
		Query:       "test",
		Description: "test",
		UserID:      nil,
		OrgID:       &org2.ID,
	}
	org2Search, err := db.SavedSearches().Create(ctx, org2Fake)
	if err != nil {
		t.Fatal(err)
	}
	if org2Search == nil {
		t.Fatalf("no saved search returned, org2 saved search create failed")
	}

	_, err = db.OrgMembers().Create(ctx, org1.ID, userID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.OrgMembers().Create(ctx, org2.ID, userID)
	if err != nil {
		t.Fatal(err)
	}

	savedSearches, err := db.SavedSearches().ListSavedSearchesByUserID(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}

	want := []*types.SavedSearch{{
		ID:          1,
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}, {
		ID:          2,
		Query:       "test",
		Description: "test",
		UserID:      nil,
		OrgID:       &org1.ID,
	}, {
		ID:          3,
		Query:       "test",
		Description: "test",
		UserID:      nil,
		OrgID:       &org2.ID,
	}}

	if !reflect.DeepEqual(savedSearches, want) {
		t.Errorf("got %v, want %v", savedSearches, want)
	}
}
