package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/audit"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SecurityEventName string

const (
	SecurityEventNameSignOutAttempted SecurityEventName = "SignOutAttempted"
	SecurityEventNameSignOutFailed    SecurityEventName = "SignOutFailed"
	SecurityEventNameSignOutSucceeded SecurityEventName = "SignOutSucceeded"

	SecurityEventNameSignInAttempted SecurityEventName = "SignInAttempted"
	SecurityEventNameSignInFailed    SecurityEventName = "SignInFailed"
	SecurityEventNameSignInSucceeded SecurityEventName = "SignInSucceeded"

	SecurityEventNameAccountCreated SecurityEventName = "AccountCreated"
	SecurityEventNameAccountDeleted SecurityEventName = "AccountDeleted"
	SecurityEventNameAccountNuked   SecurityEventName = "AccountNuked"

	SecurityEventNamPasswordResetRequested SecurityEventName = "PasswordResetRequested"
	SecurityEventNamPasswordRandomized     SecurityEventName = "PasswordRandomized"
	SecurityEventNamePasswordChanged       SecurityEventName = "PasswordChanged"

	SecurityEventNameEmailVerified SecurityEventName = "EmailVerified"

	SecurityEventNameRoleChangeDenied  SecurityEventName = "RoleChangeDenied"
	SecurityEventNameRoleChangeGranted SecurityEventName = "RoleChangeGranted"

	SecurityEventNameAccessGranted SecurityEventName = "AccessGranted"
)

// SecurityEvent contains information needed for logging a security-relevant event.
type SecurityEvent struct {
	Name            SecurityEventName
	URL             string
	UserID          uint32
	AnonymousUserID string
	Argument        json.RawMessage
	Source          string
	Timestamp       time.Time
}

func (e *SecurityEvent) parseArgument() string {
	if e.Argument == nil {
		return "{}"
	}
	return fmt.Sprintf("%s", e.Argument)
}

// SecurityEventLogsStore provides persistence for security events.
type SecurityEventLogsStore interface {
	basestore.ShareableStore

	// Insert adds a new security event to the store.
	Insert(ctx context.Context, e *SecurityEvent) error
	// Bulk "Insert" action.
	InsertList(ctx context.Context, events []*SecurityEvent) error
	// LogEvent logs the given security events.
	//
	// It logs errors directly instead of returning to callers.
	LogEvent(ctx context.Context, e *SecurityEvent)
	// Bulk "LogEvent" action.
	LogEventList(ctx context.Context, events []*SecurityEvent)
}

type securityEventLogsStore struct {
	logger log.Logger
	*basestore.Store
}

// SecurityEventLogsWith instantiates and returns a new SecurityEventLogsStore
// using the other store handle, and a scoped sub-logger of the passed base logger.
func SecurityEventLogsWith(baseLogger log.Logger, other basestore.ShareableStore) SecurityEventLogsStore {
	logger := baseLogger.Scoped("SecurityEvents", "Security events store")
	return &securityEventLogsStore{logger: logger, Store: basestore.NewWithHandle(other.Handle())}
}

func (s *securityEventLogsStore) Insert(ctx context.Context, event *SecurityEvent) error {
	return s.InsertList(ctx, []*SecurityEvent{event})
}

func (s *securityEventLogsStore) InsertList(ctx context.Context, events []*SecurityEvent) error {
	vals := make([]*sqlf.Query, len(events))
	for index, event := range events {
		vals[index] = sqlf.Sprintf(`(%s, %s, %s, %s, %s, %s, %s, %s)`,
			event.Name,
			event.URL,
			event.UserID,
			event.AnonymousUserID,
			event.Source,
			event.parseArgument(),
			version.Version(),
			event.Timestamp.UTC(),
		)
	}
	query := sqlf.Sprintf("INSERT INTO security_event_logs(name, url, user_id, anonymous_user_id, source, argument, version, timestamp) VALUES %s", sqlf.Join(vals, ","))

	if _, err := s.Handle().ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		return errors.Wrap(err, "INSERT")
	} else if auditLogEnabled() {
		for _, event := range events {
			audit.Log(ctx, s.logger, audit.Record{
				Entity: "security events",
				Action: string(event.Name),
				Fields: []log.Field{
					log.Object("event",
						log.String("URL", event.URL),
						log.String("source", event.Source),
						log.String("argument", event.parseArgument()),
						log.String("version", version.Version()),
						log.String("timestamp", event.Timestamp.UTC().String()),
					),
				},
			})
		}
	}
	return nil
}

func securityEventsAuditLogEnabled() bool {
	if logCfg := conf.Get().Log; logCfg != nil {
		if auditCfg := logCfg.AuditLog; auditCfg != nil {
			return auditCfg.SecurityEvents
		}
	}

	// enabled by default for security events
	return true
}

func (s *securityEventLogsStore) LogEvent(ctx context.Context, e *SecurityEvent) {
	s.LogEventList(ctx, []*SecurityEvent{e})
}

func (s *securityEventLogsStore) LogEventList(ctx context.Context, events []*SecurityEvent) {
	if err := s.InsertList(ctx, events); err != nil {
		names := make([]string, len(events))
		for i, e := range events {
			names[i] = string(e.Name)
		}
		j, _ := json.Marshal(&events)
		trace.Logger(ctx, s.logger).Error(strings.Join(names, ","), log.String("events", string(j)), log.Error(err))
	}
}
