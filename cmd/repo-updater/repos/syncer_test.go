package repos_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gitchander/permutation"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/awscodecommit"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/schema"
)

func testSyncerSyncWithErrors(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		githubService := types.ExternalService{
			Kind:   extsvc.KindGitHub,
			Config: `{}`,
		}

		if err := store.ExternalServiceStore().Upsert(ctx, &githubService); err != nil {
			t.Fatal(err)
		}

		for _, tc := range []struct {
			name       string
			sourcer    repos.Sourcer
			store      repos.RepoStore
			repoLister repos.RepoLister
			err        string
		}{
			{
				name:       "sourcer error aborts sync",
				sourcer:    repos.NewFakeSourcer(errors.New("boom")),
				store:      store,
				repoLister: store.RepoStore(),
				err:        "syncer.sync.sourced: 1 error occurred:\n\t* boom\n\n",
			},
			{
				name:    "store list error aborts sync",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(&githubService, nil)),
				store:   store,
				repoLister: &repoListerWithErrors{
					RepoStore: store.RepoStore(),
					ListErr:   errors.New("boom"),
				},
				err: "syncer.sync.store.list-repos: boom",
			},
			{
				name:    "store upsert error aborts sync",
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(&githubService, nil)),
				store: &storeWithErrors{
					Store:          store,
					UpsertReposErr: errors.New("booya"),
				},
				repoLister: store.RepoStore(),
				err:        "syncer.sync.store.upsert-repos: booya",
			},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				clock := dbtesting.NewFakeClock(time.Now(), time.Second)
				now := clock.Now
				ctx := context.Background()

				syncer := &repos.Syncer{
					Sourcer: tc.sourcer,
					Now:     now,
				}
				err := syncer.SyncExternalService(ctx, tc.store, tc.repoLister, store.ExternalServiceStore(), githubService.ID, time.Millisecond)

				if have, want := err.Error(), tc.err; have != want {
					t.Errorf("have error %q, want %q", have, want)
				}

				if len(syncer.SyncErrors()) != 1 {
					t.Fatal("expected 1 error")
				}

				if have, want := syncer.SyncErrors(), tc.err; have[0].Error() != want {
					t.Errorf("have SyncErrors %q, want %q", have, want)
				}
			})
		}
	}
}

type storeWithErrors struct {
	*repos.Store

	UpsertReposErr error
}

func (s *storeWithErrors) UpsertRepos(ctx context.Context, repos ...*types.Repo) error {
	if s.UpsertReposErr != nil {
		return s.UpsertReposErr
	}
	return s.Store.UpsertRepos(ctx, repos...)
}

type repoListerWithErrors struct {
	*db.RepoStore

	ListErr error
}

func (s *repoListerWithErrors) List(ctx context.Context, args db.ReposListOptions) ([]*types.Repo, error) {
	if s.ListErr != nil {
		return nil, s.ListErr
	}
	return s.RepoStore.List(ctx, args)
}

