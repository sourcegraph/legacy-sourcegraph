package database

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type SignalConfiguration struct {
	ID                   int
	Name                 string
	Description          string
	ExcludedRepoPatterns []string
	Enabled              bool
}

type SignalConfigurationStore interface {
	LoadConfigurations(ctx context.Context, args LoadSignalConfigurationArgs) ([]SignalConfiguration, error)
	UpdateConfiguration(ctx context.Context, args UpdateSignalConfigurationArgs) error
	WithTransact(context.Context, func(store SignalConfigurationStore) error) error
}

type UpdateSignalConfigurationArgs struct {
	Name                 string
	ExcludedRepoPatterns []string
	Enabled              bool
}

type signalConfigurationStore struct {
	*basestore.Store
}

func SignalConfigurationStoreWith(store basestore.ShareableStore) SignalConfigurationStore {
	return &signalConfigurationStore{Store: basestore.NewWithHandle(store.Handle())}
}

func (s *signalConfigurationStore) With(other basestore.ShareableStore) *signalConfigurationStore {
	return &signalConfigurationStore{s.Store.With(other)}
}

type LoadSignalConfigurationArgs struct {
	Name string
}

func (s *signalConfigurationStore) LoadConfigurations(ctx context.Context, args LoadSignalConfigurationArgs) ([]SignalConfiguration, error) {
	q := "SELECT id, name, description, excluded_repo_patterns, enabled FROM own_signal_configurations %s ORDER BY id;"

	var conds []*sqlf.Query
	if len(args.Name) > 0 {
		conds = append(conds, sqlf.Sprintf("name = %s", args.Name))
	}
	where := sqlf.Sprintf("")
	if len(conds) > 0 {
		where = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "AND"))
	}

	multiScan := basestore.NewSliceScanner(func(scanner dbutil.Scanner) (SignalConfiguration, error) {
		var temp SignalConfiguration
		err := scanner.Scan(
			&temp.ID,
			&temp.Name,
			&temp.Description,
			pq.Array(&temp.ExcludedRepoPatterns),
			&temp.Enabled,
		)
		if err != nil {
			return SignalConfiguration{}, err
		}
		return temp, nil
	})

	qq := sqlf.Sprintf(q, where)
	fmt.Println(qq.Query(sqlf.PostgresBindVar))
	fmt.Println(qq.Args())

	return multiScan(s.Query(ctx, sqlf.Sprintf(q, where)))
}

func (s *signalConfigurationStore) UpdateConfiguration(ctx context.Context, args UpdateSignalConfigurationArgs) error {
	q := "UPDATE own_signal_configurations SET enabled = %s, excluded_repo_patterns = %s WHERE name = %s"
	return s.Exec(ctx, sqlf.Sprintf(q, args.Enabled, pq.Array(args.ExcludedRepoPatterns), args.Name))
}

func (s *signalConfigurationStore) WithTransact(ctx context.Context, f func(store SignalConfigurationStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(s.With(tx))
	})
}
