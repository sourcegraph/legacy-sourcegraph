package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestUploadAuthMiddleware(t *testing.T) {
	type testCase struct {
		description        string
		isSiteAdmin        bool
		lsifEnforceAuth    bool
		authzForSiteAdmin  bool
		hasRepoAccess      bool
		expectedStatusCode int
	}

	nextHandler := func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isInternalActor := actor.FromContext(r.Context()).IsInternal()
			if !isInternalActor {
				t.Errorf("expected to be internal actor in next handler after auth middleware")
			}
		})
	}

	userStore := dbmocks.NewMockUserStore()
	repoStore := backend.NewMockReposService()
	authValidators := map[string]AuthValidator{}
	operation := observation.TestContext.Operation(observation.Op{})

	testCases := []testCase{
		{
			"conf.lsifEnforceAuth = false",
			false,
			false,
			false,
			true,
			http.StatusOK,
		},
		{
			"conf.lsifEnforceAuth = false && !hasRepoAccess",
			false,
			false,
			false,
			false,
			http.StatusNotFound,
		},
		{
			"conf.lsifEnforceAuth = true",
			false,
			true,
			false,
			true,
			http.StatusUnprocessableEntity,
		},
		{
			"isSiteAdmin = true",
			true,
			true,
			false,
			false,
			http.StatusOK,
		},
		{
			"isSiteAdmin = true && authzEnforcedForSiteAdmin",
			true,
			true,
			true,
			false,
			http.StatusNotFound,
		},
		// Codehost authValidators are not tested
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {

			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					LsifEnforceAuth:           testCase.lsifEnforceAuth,
					AuthzEnforceForSiteAdmins: testCase.authzForSiteAdmin,
				},
			})
			defer conf.Mock(nil)

			userStore.GetByCurrentAuthUserFunc.SetDefaultReturn(
				&types.User{SiteAdmin: testCase.isSiteAdmin}, nil,
			)

			if testCase.hasRepoAccess {
				repoStore.GetByNameFunc.SetDefaultReturn(nil, nil)
			} else {
				repoStore.GetByNameFunc.SetDefaultReturn(nil, &database.RepoNotFoundErr{})
			}

			handlerToTest := AuthMiddleware(
				nextHandler(),
				userStore,
				repoStore,
				authValidators,
				operation,
			)
			req := httptest.NewRequest("GET", "/upload", nil)
			rr := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rr, req)
			if status := rr.Code; status != testCase.expectedStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, testCase.expectedStatusCode)
			}
		})
	}

}