func testSyncerSync(t *testing.T, s *repos.Store) func(*testing.T) {
	services := types.MakeExternalServices()
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	for _, svc := range services {
		if err := s.ExternalServiceStore().Create(context.Background(), confGet, svc); err != nil {
			t.Fatal(err)
		}
	}

	servicesPerKind := types.ExternalServicesToMap(services)

	githubService := servicesPerKind[extsvc.KindGitHub]

	githubRepo := types.MakeGithubRepo(githubService)

	gitlabService := servicesPerKind[extsvc.KindGitLab]

	gitlabRepo := types.MakeGitlabRepo(gitlabService)

	bitbucketServerService := servicesPerKind[extsvc.KindBitbucketServer]

	bitbucketServerRepo := types.MakeBitbucketServerRepo(bitbucketServerService)

	awsCodeCommitService := servicesPerKind[extsvc.KindAWSCodeCommit]

	awsCodeCommitRepo := types.MakeAWSCodeCommitRepo(awsCodeCommitService)

	otherService := servicesPerKind[extsvc.KindOther]

	otherRepo := types.MakeOtherRepo(otherService)

	gitoliteService := servicesPerKind[extsvc.KindGitolite]

	gitoliteRepo := types.MakeGitoliteRepo(gitoliteService)

	bitbucketCloudService := servicesPerKind[extsvc.KindBitbucketCloud]

	bitbucketCloudRepo := types.MakeBitbucketCloudRepo(bitbucketCloudService)

	clock := dbtesting.NewFakeClock(time.Now(), 0)

	svcdup := types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "Github2 - Test",
		Config:      `{"url": "https://github.com"}`,
		CreatedAt:   clock.Now(),
		UpdatedAt:   clock.Now(),
	}

	// create a few external services
	if err := s.ExternalServiceStore().Upsert(context.Background(), &svcdup); err != nil {
		t.Fatalf("failed to insert external services: %v", err)
	}

	type testCase struct {
		name    string
		sourcer repos.Sourcer
		store   *repos.Store
		stored  types.Repos
		svcs    []*types.ExternalService
		ctx     context.Context
		now     func() time.Time
		diff    repos.Diff
		err     string
	}

	var testCases []testCase
	for _, tc := range []struct {
		repo *types.Repo
		svc  *types.ExternalService
	}{
		{repo: githubRepo, svc: githubService},
		{repo: gitlabRepo, svc: gitlabService},
		{repo: bitbucketServerRepo, svc: bitbucketServerService},
		{repo: awsCodeCommitRepo, svc: awsCodeCommitService},
		{repo: otherRepo, svc: otherService},
		{repo: gitoliteRepo, svc: gitoliteService},
		{repo: bitbucketCloudRepo, svc: bitbucketCloudService},
	} {
		testCases = append(testCases,
			testCase{
				name: string(tc.repo.Name + "/new repo"),
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone()),
				),
				store:  s,
				stored: types.Repos{},
				now:    clock.Now,
				diff: repos.Diff{Added: types.Repos{tc.repo.With(
					types.Opt.RepoCreatedAt(clock.Time(1)),
					types.Opt.RepoSources(tc.svc.Clone().URN()),
				)}},
				svcs: []*types.ExternalService{tc.svc},
				err:  "<nil>",
			},
			testCase{
				name: string(tc.repo.Name + "/new repo sources"),
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone()),
					repos.NewFakeSource(svcdup.Clone(), nil, tc.repo.Clone()),
				),
				store:  s,
				stored: types.Repos{tc.repo.Clone()},
				now:    clock.Now,
				diff: repos.Diff{Modified: types.Repos{tc.repo.With(
					types.Opt.RepoModifiedAt(clock.Time(1)),
					types.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)}},
				svcs: []*types.ExternalService{tc.svc},
				err:  "<nil>",
			},
			testCase{
				// It's expected that there could be multiple stored sources but only one will ever be returned
				// by the code host as it can't know about others.
				name: string(tc.repo.Name + "/source already stored"),
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone()),
				),
				store: s,
				stored: types.Repos{tc.repo.With(
					types.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{Unmodified: types.Repos{tc.repo.With(
					types.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)}},
				svcs: []*types.ExternalService{tc.svc},
				err:  "<nil>",
			},
			testCase{
				name:    string(tc.repo.Name + "/deleted ALL repo sources"),
				sourcer: repos.NewFakeSourcer(nil),
				store:   s,
				stored: types.Repos{tc.repo.With(
					types.Opt.RepoSources(tc.svc.URN(), svcdup.URN()),
				)},
				now: clock.Now,
				diff: repos.Diff{Deleted: types.Repos{tc.repo.With(
					types.Opt.RepoDeletedAt(clock.Time(1)),
				)}},
				svcs: []*types.ExternalService{tc.svc, &svcdup},
				err:  "<nil>",
			},
			testCase{
				name:    string(tc.repo.Name + "/renamed repo is detected via external_id"),
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone())),
				store:   s,
				stored: types.Repos{tc.repo.With(func(r *types.Repo) {
					r.Name = "old-name"
				})},
				now: clock.Now,
				diff: repos.Diff{Modified: types.Repos{
					tc.repo.With(
						types.Opt.RepoModifiedAt(clock.Time(1))),
				}},
				svcs: []*types.ExternalService{tc.svc},
				err:  "<nil>",
			},
			testCase{
				name: string(tc.repo.Name + "/repo got renamed to another repo that gets deleted"),
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo.With(func(r *types.Repo) { r.ExternalRepo.ID = "another-id" }),
					),
				),
				store: s,
				stored: types.Repos{
					tc.repo.Clone(),
					tc.repo.With(func(r *types.Repo) {
						r.Name = "another-repo"
						r.ExternalRepo.ID = "another-id"
					}),
				},
				now: clock.Now,
				diff: repos.Diff{
					Deleted: types.Repos{
						tc.repo.With(func(r *types.Repo) {
							r.Sources = map[string]*types.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdatedAt = clock.Time(0)
						}),
					},
					Modified: types.Repos{
						tc.repo.With(
							types.Opt.RepoModifiedAt(clock.Time(1)),
							func(r *types.Repo) { r.ExternalRepo.ID = "another-id" },
						),
					},
				},
				svcs: []*types.ExternalService{tc.svc},
				err:  "<nil>",
			},
			testCase{
				name: string(tc.repo.Name + "/repo inserted with same name as another repo that gets deleted"),
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo,
					),
				),
				store: s,
				stored: types.Repos{
					tc.repo.With(types.Opt.RepoExternalID("another-id")),
				},
				now: clock.Now,
				diff: repos.Diff{
					Added: types.Repos{
						tc.repo.With(
							types.Opt.RepoCreatedAt(clock.Time(1)),
							types.Opt.RepoModifiedAt(clock.Time(1)),
						),
					},
					Deleted: types.Repos{
						tc.repo.With(func(r *types.Repo) {
							r.ExternalRepo.ID = "another-id"
							r.Sources = map[string]*types.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdatedAt = clock.Time(0)
						}),
					},
				},
				svcs: []*types.ExternalService{tc.svc},
				err:  "<nil>",
			},
			testCase{
				name: string(tc.repo.Name + "/repo inserted with same name as repo without id"),
				sourcer: repos.NewFakeSourcer(nil,
					repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo,
					),
				),
				store: s,
				stored: types.Repos{
					tc.repo.With(types.Opt.RepoName("old-name")),  // same external id as sourced
					tc.repo.With(types.Opt.RepoExternalID("bar")), // same name as sourced
				}.With(types.Opt.RepoCreatedAt(clock.Time(1))),
				now: clock.Now,
				diff: repos.Diff{
					Modified: types.Repos{
						tc.repo.With(
							types.Opt.RepoCreatedAt(clock.Time(1)),
							types.Opt.RepoModifiedAt(clock.Time(1)),
						),
					},
					Deleted: types.Repos{
						tc.repo.With(func(r *types.Repo) {
							r.ExternalRepo.ID = ""
							r.Sources = map[string]*types.SourceInfo{}
							r.DeletedAt = clock.Time(0)
							r.UpdatedAt = clock.Time(0)
							r.CreatedAt = clock.Time(0)
						}),
					},
				},
				svcs: []*types.ExternalService{tc.svc},
				err:  "<nil>",
			},
			testCase{
				name:    string(tc.repo.Name + "/renamed repo which was deleted is detected and added"),
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil, tc.repo.Clone())),
				store:   s,
				stored: types.Repos{tc.repo.With(func(r *types.Repo) {
					r.Sources = map[string]*types.SourceInfo{}
					r.Name = "old-name"
					r.DeletedAt = clock.Time(0)
				})},
				now: clock.Now,
				diff: repos.Diff{Added: types.Repos{
					tc.repo.With(
						types.Opt.RepoCreatedAt(clock.Time(1))),
				}},
				svcs: []*types.ExternalService{tc.svc},
				err:  "<nil>",
			},
			testCase{
				name: string(tc.repo.Name + "/repos have their names swapped"),
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil,
					tc.repo.With(func(r *types.Repo) {
						r.Name = "foo"
						r.ExternalRepo.ID = "1"
					}),
					tc.repo.With(func(r *types.Repo) {
						r.Name = "bar"
						r.ExternalRepo.ID = "2"
					}),
				)),
				now:   clock.Now,
				store: s,
				stored: types.Repos{
					tc.repo.With(func(r *types.Repo) {
						r.Name = "bar"
						r.ExternalRepo.ID = "1"
					}),
					tc.repo.With(func(r *types.Repo) {
						r.Name = "foo"
						r.ExternalRepo.ID = "2"
					}),
				},
				diff: repos.Diff{
					Modified: types.Repos{
						tc.repo.With(func(r *types.Repo) {
							r.Name = "foo"
							r.ExternalRepo.ID = "1"
							r.UpdatedAt = clock.Time(0)
						}),
						tc.repo.With(func(r *types.Repo) {
							r.Name = "bar"
							r.ExternalRepo.ID = "2"
							r.UpdatedAt = clock.Time(0)
						}),
					},
				},
				svcs: []*types.ExternalService{tc.svc},
				err:  "<nil>",
			},
			testCase{
				name: string(tc.repo.Name + "/case insensitive name"),
				sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil,
					tc.repo.Clone(),
					tc.repo.With(types.Opt.RepoName(api.RepoName(strings.ToUpper(string(tc.repo.Name))))),
				)),
				store:  s,
				stored: types.Repos{tc.repo.With(types.Opt.RepoName(api.RepoName(strings.ToUpper(string(tc.repo.Name)))))},
				now:    clock.Now,
				diff:   repos.Diff{Modified: types.Repos{tc.repo.With(types.Opt.RepoModifiedAt(clock.Time(0)))}},
				svcs:   []*types.ExternalService{tc.svc},
				err:    "<nil>",
			},
			func() testCase {
				var update interface{}
				typ, ok := extsvc.ParseServiceType(tc.repo.ExternalRepo.ServiceType)
				if !ok {
					panic(fmt.Sprintf("test must be extended with new external service kind: %q", strings.ToLower(tc.repo.ExternalRepo.ServiceType)))
				}
				switch typ {
				case extsvc.TypeGitHub:
					update = &github.Repository{
						IsArchived:       true,
						ViewerPermission: "ADMIN",
					}
				case extsvc.TypeGitLab:
					update = &gitlab.Project{Archived: true}
				case extsvc.TypeBitbucketServer:
					update = &bitbucketserver.Repo{Public: true}
				case extsvc.TypeBitbucketCloud:
					update = &bitbucketcloud.Repo{IsPrivate: true}
				case extsvc.TypeAWSCodeCommit:
					update = &awscodecommit.Repository{Description: "new description"}
				case extsvc.TypeOther, extsvc.TypeGitolite:
					return testCase{}
				default:
					panic(fmt.Sprintf("test must be extended with new external service kind: %q", strings.ToLower(tc.repo.ExternalRepo.ServiceType)))
				}

				expected := update
				// Special case for GitHub, see Repo.Update method
				if typ == extsvc.TypeGitHub {
					expected = &github.Repository{
						IsArchived: true,
					}
				}

				return testCase{
					name: string(tc.repo.Name + "/metadata update"),
					sourcer: repos.NewFakeSourcer(nil, repos.NewFakeSource(tc.svc.Clone(), nil,
						tc.repo.With(types.Opt.RepoModifiedAt(clock.Time(1)),
							types.Opt.RepoMetadata(update)),
					)),
					store:  s,
					stored: types.Repos{tc.repo.Clone()},
					now:    clock.Now,
					diff: repos.Diff{Modified: types.Repos{tc.repo.With(
						types.Opt.RepoModifiedAt(clock.Time(1)),
						types.Opt.RepoMetadata(expected),
					)}},
					svcs: []*types.ExternalService{tc.svc},
					err:  "<nil>",
				}
			}(),
		)
	}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range testCases {
			if tc.name == "" {
				continue
			}

			tc := tc
			ctx := context.Background()

			t.Run(tc.name, repos.Transact(ctx, tc.store, func(t testing.TB, st *repos.Store) {
				defer func() {
					if err := recover(); err != nil {
						t.Fatalf("%q panicked: %v", tc.name, err)
					}
				}()

				now := tc.now
				if now == nil {
					clock := dbtesting.NewFakeClock(time.Now(), time.Second)
					now = clock.Now
				}

				ctx := tc.ctx
				if ctx == nil {
					ctx = context.Background()
				}

				if st != nil && len(tc.stored) > 0 {
					cloned := tc.stored.Clone()
					if err := st.RepoStore().Create(ctx, cloned...); err != nil {
						t.Fatalf("failed to prepare store: %v", err)
					}
				}

				syncer := &repos.Syncer{
					Sourcer: tc.sourcer,
					Now:     now,
				}

				for _, svc := range tc.svcs {
					err := syncer.SyncExternalService(ctx, st, st.RepoStore(), st.ExternalServiceStore(), svc.ID, time.Millisecond)

					if have, want := fmt.Sprint(err), tc.err; have != want {
						t.Errorf("have error %q, want %q", have, want)
					}

					if err != nil {
						return
					}
				}

				if st != nil {
					var want, have types.Repos
					want.Concat(tc.diff.Added, tc.diff.Modified, tc.diff.Unmodified)
					have, _ = st.RepoStore().List(ctx, db.ReposListOptions{})

					want = want.With(types.Opt.RepoID(0))
					have = have.With(types.Opt.RepoID(0))
					sort.Sort(want)
					sort.Sort(have)

					types.Assert.ReposEqual(want...)(t, have)
				}
			}))
		}
	}
}

