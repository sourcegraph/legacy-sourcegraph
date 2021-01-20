package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgconn"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

// An OrgInvitation is an invitation for a user to join an organization as a member.
type OrgInvitation struct {
	ID              int64
	OrgID           int32
	SenderUserID    int32 // the sender of the invitation
	RecipientUserID int32 // the recipient of the invitation
	CreatedAt       time.Time
	NotifiedAt      *time.Time
	RespondedAt     *time.Time
	ResponseType    *bool // accepted (true), rejected (false), no response (nil)
	RevokedAt       *time.Time
}

// Pending reports whether the invitation is pending (i.e., can be responded to by the recipient
// because it has not been revoked or responded to yet).
func (oi *OrgInvitation) Pending() bool {
	return oi.RespondedAt == nil && oi.RevokedAt == nil
}

type OrgInvitationStore struct {
	*basestore.Store

	once sync.Once
}

// NewOrgInvitationStoreWithDB instantiates and returns a new OrgInvitationStore with prepared statements.
func NewOrgInvitationStoreWithDB(db dbutil.DB) *OrgInvitationStore {
	return &OrgInvitationStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// NewOrgInvitationStoreWithDB instantiates and returns a new OrgInvitationStore using the other store handle.
func NewOrgInvitationStoreWith(other basestore.ShareableStore) *OrgInvitationStore {
	return &OrgInvitationStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *OrgInvitationStore) With(other basestore.ShareableStore) *OrgInvitationStore {
	return &OrgInvitationStore{Store: s.Store.With(other)}
}

func (s *OrgInvitationStore) Transact(ctx context.Context) (*OrgInvitationStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &OrgInvitationStore{Store: txBase}, err
}

// ensureStore instantiates a basestore.Store if necessary, using the dbconn.Global handle.
// This function ensures access to dbconn happens after the rest of the code or tests have
// initialized it.
func (s *OrgInvitationStore) ensureStore() {
	s.once.Do(func() {
		if s.Store == nil {
			s.Store = basestore.NewWithDB(dbconn.Global, sql.TxOptions{})
		}
	})
}

// OrgInvitationNotFoundError occurs when an org invitation is not found.
type OrgInvitationNotFoundError struct {
	args []interface{}
}

// NotFound implements errcode.NotFounder.
func (err OrgInvitationNotFoundError) NotFound() bool { return true }

func (err OrgInvitationNotFoundError) Error() string {
	return fmt.Sprintf("org invitation not found: %v", err.args)
}

func (s *OrgInvitationStore) Create(ctx context.Context, orgID, senderUserID, recipientUserID int32) (*OrgInvitation, error) {
	if Mocks.OrgInvitations.Create != nil {
		return Mocks.OrgInvitations.Create(orgID, senderUserID, recipientUserID)
	}
	s.ensureStore()

	t := &OrgInvitation{
		OrgID:           orgID,
		SenderUserID:    senderUserID,
		RecipientUserID: recipientUserID,
	}
	if err := s.Handle().DB().QueryRowContext(
		ctx,
		"INSERT INTO org_invitations(org_id, sender_user_id, recipient_user_id) VALUES($1, $2, $3) RETURNING id, created_at",
		orgID, senderUserID, recipientUserID,
	).Scan(&t.ID, &t.CreatedAt); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.ConstraintName {
			case "org_invitations_singleflight":
				return nil, errors.New("user was already invited to organization (and has not responded yet)")
			}
		}
		return nil, err
	}
	return t, nil
}

// GetByID retrieves the org invitation (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this org invitation.
func (s *OrgInvitationStore) GetByID(ctx context.Context, id int64) (*OrgInvitation, error) {
	if Mocks.OrgInvitations.GetByID != nil {
		return Mocks.OrgInvitations.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("id=%d", id)}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, OrgInvitationNotFoundError{[]interface{}{id}}
	}
	return results[0], nil
}

// GetPending retrieves the pending invitation (if any) for the recipient to join the org. At most
// one invitation may be pending for an (org,recipient).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this org invitation.
func (s *OrgInvitationStore) GetPending(ctx context.Context, orgID, recipientUserID int32) (*OrgInvitation, error) {
	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("org_id=%d AND recipient_user_id=%d AND responded_at IS NULL AND revoked_at IS NULL", orgID, recipientUserID),
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, OrgInvitationNotFoundError{[]interface{}{fmt.Sprintf("pending for org %d recipient %d", orgID, recipientUserID)}}
	}
	return results[0], nil
}

