package rds

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
	"github.com/jackc/pgx/v4"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Auth implements the dbconn.AuthProvider interface for getting an auth token for RDS IAM auth.
// It will retrieve an auth token from EC2 metadata service during startup time, and use it to
// create a connection to RDS.
type Auth struct{}

func NewAuth() *Auth {
	return &Auth{}
}

func (a *Auth) IsRefresh(logger log.Logger, cfg *pgx.ConnConfig) bool {
	token, err := parseRDSAuthToken(cfg.Password)
	if err != nil {
		logger.Warn("Error parsing RDS auth token, refreshing", log.Error(err))
		return false
	}

	return token.isExpired(time.Now().UTC())
}

func (a *Auth) Apply(logger log.Logger, cfg *pgx.ConnConfig) error {
	if cfg.Password != "" {
		logger.Warn("'PG_AUTH_PROVIDER' is 'EC2_ROLE_CREDENTIALS', but 'PGPASSWORD' is also set. Ignoring 'PGPASSWORD'.")
	}

	var err error
	cfg.Password, err = authToken(cfg.Host, cfg.Port, cfg.User)
	if err != nil {
		return errors.Wrap(err, "Error getting auth token for RDS IAM auth")
	}

	return nil
}

func authToken(hostname string, port uint16, user string) (string, error) {
	instance, err := parseRDSHostname(hostname)
	if err != nil {
		return "", errors.Wrap(err, "Error parsing RDS hostname")
	}

	sess, err := session.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "Error creating AWS session for RDS IAM auth")
	}
	creds := sess.Config.Credentials
	if creds == nil {
		return "", errors.New("No AWS credentials found from current session")
	}

	authToken, err := rdsutils.BuildAuthToken(
		fmt.Sprintf("%s:%d", instance.hostname, port),
		instance.region, user, creds)
	if err != nil {
		return "", errors.Wrap(err, "Error building auth token for RDS IAM auth")
	}

	return authToken, nil
}

type rdsAuthToken struct {
	IssuedAt  time.Time
	ExpiresIn time.Duration
}

// parseRDSAuthToken parses the auth token from RDS IAM auth.
// Learn more from unit test cases
func parseRDSAuthToken(token string) (*rdsAuthToken, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s", token))
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing RDS auth token")
	}

	// specific about the query string parameters can be found from:
	// https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-query-string-auth.html
	xAmzDate := u.Query().Get("X-Amz-Date")
	if xAmzDate == "" {
		return nil, errors.New("Missing X-Amz-Date in RDS auth token, <redacted>")
	}
	// e.g., 20160801T223241Z
	issuedAt, err := time.Parse("20060102T150405Z", xAmzDate)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing X-Amz-Date in RDS auth token, %q", xAmzDate)
	}

	xAmzExpires := u.Query().Get("X-Amz-Expires")
	if xAmzExpires == "" {
		return nil, errors.New("Missing X-Amz-Expires in RDS auth token, <redacted>")
	}
	expiresIn, err := strconv.Atoi(xAmzExpires)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing X-Amz-Expires in RDS auth token, %q", xAmzExpires)
	}

	return &rdsAuthToken{
		IssuedAt:  issuedAt,
		ExpiresIn: time.Duration(expiresIn) * time.Second,
	}, nil
}

// isExpired returns true if the token is expired
// with a 5 minutes grace period.
func (t *rdsAuthToken) isExpired(now time.Time) bool {
	// 300 secs buffer to avoid the token being expired when it is used
	return now.UTC().Add(-300 * time.Second).After(t.IssuedAt.Add(t.ExpiresIn))
}

type rdsInstance struct {
	region   string
	hostname string
}

// parseRDSHostname parses the RDS hostname and returns the region and instance name.
// It is in the form of <instance-name>.<account-id>.<region>.rds.amazonaws.com
// e.g., postgresmydb.123456789012.us-east-1.rds.amazonaws.com
func parseRDSHostname(name string) (*rdsInstance, error) {
	if !strings.HasSuffix(name, ".rds.amazonaws.com") {
		return nil, errors.Newf("not an RDS hostname, expecting '.rds.amazonaws.com' suffix, %q", name)
	}

	parts := strings.Split(name, ".")
	if len(parts) != 6 {
		return nil, errors.Newf("unexpected RDS hostname format, %q", name)
	}

	if parts[0] == "" {
		return nil, errors.Newf("unexpected instance name in RDS hostname format, %q", name)
	}

	if parts[1] == "" {
		return nil, errors.Newf("unexpected account ID in RDS hostname format, %q", name)
	}

	if parts[2] == "" {
		return nil, errors.Newf("unexpected region in RDS hostname format, %q", name)
	}

	return &rdsInstance{
		region:   parts[2],
		hostname: name,
	}, nil
}