func createExternalServices(t *testing.T, s *repos.Store) map[string]*types.ExternalService {
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}

	services := types.MakeExternalServices()
	for _, svc := range services {
		if err := s.ExternalServiceStore().Create(context.Background(), confGet, svc); err != nil {
			t.Fatal(err)
		}
	}

	return types.ExternalServicesToMap(services)
}

func testSyncRepo(t *testing.T, s *repos.Store) func(*testing.T) {
	clock := dbtesting.NewFakeClock(time.Now(), time.Second)

	servicesPerKind := createExternalServices(t, s)

	repo := types.MakeGithubRepo(servicesPerKind[extsvc.KindGitHub])

	testCases := []struct {
		name    string
		sourced *types.Repo
		stored  types.Repos
		assert  types.ReposAssertion
	}{{
		name:    "insert",
		sourced: repo,
		assert:  types.Assert.ReposEqual(repo.With(types.Opt.RepoCreatedAt(clock.Time(2)))),
	}, {
		name:    "update",
		sourced: repo,
		stored:  types.Repos{repo.With(types.Opt.RepoCreatedAt(clock.Time(2)))},
		assert: types.Assert.ReposEqual(repo.With(
			types.Opt.RepoModifiedAt(clock.Time(2)),
			types.Opt.RepoCreatedAt(clock.Time(2)))),
	}, {
		name:    "update name",
		sourced: repo,
		stored: types.Repos{repo.With(
			types.Opt.RepoName("old/name"),
			types.Opt.RepoCreatedAt(clock.Time(2)))},
		assert: types.Assert.ReposEqual(repo.With(
			types.Opt.RepoModifiedAt(clock.Time(2)),
			types.Opt.RepoCreatedAt(clock.Time(2)))),
	}, {
		name:    "delete conflicting name",
		sourced: repo,
		stored: types.Repos{repo.With(
			types.Opt.RepoExternalID("old id"),
			types.Opt.RepoCreatedAt(clock.Time(2)))},
		assert: types.Assert.ReposEqual(repo.With(
			types.Opt.RepoCreatedAt(clock.Time(2)))),
	}, {
		name:    "rename and delete conflicting name",
		sourced: repo,
		stored: types.Repos{
			repo.With(
				types.Opt.RepoExternalID("old id"),
				types.Opt.RepoCreatedAt(clock.Time(2))),
			repo.With(
				types.Opt.RepoName("old name"),
				types.Opt.RepoCreatedAt(clock.Time(2))),
		},
		assert: types.Assert.ReposEqual(repo.With(
			types.Opt.RepoCreatedAt(clock.Time(2)))),
	}}

	return func(t *testing.T) {
		t.Helper()

		for _, tc := range testCases {
			if tc.name == "" {
				continue
			}

			tc := tc
			ctx := context.Background()

			t.Run(tc.name, repos.Transact(ctx, s, func(t testing.TB, st *repos.Store) {
				defer func() {
					if err := recover(); err != nil {
						t.Fatalf("%q panicked: %v", tc.name, err)
					}
				}()

				if len(tc.stored) > 0 {
					if err := st.RepoStore().Create(ctx, tc.stored.Clone()...); err != nil {
						t.Fatalf("failed to prepare store: %v", err)
					}
				}

				clock := clock
				syncer := &repos.Syncer{
					Now: clock.Now,
				}
				err := syncer.SyncRepo(ctx, st, tc.sourced.Clone())
				if err != nil {
					t.Fatal(err)
				}

				have, err := st.RepoStore().List(ctx, db.ReposListOptions{})
				if err != nil {
					t.Fatal(err)
				}

				tc.assert(t, have)
			}))
		}
	}
}

