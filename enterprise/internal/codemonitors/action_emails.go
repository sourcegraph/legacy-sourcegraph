package codemonitors

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type MonitorEmail struct {
	Id        int64
	Monitor   int64
	Enabled   bool
	Priority  string
	Header    string
	CreatedBy int32
	CreatedAt time.Time
	ChangedBy int32
	ChangedAt time.Time
}

func (s *codeMonitorStore) UpdateEmailAction(ctx context.Context, monitorID int64, action *graphqlbackend.EditActionArgs) (*MonitorEmail, error) {
	q, err := s.updateActionEmailQuery(ctx, monitorID, action.Email)
	if err != nil {
		return nil, err
	}
	row := s.QueryRow(ctx, q)
	return scanEmail(row)
}

func (s *codeMonitorStore) CreateEmailAction(ctx context.Context, monitorID int64, action *graphqlbackend.CreateActionArgs) (*MonitorEmail, error) {
	q, err := s.createActionEmailQuery(ctx, monitorID, action.Email)
	if err != nil {
		return nil, err
	}
	row := s.QueryRow(ctx, q)
	return scanEmail(row)
}

func (s *codeMonitorStore) DeleteEmailActions(ctx context.Context, actionIDs []int64, monitorID int64) error {
	if len(actionIDs) == 0 {
		return nil
	}
	q, err := deleteActionsEmailQuery(ctx, actionIDs, monitorID)
	if err != nil {
		return err
	}
	return s.Exec(ctx, q)
}

const totalCountActionEmailsFmtStr = `
SELECT COUNT(*)
FROM cm_emails
WHERE monitor = %s;
`

func (s *codeMonitorStore) TotalCountActionEmails(ctx context.Context, monitorID int64) (int32, error) {
	var count int32
	err := s.QueryRow(ctx, sqlf.Sprintf(totalCountActionEmailsFmtStr, monitorID)).Scan(&count)
	return count, err
}

const actionEmailByIDFmtStr = `
SELECT %s -- EmailsColumns
FROM cm_emails
WHERE id = %s
`

func (s *codeMonitorStore) ActionEmailByIDInt64(ctx context.Context, emailID int64) (m *MonitorEmail, err error) {
	q := sqlf.Sprintf(
		actionEmailByIDFmtStr,
		sqlf.Join(emailsColumns, ","),
		emailID,
	)
	row := s.QueryRow(ctx, q)
	return scanEmail(row)
}

const updateActionEmailFmtStr = `
UPDATE cm_emails
SET enabled = %s,
	priority = %s,
	header = %s,
	changed_by = %s,
	changed_at = %s
WHERE id = %s
AND monitor = %s
RETURNING %s;
`

func (s *codeMonitorStore) updateActionEmailQuery(ctx context.Context, monitorID int64, args *graphqlbackend.EditActionEmailArgs) (*sqlf.Query, error) {
	if args.Id == nil {
		return nil, errors.Errorf("nil is not a valid action ID")
	}

	var actionID int64
	err := relay.UnmarshalSpec(*args.Id, &actionID)
	if err != nil {
		return nil, err
	}

	now := s.Now()
	a := actor.FromContext(ctx)
	return sqlf.Sprintf(
		updateActionEmailFmtStr,
		args.Update.Enabled,
		args.Update.Priority,
		args.Update.Header,
		a.UID,
		now,
		actionID,
		monitorID,
		sqlf.Join(emailsColumns, ", "),
	), nil
}

// ListActionsOpts holds list options for listing actions
type ListActionsOpts struct {
	// MonitorID, if set, will constrain the listed actions to only
	// those that are defined as part of the given monitor.
	// References cm_monitors(id)
	MonitorID *int

	// First, if set, limits the number of actions returned
	// to the first n.
	First *int

	// After, if set, begins listing actions after the given id
	After *int
}

