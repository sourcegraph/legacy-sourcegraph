package subscriptions_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hexops/autogold/v2"
	"github.com/hexops/valast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/databasetest"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/tables"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/utctime"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestLicensesStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := databasetest.NewTestDB(t, "enterprise-portal", t.Name(), tables.All()...)

	subscriptionID1 := uuid.NewString()
	subscriptionID2 := uuid.NewString()

	subs := subscriptions.NewStore(db)
	_, err := subs.Upsert(ctx, subscriptionID1, subscriptions.UpsertSubscriptionOptions{
		DisplayName: pointers.Ptr(database.NewNullString("Acme, Inc. 1")),
	})
	require.NoError(t, err)
	_, err = subs.Upsert(ctx, subscriptionID2, subscriptions.UpsertSubscriptionOptions{
		DisplayName: pointers.Ptr(database.NewNullString("Acme, Inc. 2")),
	})
	require.NoError(t, err)

	licenses := subscriptions.NewLicensesStore(db)

	var createdLicenses []*subscriptions.LicenseWithConditions
	getCreatedByLicenseID := func(t *testing.T, licenseID string) *subscriptions.LicenseWithConditions {
		for _, l := range createdLicenses {
			if l.ID == licenseID {
				return l
			}
		}
		t.Errorf("license %q not found", licenseID)
		t.FailNow()
		return nil
	}
	t.Run("CreateLicenseKey", func(t *testing.T) {
		testLicense := func(
			got *subscriptions.LicenseWithConditions,
			wantMessage autogold.Value,
			wantLicenseData autogold.Value,
		) {
			assert.NotEmpty(t, got.ID)
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.ExpireAt)
			assert.Equal(t, "ENTERPRISE_SUBSCRIPTION_LICENSE_TYPE_KEY", got.LicenseType)
			wantLicenseData.Equal(t, string(got.LicenseData))

			assert.Len(t, got.Conditions, 1)
			wantMessage.Equal(t, got.Conditions[0].Message)
			assert.Equal(t, "STATUS_CREATED", got.Conditions[0].Status)
			assert.Equal(t, got.CreatedAt, got.Conditions[0].TransitionTime)
		}

		got, err := licenses.CreateLicenseKey(ctx, subscriptionID1,
			&subscriptions.LicenseKey{
				Info: license.Info{
					Tags:      []string{"foo"},
					CreatedAt: time.Time{}.Add(1 * time.Hour),
					ExpiresAt: time.Time{}.Add(48 * time.Hour),
				},
				SignedKey: "asdfasdf",
			},
			subscriptions.CreateLicenseOpts{
				Message:    t.Name() + " 1 old",
				Time:       pointers.Ptr(utctime.FromTime(time.Time{}.Add(1 * time.Hour))),
				ExpireTime: utctime.FromTime(time.Time{}.Add(48 * time.Hour)),
			})
		require.NoError(t, err)
		testLicense(
			got,
			autogold.Expect(valast.Ptr("TestLicensesStore/CreateLicenseKey 1 old")),
			autogold.Expect(`{"Info": {"c": "0001-01-01T01:00:00Z", "e": "0001-01-03T00:00:00Z", "t": ["foo"], "u": 0}, "SignedKey": "asdfasdf"}`),
		)
		createdLicenses = append(createdLicenses, got)

		got, err = licenses.CreateLicenseKey(ctx, subscriptionID1,
			&subscriptions.LicenseKey{
				Info: license.Info{
					Tags:      []string{"baz"},
					CreatedAt: time.Time{}.Add(24 * time.Hour),
					ExpiresAt: time.Time{}.Add(48 * time.Hour),
				},
				SignedKey: "barasdf",
			},
			subscriptions.CreateLicenseOpts{
				Message:    t.Name() + " 1",
				Time:       pointers.Ptr(utctime.FromTime(time.Time{}.Add(24 * time.Hour))),
				ExpireTime: utctime.FromTime(time.Time{}.Add(48 * time.Hour)),
			})
		require.NoError(t, err)
		testLicense(
			got,
			autogold.Expect(valast.Ptr("TestLicensesStore/CreateLicenseKey 1")),
			autogold.Expect(`{"Info": {"c": "0001-01-02T00:00:00Z", "e": "0001-01-03T00:00:00Z", "t": ["baz"], "u": 0}, "SignedKey": "barasdf"}`),
		)
		createdLicenses = append(createdLicenses, got)

		got, err = licenses.CreateLicenseKey(ctx, subscriptionID2,
			&subscriptions.LicenseKey{
				Info: license.Info{
					Tags:      []string{"tag"},
					CreatedAt: time.Time{}.Add(24 * time.Hour),
					ExpiresAt: time.Time{}.Add(48 * time.Hour),
				},
				SignedKey: "asdffdsadf",
			},
			subscriptions.CreateLicenseOpts{
				Message:    t.Name() + " 2",
				Time:       pointers.Ptr(utctime.FromTime(time.Time{}.Add(24 * time.Hour))),
				ExpireTime: utctime.FromTime(time.Time{}.Add(48 * time.Hour)),
			})
		require.NoError(t, err)
		testLicense(
			got,
			autogold.Expect(valast.Ptr("TestLicensesStore/CreateLicenseKey 2")),
			autogold.Expect(`{"Info": {"c": "0001-01-02T00:00:00Z", "e": "0001-01-03T00:00:00Z", "t": ["tag"], "u": 0}, "SignedKey": "asdffdsadf"}`),
		)
		createdLicenses = append(createdLicenses, got)

		t.Run("createdAt does not match", func(t *testing.T) {
			_, err = licenses.CreateLicenseKey(ctx, subscriptionID2,
				&subscriptions.LicenseKey{
					Info: license.Info{
						Tags:      []string{"tag"},
						CreatedAt: time.Time{}.Add(24 * time.Hour),
					},
					SignedKey: "asdffdsadf",
				},
				subscriptions.CreateLicenseOpts{
					Message: t.Name(),
					Time:    pointers.Ptr(utctime.Now()),
				})
			require.Error(t, err)
			autogold.Expect("creation time must match the license key information").Equal(t, err.Error())
		})
		t.Run("expiresAt does not match", func(t *testing.T) {
			_, err = licenses.CreateLicenseKey(ctx, subscriptionID2,
				&subscriptions.LicenseKey{
					Info: license.Info{
						Tags:      []string{"tag"},
						CreatedAt: time.Time{},
						ExpiresAt: time.Time{}.Add(48 * time.Hour),
					},
					SignedKey: "asdffdsadf",
				},
				subscriptions.CreateLicenseOpts{
					Message:    t.Name(),
					Time:       pointers.Ptr(utctime.FromTime(time.Time{})),
					ExpireTime: utctime.Now(),
				})
			require.Error(t, err)
			autogold.Expect("expiration time must match the license key information").Equal(t, err.Error())
		})
	})

	// No point continuing if test licenses did not create, all tests after this
	// will fail
	if t.Failed() {
		t.FailNow()
	}

	t.Run("List", func(t *testing.T) {
		listedLicenses, err := licenses.List(ctx, subscriptions.ListLicensesOpts{})
		require.NoError(t, err)
		assert.Len(t, listedLicenses, len(createdLicenses))
		for _, l := range listedLicenses {
			created := getCreatedByLicenseID(t, l.ID)
			assert.Equal(t, *created, *l)
		}

		t.Run("List by subscription", func(t *testing.T) {
			listedLicenses, err := licenses.List(ctx, subscriptions.ListLicensesOpts{
				SubscriptionID: subscriptionID1,
			})
			require.NoError(t, err)
			assert.Len(t, listedLicenses, 2)
			for _, l := range listedLicenses {
				assert.Equal(t, subscriptionID1, l.SubscriptionID)
				assert.Equal(t, *getCreatedByLicenseID(t, l.ID), *l)
			}

			listedLicenses, err = licenses.List(ctx, subscriptions.ListLicensesOpts{
				SubscriptionID: subscriptionID2,
			})
			require.NoError(t, err)
			assert.Len(t, listedLicenses, 1)
			for _, l := range listedLicenses {
				assert.Equal(t, subscriptionID2, l.SubscriptionID)
				assert.Equal(t, *getCreatedByLicenseID(t, l.ID), *l)
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		for _, license := range createdLicenses {
			got, err := licenses.Get(ctx, license.ID)
			require.NoError(t, err)
			assert.Equal(t, *license, *got)
		}
	})

	t.Run("Revoke", func(t *testing.T) {
		for idx, license := range createdLicenses {
			revokeTime := utctime.FromTime(time.Now().Add(-time.Second))
			got, err := licenses.Revoke(ctx, license.ID, subscriptions.RevokeLicenseOpts{
				Message: fmt.Sprintf("%s %d", t.Name(), idx),
				Time:    pointers.Ptr(revokeTime),
			})
			require.NoError(t, err)
			assert.Equal(t, revokeTime.AsTime(), got.RevokedAt.AsTime())
			require.Len(t, got.Conditions, 2)
			// Most recent condition is sorted first, and should be the revocation
			assert.Equal(t, "STATUS_REVOKED", got.Conditions[0].Status)
			assert.Equal(t, revokeTime.AsTime(), got.Conditions[0].TransitionTime.AsTime())
			assert.Equal(t, "STATUS_CREATED", got.Conditions[1].Status)
		}
	})
}
