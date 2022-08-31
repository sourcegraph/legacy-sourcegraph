package enforcement

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestNewBeforeCreateExternalServiceHook(t *testing.T) {
	tests := []struct {
		name                 string
		license              *license.Info
		externalServiceCount int
		externalService      *types.ExternalService
		wantErr              bool
	}{
		{
			name:    "Free plan",
			license: nil,
			wantErr: false,
		},

		{
			name:                 "team-0 exceeded limit",
			license:              &license.Info{Tags: []string{"plan:team-0"}},
			externalServiceCount: 1,
			externalService:      nil,
			wantErr:              true,
		},
		{
			name:                 "team-0 within limit",
			license:              &license.Info{Tags: []string{"plan:team-0"}},
			externalServiceCount: 0,
			externalService:      nil,
			wantErr:              false,
		},

		{
			name:    "business-0 with self-hosted GitHub",
			license: &license.Info{Tags: []string{"plan:business-0"}},
			externalService: &types.ExternalService{
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://github.mycompany.com/"}`),
			},
			wantErr: true,
		},
		{
			name:    "business-0 with self-hosted GitLab",
			license: &license.Info{Tags: []string{"plan:business-0"}},
			externalService: &types.ExternalService{
				Kind:   extsvc.KindGitLab,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://gitlab.mycompany.com/"}`),
			},
			wantErr: true,
		},
		{
			name:    "business-0 with GitHub.com",
			license: &license.Info{Tags: []string{"plan:business-0"}},
			externalService: &types.ExternalService{
				Kind:   extsvc.KindGitHub,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://github.com"}`),
			},
			wantErr: false,
		},
		{
			name:    "business-0 with GitLab.com",
			license: &license.Info{Tags: []string{"plan:business-0"}},
			externalService: &types.ExternalService{
				Kind:   extsvc.KindGitLab,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://gitlab.com"}`),
			},
			wantErr: false,
		},
		{
			name:    "business-0 with Bitbucket.org",
			license: &license.Info{Tags: []string{"plan:business-0"}},
			externalService: &types.ExternalService{
				Kind:   extsvc.KindBitbucketCloud,
				Config: extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.org"}`),
			},
			wantErr: false,
		},

		{
			name:    "old-starter-0 should have no limit",
			license: &license.Info{Tags: []string{"plan:old-starter-0"}},
			wantErr: false,
		},
		{
			name:    "old-enterprise-0 should have no limit",
			license: &license.Info{Tags: []string{"plan:old-enterprise-0"}},
			wantErr: false,
		},
		{
			name:    "enterprise-0 should have no limit",
			license: &license.Info{Tags: []string{"plan:enterprise-0"}},
			wantErr: false,
		},
		{
			name:    "enterprise-1 should have no limit",
			license: &license.Info{Tags: []string{"plan:enterprise-1"}},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			externalServices := database.NewMockExternalServiceStore()
			externalServices.CountFunc.SetDefaultReturn(test.externalServiceCount, nil)
			got := NewBeforeCreateExternalServiceHook()(context.Background(), externalServices, test.externalService)
			assert.Equal(t, test.wantErr, got != nil)
		})
	}
}
