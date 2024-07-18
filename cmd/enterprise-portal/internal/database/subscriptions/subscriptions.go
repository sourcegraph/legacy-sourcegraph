package subscriptions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/upsert"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/internal/utctime"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ⚠️ DO NOT USE: This type is only used for creating foreign key constraints
// and initializing tables with gorm.
type TableSubscription struct {
	// Each Subscription has many Licenses.
	Licenses []*TableSubscriptionLicense `gorm:"foreignKey:SubscriptionID"`

	// Each Subscription has many Conditions.
	Conditions []*SubscriptionCondition `gorm:"foreignKey:SubscriptionID"`

	Subscription
}

func (*TableSubscription) TableName() string {
	return "enterprise_portal_subscriptions"
}

// Subscription is an Enterprise subscription record.
type Subscription struct {
	// ID is the internal (unprefixed) UUID-format identifier for the subscription.
	ID string `gorm:"type:uuid;primaryKey"`
	// InstanceDomain is the instance domain associated with the subscription, e.g.
	// "acme.sourcegraphcloud.com". This is set explicitly.
	//
	// It must be unique across all currently un-archived subscriptions.
	InstanceDomain *string `gorm:"uniqueIndex:,where:archived_at IS NULL"`

	// WARNING: The below fields are not yet used in production.

	// DisplayName is the human-friendly name of this subscription, e.g. "Acme, Inc."
	//
	// It must be unique across all currently un-archived subscriptions, unless
	// it is not set.
	DisplayName *string `gorm:"size:256;uniqueIndex:,where:archived_at IS NULL"`

	// Timestamps representing the latest timestamps of key conditions related
	// to this subscription.
	//
	// Condition transition details are tracked in 'enterprise_portal_subscription_conditions'.
	CreatedAt  utctime.Time  `gorm:"not null;default:current_timestamp"`
	UpdatedAt  utctime.Time  `gorm:"not null;default:current_timestamp"`
	ArchivedAt *utctime.Time // Null indicates the subscription is not archived.

	// SalesforceSubscriptionID associated with this Enterprise subscription.
	SalesforceSubscriptionID *string
	// SalesforceOpportunityID associated with this Enterprise subscription.
	SalesforceOpportunityID *string
}

// subscriptionTableColumns must match scanSubscription() values.
func subscriptionTableColumns() []string {
	return []string{
		"id",
		"instance_domain",
		"display_name",
		"created_at",
		"updated_at",
		"archived_at",
		"salesforce_subscription_id",
		"salesforce_opportunity_id",
	}
}

// scanSubscription matches subscriptionTableColumns() values.
func scanSubscription(row pgx.Row) (*Subscription, error) {
	var s Subscription
	err := row.Scan(
		&s.ID,
		&s.InstanceDomain,
		&s.DisplayName,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.ArchivedAt,
		&s.SalesforceSubscriptionID,
		&s.SalesforceOpportunityID,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Store is the storage layer for product subscriptions.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

// ListEnterpriseSubscriptionsOptions is the set of options to filter subscriptions.
// Non-empty fields are treated as AND-concatenated.
type ListEnterpriseSubscriptionsOptions struct {
	// IDs is a list of subscription IDs to filter by.
	IDs []string
	// InstanceDomains is a list of instance domains to filter by.
	InstanceDomains []string
	// IsArchived indicates whether to only list archived subscriptions, or only
	// non-archived subscriptions.
	IsArchived bool
	// PageSize is the maximum number of subscriptions to return.
	PageSize int
}

func (opts ListEnterpriseSubscriptionsOptions) toQueryConditions() (where, limit string, _ pgx.NamedArgs) {
	whereConds := []string{"TRUE"}
	namedArgs := pgx.NamedArgs{}
	if len(opts.IDs) > 0 {
		whereConds = append(whereConds, "id = ANY(@ids)")
		namedArgs["ids"] = opts.IDs
	}
	if len(opts.InstanceDomains) > 0 {
		whereConds = append(whereConds, "instance_domain = ANY(@instanceDomains)")
		namedArgs["instanceDomains"] = opts.InstanceDomains
	}
	// Future: Uncomment the following block when the archived field is added to the table.
	// if opts.OnlyArchived {
	// whereConds = append(whereConds, "archived = TRUE")
	// }
	where = strings.Join(whereConds, " AND ")

	if opts.PageSize > 0 {
		limit = "LIMIT @pageSize"
		namedArgs["pageSize"] = opts.PageSize
	}
	return where, limit, namedArgs
}

// List returns a list of subscriptions based on the given options.
func (s *Store) List(ctx context.Context, opts ListEnterpriseSubscriptionsOptions) ([]*Subscription, error) {
	where, limit, namedArgs := opts.toQueryConditions()
	query := fmt.Sprintf(`
SELECT
	%s
FROM enterprise_portal_subscriptions
WHERE %s
%s`,
		strings.Join(subscriptionTableColumns(), ", "),
		where, limit,
	)
	rows, err := s.db.Query(ctx, query, namedArgs)
	if err != nil {
		return nil, errors.Wrap(err, "query rows")
	}
	defer rows.Close()

	var subscriptions []*Subscription
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scan row")
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, rows.Err()
}

type UpsertSubscriptionOptions struct {
	InstanceDomain *sql.NullString
	DisplayName    *sql.NullString

	CreatedAt  time.Time
	ArchivedAt *time.Time

	SalesforceSubscriptionID *string
	SalesforceOpportunityID  *string

	// ForceUpdate indicates whether to force update all fields of the subscription
	// record.
	ForceUpdate bool
}

// toQuery returns the query based on the options. It returns an empty query if
// nothing to update.
func (opts UpsertSubscriptionOptions) Exec(ctx context.Context, db *pgxpool.Pool, id string) error {
	b := upsert.New("enterprise_portal_subscriptions", "id", opts.ForceUpdate)
	upsert.Field(b, "id", id)
	upsert.Field(b, "instance_domain", opts.InstanceDomain)
	upsert.Field(b, "display_name", opts.DisplayName)

	upsert.Field(b, "created_at", opts.CreatedAt,
		upsert.WithColumnDefault(),
		upsert.WithIgnoreOnForceUpdate())
	upsert.Field(b, "updated_at", time.Now()) // always updated now
	upsert.Field(b, "archived_at", opts.ArchivedAt)
	upsert.Field(b, "salesforce_subscription_id", opts.SalesforceSubscriptionID)
	upsert.Field(b, "salesforce_opportunity_id", opts.SalesforceOpportunityID)
	return b.Exec(ctx, db)
}

// Upsert upserts a subscription record based on the given options.
func (s *Store) Upsert(ctx context.Context, subscriptionID string, opts UpsertSubscriptionOptions) (*Subscription, error) {
	if err := opts.Exec(ctx, s.db, subscriptionID); err != nil {
		return nil, errors.Wrap(err, "exec")
	}
	return s.Get(ctx, subscriptionID)
}

// Get returns a subscription record with the given subscription ID. It returns
// pgx.ErrNoRows if no such subscription exists.
func (s *Store) Get(ctx context.Context, subscriptionID string) (*Subscription, error) {
	query := fmt.Sprintf(`SELECT
		%s
	FROM
		enterprise_portal_subscriptions
	WHERE
		id = @id`,
		strings.Join(subscriptionTableColumns(), ", "))
	namedArgs := pgx.NamedArgs{"id": subscriptionID}

	sub, err := scanSubscription(s.db.QueryRow(ctx, query, namedArgs))
	if err != nil {
		return nil, err
	}
	return sub, nil
}