func TestDiff(t *testing.T) {
	t.Parallel()

	eid := func(id string) api.ExternalRepoSpec {
		return api.ExternalRepoSpec{
			ID:          id,
			ServiceType: "fake",
			ServiceID:   "https://fake.com",
		}
	}
	now := time.Now()

	type testCase struct {
		name   string
		store  types.Repos
		source types.Repos
		diff   repos.Diff
	}

	cases := []testCase{
		{
			name: "empty",
			diff: repos.Diff{},
		},
		{
			name:   "added",
			source: types.Repos{{ExternalRepo: eid("1")}},
			diff:   repos.Diff{Added: types.Repos{{ExternalRepo: eid("1")}}},
		},
		{
			name:  "deleted",
			store: types.Repos{{ExternalRepo: eid("1")}},
			diff:  repos.Diff{Deleted: types.Repos{{ExternalRepo: eid("1")}}},
		},
		{
			name: "modified",
			store: types.Repos{
				{ExternalRepo: eid("1"), Name: "foo", RepoFields: &types.RepoFields{Description: "foo"}},
				{ExternalRepo: eid("2"), Name: "bar"},
			},
			source: types.Repos{
				{ExternalRepo: eid("1"), Name: "foo", RepoFields: &types.RepoFields{Description: "bar"}},
				{ExternalRepo: eid("2"), Name: "bar", RepoFields: &types.RepoFields{URI: "2"}},
			},
			diff: repos.Diff{Modified: types.Repos{
				{ExternalRepo: eid("1"), Name: "foo", RepoFields: &types.RepoFields{Description: "bar"}},
				{ExternalRepo: eid("2"), Name: "bar", RepoFields: &types.RepoFields{URI: "2"}},
			}},
		},
		{
			name:   "unmodified",
			store:  types.Repos{{ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}}},
			source: types.Repos{{ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}}},
			diff: repos.Diff{Unmodified: types.Repos{
				{ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}},
			}},
		},
		{
			name: "duplicates in source are merged",
			source: types.Repos{
				{ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo", Sources: map[string]*types.SourceInfo{
					"a": {ID: "a"},
				}}},
				{ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "bar", Sources: map[string]*types.SourceInfo{
					"b": {ID: "b"},
				}}},
			},
			diff: repos.Diff{Added: types.Repos{
				{ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "bar", Sources: map[string]*types.SourceInfo{
					"a": {ID: "a"},
					"b": {ID: "b"},
				}}},
			}},
		},
		{
			name: "duplicate with a changed name is merged correctly",
			store: types.Repos{
				{Name: "1", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}},
			},
			source: types.Repos{
				{Name: "2", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}},
			},
			diff: repos.Diff{Modified: types.Repos{
				{Name: "2", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}},
			}},
		},
		{
			name: "unmodified preserves stored repo",
			store: types.Repos{
				{ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo", UpdatedAt: now}},
			},
			source: types.Repos{
				{ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}},
			},
			diff: repos.Diff{Unmodified: types.Repos{
				{ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo", UpdatedAt: now}},
			}},
		},
	}

	permutedCases := []testCase{
		{
			// Repo renamed and repo created with old name
			name: "renamed repo",
			store: types.Repos{
				{Name: "1", ExternalRepo: eid("old"), RepoFields: &types.RepoFields{Description: "foo"}},
			},
			source: types.Repos{
				{Name: "2", ExternalRepo: eid("old"), RepoFields: &types.RepoFields{Description: "foo"}},
				{Name: "1", ExternalRepo: eid("new"), RepoFields: &types.RepoFields{Description: "bar"}},
			},
			diff: repos.Diff{
				Modified: types.Repos{
					{Name: "2", ExternalRepo: eid("old"), RepoFields: &types.RepoFields{Description: "foo"}},
				},
				Added: types.Repos{
					{Name: "1", ExternalRepo: eid("new"), RepoFields: &types.RepoFields{Description: "bar"}},
				},
			},
		},
		{
			name:  "repo renamed to an already deleted repo",
			store: types.Repos{},
			source: types.Repos{
				{Name: "a", ExternalRepo: eid("b")},
			},
			diff: repos.Diff{
				Added: types.Repos{
					{Name: "a", ExternalRepo: eid("b")},
				},
			},
		},
		{
			name: "repo renamed to a repo that gets deleted",
			store: types.Repos{
				{Name: "a", ExternalRepo: eid("a")},
				{Name: "b", ExternalRepo: eid("b")},
			},
			source: types.Repos{
				{Name: "a", ExternalRepo: eid("b")},
			},
			diff: repos.Diff{
				Deleted:  types.Repos{{Name: "a", ExternalRepo: eid("a")}},
				Modified: types.Repos{{Name: "a", ExternalRepo: eid("b")}},
			},
		},
		{
			name: "swapped repo",
			store: types.Repos{
				{Name: "foo", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}},
				{Name: "bar", ExternalRepo: eid("2"), RepoFields: &types.RepoFields{Description: "bar"}},
			},
			source: types.Repos{
				{Name: "bar", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "bar"}},
				{Name: "foo", ExternalRepo: eid("2"), RepoFields: &types.RepoFields{Description: "foo"}},
			},
			diff: repos.Diff{
				Modified: types.Repos{
					{Name: "bar", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "bar"}},
					{Name: "foo", ExternalRepo: eid("2"), RepoFields: &types.RepoFields{Description: "foo"}},
				},
			},
		},
		{
			name: "deterministic merging of source",
			source: types.Repos{
				{Name: "foo", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "desc1", Sources: map[string]*types.SourceInfo{"a": nil}}},
				{Name: "foo", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "desc2", Sources: map[string]*types.SourceInfo{"b": nil}}},
			},
			diff: repos.Diff{
				Added: types.Repos{
					{Name: "foo", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "desc2", Sources: map[string]*types.SourceInfo{"a": nil, "b": nil}}},
				},
			},
		},
		{
			name: "conflict on case insensitive name",
			source: types.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
				{Name: "Foo", ExternalRepo: eid("2")},
			},
			diff: repos.Diff{
				Added: types.Repos{
					{Name: "Foo", ExternalRepo: eid("2")},
				},
			},
		},
		{
			name: "conflict on case insensitive name exists 1",
			store: types.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
			},
			source: types.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
				{Name: "Foo", ExternalRepo: eid("2")},
			},
			diff: repos.Diff{
				Added: types.Repos{
					{Name: "Foo", ExternalRepo: eid("2")},
				},
				Deleted: types.Repos{
					{Name: "foo", ExternalRepo: eid("1")},
				},
			},
		},
		{
			name: "conflict on case insensitive name exists 2",
			store: types.Repos{
				{Name: "Foo", ExternalRepo: eid("2")},
			},
			source: types.Repos{
				{Name: "foo", ExternalRepo: eid("1")},
				{Name: "Foo", ExternalRepo: eid("2")},
			},
			diff: repos.Diff{
				Unmodified: types.Repos{
					{Name: "Foo", ExternalRepo: eid("2")},
				},
			},
		},
		// 🚨 SECURITY: Tests to ensure we detect repository visibility changes.
		{
			name: "repository visiblity changed",
			store: types.Repos{
				{Name: "foo", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}, Private: false},
			},
			source: types.Repos{
				{Name: "foo", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}, Private: true},
			},
			diff: repos.Diff{
				Modified: types.Repos{
					{Name: "foo", ExternalRepo: eid("1"), RepoFields: &types.RepoFields{Description: "foo"}, Private: true},
				},
			},
		},
	}

	for _, tc := range permutedCases {
		store := permutation.New(tc.store)
		for i := 0; store.Next(); i++ {
			source := permutation.New(tc.source)
			for j := 0; source.Next(); j++ {
				cases = append(cases, testCase{
					name:   fmt.Sprintf("%s/permutation_%d_%d", tc.name, i, j),
					store:  tc.store.Clone(),
					source: tc.source.Clone(),
					diff:   tc.diff,
				})
			}
		}
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			diff := repos.NewDiff(tc.source, tc.store)
			diff.Sort()
			tc.diff.Sort()
			if cDiff := cmp.Diff(diff, tc.diff); cDiff != "" {
				t.Fatalf("unexpected diff:\n%s", cDiff)
			}
		})
	}
}