// OrgInvitationsListOptions contains options for listing org invitations.
type OrgInvitationsListOptions struct {
	OrgID           int32 // only list org invitations for this org
	RecipientUserID int32 // only list org invitations with this user as the recipient
	*LimitOffset
}

func (o OrgInvitationsListOptions) sqlConditions() []*sqlf.Query {
	var conds []*sqlf.Query
	if o.OrgID != 0 {
		conds = append(conds, sqlf.Sprintf("org_id=%d", o.OrgID))
	}
	if o.RecipientUserID != 0 {
		conds = append(conds, sqlf.Sprintf("recipient_user_id=%d", o.RecipientUserID))
	}
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}
	return conds
}

// List lists all access tokens that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s *OrgInvitationStore) List(ctx context.Context, opt OrgInvitationsListOptions) ([]*OrgInvitation, error) {
	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s *OrgInvitationStore) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*OrgInvitation, error) {
	s.ensureStore()

	q := sqlf.Sprintf(`
SELECT id, org_id, sender_user_id, recipient_user_id, created_at, notified_at, responded_at, response_type, revoked_at FROM org_invitations
WHERE (%s) AND deleted_at IS NULL
ORDER BY id ASC
%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*OrgInvitation
	for rows.Next() {
		var t OrgInvitation
		if err := rows.Scan(&t.ID, &t.OrgID, &t.SenderUserID, &t.RecipientUserID, &t.CreatedAt, &t.NotifiedAt, &t.RespondedAt, &t.ResponseType, &t.RevokedAt); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}
	return results, nil
}

// Count counts all org invitations that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the invitations.
func (s *OrgInvitationStore) Count(ctx context.Context, opt OrgInvitationsListOptions) (int, error) {
	s.ensureStore()

	q := sqlf.Sprintf("SELECT COUNT(*) FROM org_invitations WHERE (%s) AND deleted_at IS NULL", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateEmailSentTimestamp updates the email-sent timestam[ for the org invitation to the current
// time.
func (s *OrgInvitationStore) UpdateEmailSentTimestamp(ctx context.Context, id int64) error {
	s.ensureStore()

	res, err := s.Handle().DB().ExecContext(ctx, "UPDATE org_invitations SET notified_at=now() WHERE id=$1 AND revoked_at IS NULL AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return OrgInvitationNotFoundError{[]interface{}{id}}
	}
	return nil
}

// Respond sets the recipient's response to the org invitation and returns the organization's ID to
// which the recipient was invited. If the recipient user ID given is incorrect, an
// OrgInvitationNotFoundError error is returned.
func (s *OrgInvitationStore) Respond(ctx context.Context, id int64, recipientUserID int32, accept bool) (orgID int32, err error) {
	s.ensureStore()

	if err := s.Handle().DB().QueryRowContext(ctx, "UPDATE org_invitations SET responded_at=now(), response_type=$3 WHERE id=$1 AND recipient_user_id=$2 AND responded_at IS NULL AND revoked_at IS NULL AND deleted_at IS NULL RETURNING org_id", id, recipientUserID, accept).Scan(&orgID); err == sql.ErrNoRows {
		return 0, OrgInvitationNotFoundError{[]interface{}{fmt.Sprintf("id %d recipient %d", id, recipientUserID)}}
	} else if err != nil {
		return 0, err
	}
	return orgID, nil
}

// Revoke marks an org invitation as revoked. The recipient is forbidden from responding to it after
// it has been revoked.
func (s *OrgInvitationStore) Revoke(ctx context.Context, id int64) error {
	if Mocks.OrgInvitations.Revoke != nil {
		return Mocks.OrgInvitations.Revoke(id)
	}
	s.ensureStore()

	res, err := s.Handle().DB().ExecContext(ctx, "UPDATE org_invitations SET revoked_at=now() WHERE id=$1 AND revoked_at IS NULL AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return OrgInvitationNotFoundError{[]interface{}{id}}
	}
	return nil
}

// MockOrgInvitations mocks the org invitations store.
type MockOrgInvitations struct {
	Create  func(orgID, senderUserID, recipientUserID int32) (*OrgInvitation, error)
	GetByID func(id int64) (*OrgInvitation, error)
	Revoke  func(id int64) error
}