func (o ListActionsOpts) Conds() *sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if o.MonitorID != nil {
		conds = append(conds, sqlf.Sprintf("monitor = %s", *o.MonitorID))
	}
	if o.After != nil {
		conds = append(conds, sqlf.Sprintf("id > %s", *o.After))
	}
	return sqlf.Join(conds, "AND")
}

func (o ListActionsOpts) Limit() *sqlf.Query {
	if o.First == nil {
		return sqlf.Sprintf("ALL")
	}
	return sqlf.Sprintf("%s", *o.First)
}

const listEmailActionsFmtStr = `
SELECT %s -- EmailsColumns
FROM cm_emails
WHERE %s
ORDER BY id ASC
LIMIT %s;
`

// ListEmailActions lists emails from cm_emails with the given opts
func (s *codeMonitorStore) ListEmailActions(ctx context.Context, opts ListActionsOpts) ([]*MonitorEmail, error) {
	q := sqlf.Sprintf(
		listEmailActionsFmtStr,
		sqlf.Join(emailsColumns, ","),
		opts.Conds(),
		opts.Limit(),
	)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEmails(rows)
}

const createActionEmailFmtStr = `
INSERT INTO cm_emails
(monitor, enabled, priority, header, created_by, created_at, changed_by, changed_at)
VALUES (%s,%s,%s,%s,%s,%s,%s,%s)
RETURNING %s;
`

func (s *codeMonitorStore) createActionEmailQuery(ctx context.Context, monitorID int64, args *graphqlbackend.CreateActionEmailArgs) (*sqlf.Query, error) {
	now := s.Now()
	a := actor.FromContext(ctx)
	return sqlf.Sprintf(
		createActionEmailFmtStr,
		monitorID,
		args.Enabled,
		args.Priority,
		args.Header,
		a.UID,
		now,
		a.UID,
		now,
		sqlf.Join(emailsColumns, ", "),
	), nil
}

const deleteActionEmailFmtStr = `
DELETE FROM cm_emails
WHERE id in (%s)
	AND MONITOR = %s
`

func deleteActionsEmailQuery(ctx context.Context, actionIDs []int64, monitorID int64) (*sqlf.Query, error) {
	deleteIDs := make([]*sqlf.Query, 0, len(actionIDs))
	for _, ids := range actionIDs {
		deleteIDs = append(deleteIDs, sqlf.Sprintf("%d", ids))
	}
	return sqlf.Sprintf(
		deleteActionEmailFmtStr,
		sqlf.Join(deleteIDs, ", "),
		monitorID,
	), nil
}

// emailColumns is the set of columns in the cm_emails table
// This must be kept in sync with scanEmail
var emailsColumns = []*sqlf.Query{
	sqlf.Sprintf("cm_emails.id"),
	sqlf.Sprintf("cm_emails.monitor"),
	sqlf.Sprintf("cm_emails.enabled"),
	sqlf.Sprintf("cm_emails.priority"),
	sqlf.Sprintf("cm_emails.header"),
	sqlf.Sprintf("cm_emails.created_by"),
	sqlf.Sprintf("cm_emails.created_at"),
	sqlf.Sprintf("cm_emails.changed_by"),
	sqlf.Sprintf("cm_emails.changed_at"),
}

func scanEmails(rows *sql.Rows) ([]*MonitorEmail, error) {
	var ms []*MonitorEmail
	for rows.Next() {
		m, err := scanEmail(rows)
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	return ms, rows.Err()
}

// scanEmail scans a MonitorEmail from a *sql.Row or *sql.Rows.
// It must be kept in sync with emailsColumns.
func scanEmail(scanner dbutil.Scanner) (*MonitorEmail, error) {
	m := &MonitorEmail{}
	err := scanner.Scan(
		&m.Id,
		&m.Monitor,
		&m.Enabled,
		&m.Priority,
		&m.Header,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.ChangedBy,
		&m.ChangedAt,
	)
	return m, err
}