func testSyncRun(db *sql.DB) func(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T, store *repos.Store) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			svc := &types.ExternalService{
				Config: `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
				Kind:   extsvc.KindGitHub,
			}

			if err := store.ExternalServiceStore().Upsert(ctx, svc); err != nil {
				t.Fatal(err)
			}

			mk := func(name string) *types.Repo {
				return &types.Repo{
					Name: api.RepoName(name),
					ExternalRepo: api.ExternalRepoSpec{
						ID:          name,
						ServiceID:   "https://github.com",
						ServiceType: svc.Kind,
					},
					RepoFields: &types.RepoFields{
						Metadata: &github.Repository{},
					},
				}
			}

			// Our test will have 1 initial repo, and discover a new repo on sourcing.
			stored := types.Repos{mk("initial")}.With(types.Opt.RepoSources(svc.URN()))
			sourced := types.Repos{mk("initial"), mk("new")}

			syncer := &repos.Syncer{
				Sourcer:      repos.NewFakeSourcer(nil, repos.NewFakeSource(svc, nil, sourced...)),
				Store:        store,
				Synced:       make(chan repos.Diff),
				SubsetSynced: make(chan repos.Diff),
				Now:          time.Now,
			}

			// Initial repos in store
			if err := store.RepoStore().Create(ctx, stored...); err != nil {
				t.Fatal(err)
			}

			done := make(chan error)
			go func() {
				done <- syncer.Run(ctx, db, store, repos.RunOptions{
					EnqueueInterval: func() time.Duration { return time.Second },
					IsCloud:         false,
					MinSyncInterval: func() time.Duration { return 1 * time.Millisecond },
					DequeueInterval: 1 * time.Millisecond,
				})
			}()

			// Ignore fields store adds
			ignore := cmpopts.IgnoreFields(types.Repo{}, "ID", "CreatedAt", "UpdatedAt", "Sources")

			// The first thing sent down Synced is the list of repos in store.
			diff := <-syncer.Synced
			if d := cmp.Diff(repos.Diff{Unmodified: stored}, diff, ignore); d != "" {
				t.Fatalf("initial Synced mismatch (-want +got):\n%s", d)
			}

			// Next up it should find the new repo and send it down SubsetSynced
			diff = <-syncer.SubsetSynced
			if d := cmp.Diff(repos.Diff{Added: types.Repos{mk("new")}}, diff, ignore); d != "" {
				t.Fatalf("SubsetSynced mismatch (-want +got):\n%s", d)
			}

			// Finally we get the final diff, which will have everything listed as
			// Unmodified since we added when we did SubsetSynced.
			diff = <-syncer.Synced
			if d := cmp.Diff(repos.Diff{Unmodified: sourced}, diff, ignore); d != "" {
				t.Fatalf("final Synced mismatch (-want +got):\n%s", d)
			}

			// We check synced again to test us going around the Run loop 2 times in
			// total.
			diff = <-syncer.Synced
			if d := cmp.Diff(repos.Diff{Unmodified: sourced}, diff, ignore); d != "" {
				t.Fatalf("second final Synced mismatch (-want +got):\n%s", d)
			}

			// Cancel context and the run loop should stop
			cancel()
			err := <-done
			if err != nil && err != context.Canceled {
				t.Fatal(err)
			}
		}
	}
}

func testSyncer(db *sql.DB) func(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T, store *repos.Store) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			services := types.MakeExternalServices()

			githubService := services[0]
			gitlabService := services[1]
			bitbucketCloudService := services[3]

			services = types.ExternalServices{
				githubService,
				gitlabService,
				bitbucketCloudService,
			}

			// setup services
			if err := store.ExternalServiceStore().Upsert(ctx, services...); err != nil {
				t.Fatal(err)
			}

			githubRepo := types.MakeGithubRepo(githubService)

			gitlabRepo := types.MakeGitlabRepo(gitlabService)

			bitbucketCloudRepo := types.MakeBitbucketCloudRepo(bitbucketCloudService)

			removeSources := func(r *types.Repo) {
				r.Sources = nil
			}

			baseGithubRepos := types.GenerateRepos(10, githubRepo)
			githubSourced := baseGithubRepos.Clone().With(removeSources)
			baseGitlabRepos := types.GenerateRepos(10, gitlabRepo)
			gitlabSourced := baseGitlabRepos.Clone().With(removeSources)
			baseBitbucketCloudRepos := types.GenerateRepos(10, bitbucketCloudRepo)
			bitbucketCloudSourced := baseBitbucketCloudRepos.Clone().With(removeSources)

			sourcers := map[int64]repos.Source{
				githubService.ID:         repos.NewFakeSource(githubService, nil, githubSourced...),
				gitlabService.ID:         repos.NewFakeSource(gitlabService, nil, gitlabSourced...),
				bitbucketCloudService.ID: repos.NewFakeSource(bitbucketCloudService, nil, bitbucketCloudSourced...),
			}

			syncer := &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					if len(services) > 1 {
						t.Fatalf("Expected 1 service, got %d", len(services))
					}
					s, ok := sourcers[services[0].ID]
					if !ok {
						t.Fatalf("sourcer not found: %d", services[0].ID)
					}
					return repos.Sources{s}, nil
				},
				Store:  store,
				Synced: make(chan repos.Diff),
				Now:    time.Now,
			}

			done := make(chan error)
			go func() {
				done <- syncer.Run(ctx, db, store, repos.RunOptions{
					EnqueueInterval: func() time.Duration { return time.Second },
					IsCloud:         false,
					MinSyncInterval: func() time.Duration { return 1 * time.Minute },
					DequeueInterval: 1 * time.Millisecond,
				})
			}()

			// Ignore fields store adds
			ignore := cmpopts.IgnoreFields(types.Repo{}, "ID", "CreatedAt", "UpdatedAt", "Sources")

			// The first thing sent down Synced is an empty list of repos in store.
			diff := <-syncer.Synced
			if d := cmp.Diff(repos.Diff{}, diff, ignore); d != "" {
				t.Fatalf("initial Synced mismatch (-want +got):\n%s", d)
			}

			// it should add a job for all external services
			var jobCount int
			for i := 0; i < 10; i++ {
				if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM external_service_sync_jobs").Scan(&jobCount); err != nil {
					t.Fatal(err)
				}
				if jobCount == len(services) {
					break
				}
				// We need to give the worker package time to create the jobs
				time.Sleep(10 * time.Millisecond)
			}
			if jobCount != len(services) {
				t.Fatalf("expected %d sync jobs, got %d", len(services), jobCount)
			}

			for i := 0; i < len(services); i++ {
				diff = <-syncer.Synced
				if len(diff.Added) != 10 {
					t.Fatalf("Expected 10 Added repos. got %d", len(diff.Added))
				}
				if len(diff.Deleted) != 0 {
					t.Fatalf("Expected 0 Deleted repos. got %d", len(diff.Added))
				}
				if len(diff.Modified) != 0 {
					t.Fatalf("Expected 0 Modified repos. got %d", len(diff.Added))
				}
				if len(diff.Unmodified) != 0 {
					t.Fatalf("Expected 0 Unmodified repos. got %d", len(diff.Added))
				}
			}

			var jobsCompleted int
			for i := 0; i < 10; i++ {
				if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM external_service_sync_jobs where state = 'completed'").Scan(&jobsCompleted); err != nil {
					t.Fatal(err)
				}
				if jobsCompleted == len(services) {
					break
				}
				// We need to give the worker package time to create the jobs
				time.Sleep(10 * time.Millisecond)
			}

			// Cancel context and the run loop should stop
			cancel()
			err := <-done
			if err != nil && err != context.Canceled {
				t.Fatal(err)
			}
		}
	}
}

func testOrphanedRepo(sqlDB *sql.DB) func(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T, store *repos.Store) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			now := time.Now()

			svc1 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test1",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			svc2 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test2",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// setup services
			if err := store.ExternalServiceStore().Upsert(ctx, svc1, svc2); err != nil {
				t.Fatal(err)
			}

			githubRepo := types.MakeGithubRepo()

			// Add two services, both pointing at the same repo

			// Sync first service
			syncer := &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc1, nil, githubRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc1.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Sync second service
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc2, nil, githubRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc2.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Confirm that there are two relationships
			assertSourceCount(ctx, t, sqlDB, 2)

			// We should have no deleted repos
			assertDeletedRepoCount(ctx, t, sqlDB, 0)

			// Remove the repo from one service and sync again
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc1, nil)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc1.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Confirm that the repository hasn't been deleted
			rs, err := store.RepoStore().List(ctx, db.ReposListOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if len(rs) != 1 {
				t.Fatalf("Expected 1 repo, got %d", len(rs))
			}

			// Confirm that there is one relationship
			assertSourceCount(ctx, t, sqlDB, 1)

			// We should have no deleted repos
			assertDeletedRepoCount(ctx, t, sqlDB, 0)

			// Remove the repo from the second service and sync again
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc2, nil)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc2.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Confirm that there no relationships
			assertSourceCount(ctx, t, sqlDB, 0)

			// We should have one deleted repo
			assertDeletedRepoCount(ctx, t, sqlDB, 1)
		}
	}
}

// storeWrapper executes arbitrary code before store methods.
type storeWrapper struct {
	*repos.Store

	onUpsertRepos func()
}

func (s *storeWrapper) UpsertRepos(ctx context.Context, rs ...*types.Repo) error {
	if s.onUpsertRepos != nil {
		s.onUpsertRepos()
	}

	return s.Store.UpsertRepos(ctx, rs...)
}

func testConflictingSyncers(sqlDB *sql.DB) func(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T, store *repos.Store) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			now := time.Now()

			svc1 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test1",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			svc2 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test2",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// setup services
			if err := store.ExternalServiceStore().Upsert(ctx, svc1, svc2); err != nil {
				t.Fatal(err)
			}

			githubRepo := types.MakeGithubRepo()
			githubRepo.Description = ""

			// Add two services, both pointing at the same repo

			// Sync first service
			syncer := &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc1, nil, githubRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc1.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Sync second service
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc2, nil, githubRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc2.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Confirm that there are two relationships
			assertSourceCount(ctx, t, sqlDB, 2)

			fromDB, err := store.RepoStore().List(ctx, db.ReposListOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if len(fromDB) != 1 {
				t.Fatalf("Expected 1 repo, got %d", len(fromDB))
			}
			beforeUpdate := fromDB[0]
			if beforeUpdate.Description != "" {
				t.Fatalf("Expected %q, got %q", "", beforeUpdate.Description)
			}

			// Create two transactions
			tx1, err := store.Transact(ctx)
			if err != nil {
				t.Fatal(err)
			}

			tx2, err := store.Transact(ctx)
			if err != nil {
				t.Fatal(err)
			}

			newDescription := "This has changed"
			updatedRepo := githubRepo.With(func(r *types.Repo) {
				r.Description = newDescription
			})

			// Start syncing using tx1
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc1, nil, updatedRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, tx1, tx1.RepoStore(), tx1.ExternalServiceStore(), svc1.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			errChan := make(chan error)
			upsertCalledCh := make(chan struct{})
			go func() {
				// Start syncing using tx2
				syncer2 := &repos.Syncer{
					Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
						s := repos.NewFakeSource(svc2, nil, updatedRepo.With(func(r *types.Repo) {
							r.Description = newDescription
						}))
						return repos.Sources{s}, nil
					},
					Now: time.Now,
				}
				err := syncer2.SyncExternalService(ctx, &storeWrapper{Store: tx2, onUpsertRepos: func() {
					close(upsertCalledCh)
				}}, tx2.RepoStore(), tx2.ExternalServiceStore(), svc2.ID, 10*time.Second)
				errChan <- err
			}()

			<-upsertCalledCh
			tx1.Done(nil)

			err = <-errChan
			if err != nil {
				t.Fatalf("Error in syncer2: %v", err)
			}
			tx2.Done(nil)

			fromDB, err = store.RepoStore().List(ctx, db.ReposListOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if len(fromDB) != 1 {
				t.Fatalf("Expected 1 repo, got %d", len(fromDB))
			}
			afterUpdate := fromDB[0]
			if afterUpdate.Description != newDescription {
				t.Fatalf("Expected %q, got %q", newDescription, afterUpdate.Description)
			}
		}
	}
}

func testUserAddedRepos(sqlDB *sql.DB, userID int32) func(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T, store *repos.Store) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			now := time.Now()

			userService := &types.ExternalService{
				Kind:            extsvc.KindGitHub,
				DisplayName:     "Github - User",
				Config:          `{"url": "https://github.com"}`,
				CreatedAt:       now,
				UpdatedAt:       now,
				NamespaceUserID: userID,
			}

			adminService := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Private",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// setup services
			if err := store.ExternalServiceStore().Upsert(ctx, userService, adminService); err != nil {
				t.Fatal(err)
			}

			publicRepo := &types.Repo{
				Name: "github.com/org/user",
				RepoFields: &types.RepoFields{
					Metadata: &github.Repository{},
				},
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "foo-external-user",
					ServiceID:   "https://github.com/",
					ServiceType: extsvc.TypeGitHub,
				},
			}

			publicRepo2 := &types.Repo{
				Name: "github.com/org/user2",
				RepoFields: &types.RepoFields{
					Metadata: &github.Repository{},
				},
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "foo-external-user2",
					ServiceID:   "https://github.com/",
					ServiceType: extsvc.TypeGitHub,
				},
			}

			privateRepo := &types.Repo{
				Name: "github.com/org/private",
				RepoFields: &types.RepoFields{
					Metadata: &github.Repository{},
				},
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "foo-external-private",
					ServiceID:   "https://github.com/",
					ServiceType: extsvc.TypeGitHub,
				},
				Private: true,
			}

			// Admin service will sync both repos
			syncer := &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(adminService, nil, publicRepo, privateRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), adminService.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Confirm that there are two relationships
			assertSourceCount(ctx, t, sqlDB, 2)

			// Unsync the repo to clean things up
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(adminService, nil)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), adminService.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Confirm that there are zero relationships
			assertSourceCount(ctx, t, sqlDB, 0)

			// By default, user service can only sync public code, even if they have access to private code
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(userService, nil, publicRepo, privateRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), userService.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Confirm that there is one relationship
			assertSourceCount(ctx, t, sqlDB, 1)

			// If the private code feature flag is set, user service can also sync private code
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExternalServiceUserMode: "all",
				},
			})

			syncer = &repos.Syncer{
				Sourcer: func(services ...*repos.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(userService, nil, publicRepo, privateRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, userService.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Confirm that there are two relationships
			assertSourceCount(ctx, t, db, 2)
			conf.Mock(nil)

			// Attempt to add some repos with a per user limit set
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(userService, nil, publicRepo, publicRepo2)
					return repos.Sources{s}, nil
				},
				Now:                 time.Now,
				UserReposMaxPerUser: 1,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), userService.ID, 10*time.Second); err == nil {
				t.Fatal("Expected an error, got none")
			}

			// Attempt to add some repos with a total limit set
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(userService, nil, publicRepo, publicRepo2)
					return repos.Sources{s}, nil
				},
				Now:                 time.Now,
				UserReposMaxPerSite: 1,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), userService.ID, 10*time.Second); err == nil {
				t.Fatal("Expected an error, got none")
			}
		}
	}
}

func testNameOnConflictDiscardOld(sqlDB *sql.DB) func(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T, store *repos.Store) func(t *testing.T) {
		return func(t *testing.T) {
			// Test the case where more than one external service returns the same name for different repos. The names
			// are the same, but the external id are different.

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			now := time.Now()

			svc1 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test1",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			svc2 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test2",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// setup services
			if err := store.ExternalServiceStore().Upsert(ctx, svc1, svc2); err != nil {
				t.Fatal(err)
			}

			githubRepo1 := &types.Repo{
				Name: "github.com/org/foo",
				RepoFields: &types.RepoFields{
					Metadata: &github.Repository{},
				},
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "foo-external-foo",
					ServiceID:   "https://github.com/",
					ServiceType: extsvc.TypeGitHub,
				},
			}

			githubRepo2 := &types.Repo{
				Name: "github.com/org/foo",
				RepoFields: &types.RepoFields{
					Metadata: &github.Repository{},
				},
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "foo-external-bar",
					ServiceID:   "https://github.com/",
					ServiceType: extsvc.TypeGitHub,
				},
			}

			// Add two services, one with each repo

			// Sync first service
			syncer := &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc1, nil, githubRepo1)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc1.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Sync second service
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc2, nil, githubRepo2)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc2.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// We expect repo2 to be synced since it sorts before repo1 because the ID is alphabetically first
			fromDB, err := store.RepoStore().List(ctx, db.ReposListOptions{
				Names: []string{"github.com/org/foo"},
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(fromDB) != 1 {
				t.Fatalf("Expected 1 repo, have %d", len(fromDB))
			}

			found := fromDB[0]
			expectedID := "foo-external-bar"
			if found.ExternalRepo.ID != expectedID {
				t.Fatalf("Want %q, got %q", expectedID, found.ExternalRepo.ID)
			}
		}
	}
}

func testNameOnConflictDiscardNew(sqlDB *sql.DB) func(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T, store *repos.Store) func(t *testing.T) {
		return func(t *testing.T) {
			// Test the case where more than one external service returns the same name for different repos. The names
			// are the same, but the external id are different.

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			now := time.Now()

			svc1 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test1",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			svc2 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test2",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// setup services
			if err := store.ExternalServiceStore().Upsert(ctx, svc1, svc2); err != nil {
				t.Fatal(err)
			}

			githubRepo1 := &types.Repo{
				Name: "github.com/org/foo",
				RepoFields: &types.RepoFields{
					Metadata: &github.Repository{},
				},
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "foo-external-bar",
					ServiceID:   "https://github.com/",
					ServiceType: extsvc.TypeGitHub,
				},
			}

			githubRepo2 := &types.Repo{
				Name: "github.com/org/foo",
				RepoFields: &types.RepoFields{
					Metadata: &github.Repository{},
				},
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "foo-external-foo",
					ServiceID:   "https://github.com/",
					ServiceType: extsvc.TypeGitHub,
				},
			}

			// Add two services, one with each repo

			// Sync first service
			syncer := &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc1, nil, githubRepo1)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc1.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Sync second service
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc2, nil, githubRepo2)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc2.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// We expect repo1 to be synced since it sorts before repo2 because the ID is alphabetically first
			fromDB, err := store.RepoStore().List(ctx, db.ReposListOptions{
				Names: []string{"github.com/org/foo"},
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(fromDB) != 1 {
				t.Fatalf("Expected 1 repo, have %d", len(fromDB))
			}

			found := fromDB[0]
			expectedID := "foo-external-bar"
			if found.ExternalRepo.ID != expectedID {
				t.Fatalf("Want %q, got %q", expectedID, found.ExternalRepo.ID)
			}
		}
	}
}

func testNameOnConflictOnRename(sqlDB *sql.DB) func(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T, store *repos.Store) func(t *testing.T) {
		return func(t *testing.T) {
			// Test the case where more than one external service returns the same name for different repos. The names
			// are the same, but the external id are different.

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			now := time.Now()

			svc1 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test1",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			svc2 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test2",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// setup services
			if err := store.ExternalServiceStore().Upsert(ctx, svc1, svc2); err != nil {
				t.Fatal(err)
			}

			githubRepo1 := &types.Repo{
				Name: "github.com/org/foo",
				RepoFields: &types.RepoFields{
					Metadata: &github.Repository{},
				},
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "foo-external-foo",
					ServiceID:   "https://github.com/",
					ServiceType: extsvc.TypeGitHub,
				},
			}

			githubRepo2 := &types.Repo{
				Name: "github.com/org/bar",
				RepoFields: &types.RepoFields{
					Metadata: &github.Repository{},
				},
				ExternalRepo: api.ExternalRepoSpec{
					ID:          "foo-external-bar",
					ServiceID:   "https://github.com/",
					ServiceType: extsvc.TypeGitHub,
				},
			}

			// Add two services, one with each repo

			// Sync first service
			syncer := &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc1, nil, githubRepo1)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc1.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Sync second service
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc2, nil, githubRepo2)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc2.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Rename repo1 with the same name as repo2
			renamedRepo1 := githubRepo1.With(func(r *types.Repo) {
				r.Name = githubRepo2.Name
			})

			// Sync first service
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc1, nil, renamedRepo1)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc1.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// We expect repo1 to be synced since it sorts before repo2 because the ID is alphabetically first
			fromDB, err := store.RepoStore().List(ctx, db.ReposListOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if len(fromDB) != 1 {
				t.Fatalf("Expected 1 repo, have %d", len(fromDB))
			}

			found := fromDB[0]
			expectedID := "foo-external-bar"
			if found.ExternalRepo.ID != expectedID {
				t.Fatalf("Want %q, got %q", expectedID, found.ExternalRepo.ID)
			}
		}
	}
}
func testDeleteExternalService(db *sql.DB) func(t *testing.T, store *repos.Store) func(t *testing.T) {
	return func(t *testing.T, store *repos.Store) func(t *testing.T) {
		return func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			now := time.Now()

			svc1 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test1",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			svc2 := &types.ExternalService{
				Kind:        extsvc.KindGitHub,
				DisplayName: "Github - Test2",
				Config:      `{"url": "https://github.com"}`,
				CreatedAt:   now,
				UpdatedAt:   now,
			}

			// setup services
			if err := store.ExternalServiceStore().Upsert(ctx, svc1, svc2); err != nil {
				t.Fatal(err)
			}

			githubRepo := types.MakeGithubRepo()

			// Add two services, both pointing at the same repo

			// Sync first service
			syncer := &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc1, nil, githubRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc1.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Sync second service
			syncer = &repos.Syncer{
				Sourcer: func(services ...*types.ExternalService) (repos.Sources, error) {
					s := repos.NewFakeSource(svc2, nil, githubRepo)
					return repos.Sources{s}, nil
				},
				Now: time.Now,
			}
			if err := syncer.SyncExternalService(ctx, store, store.RepoStore(), store.ExternalServiceStore(), svc2.ID, 10*time.Second); err != nil {
				t.Fatal(err)
			}

			// Delete the first service
			svc1.DeletedAt = now
			if err := store.ExternalServiceStore().Upsert(ctx, svc1); err != nil {
				t.Fatal(err)
			}

			// Confirm that there is one relationship
			assertSourceCount(ctx, t, db, 1)

			// We should have no deleted repos
			assertDeletedRepoCount(ctx, t, db, 0)

			// Delete the second service
			svc2.DeletedAt = now
			if err := store.ExternalServiceStore().Upsert(ctx, svc2); err != nil {
				t.Fatal(err)
			}

			// Confirm that there no relationships
			assertSourceCount(ctx, t, db, 0)

			// We should have one deleted repo
			assertDeletedRepoCount(ctx, t, db, 1)
		}
	}
}

func assertSourceCount(ctx context.Context, t *testing.T, db *sql.DB, want int) {
	t.Helper()
	var rowCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM external_service_repos").Scan(&rowCount); err != nil {
		t.Fatal(err)
	}
	if rowCount != want {
		t.Fatalf("Expected %d rows, got %d", want, rowCount)
	}
}

func assertDeletedRepoCount(ctx context.Context, t *testing.T, db *sql.DB, want int) {
	t.Helper()
	var rowCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM repo where deleted_at is not null").Scan(&rowCount); err != nil {
		t.Fatal(err)
	}
	if rowCount != want {
		t.Fatalf("Expected %d rows, got %d", want, rowCount)
	}
}
