package userpasswd

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var MockSendEmail func(context.Context, string, txemail.Message) (err error)

// LockoutStore provides semantics for account lockout management.
type LockoutStore interface {
	// IsLockedOut returns true if the given user has been locked along with the
	// reason.
	IsLockedOut(userID int32) (reason string, locked bool)
	// IncreaseFailedAttempt increases the failed login attempt count by 1.
	IncreaseFailedAttempt(userID int32)
	// Reset clears the failed login attempt count and releases the lockout.
	Reset(userID int32)
	// GenerateUnlockAccountURL creates the URL to unlock account with a signet
	// unlock token.
	GenerateUnlockAccountURL(userID int32) (string, string, error)
	// VerifyUnlockAccountTokenAndReset verifies the provided unlock token is valid.
	VerifyUnlockAccountTokenAndReset(urlToken string) (bool, error)
	// SendUnlockAccountEmail sends an email to the locked account email providing a
	// temporary unlock link.
	SendUnlockAccountEmail(ctx context.Context, userID int32, userEmail string) error
	// Returns true if the unlock account email has already been sent
	UnlockEmailSent(userID int32) bool
}

type lockoutStore struct {
	failedThreshold int
	lockouts        *rcache.Cache
	failedAttempts  *rcache.Cache
	unlockToken     *rcache.Cache
	unlockEmailSent *rcache.Cache
}

// NewLockoutStore returns a new LockoutStore with given durations using the
// Redis cache.
func NewLockoutStore(failedThreshold int, lockoutPeriod, consecutivePeriod time.Duration) LockoutStore {
	return &lockoutStore{
		failedThreshold: failedThreshold,
		lockouts:        rcache.NewWithTTL("account_lockout", int(lockoutPeriod.Seconds())),
		failedAttempts:  rcache.NewWithTTL("account_failed_attempts", int(consecutivePeriod.Seconds())),
		unlockToken:     rcache.NewWithTTL("account_unlock_token", int(lockoutPeriod.Seconds())),
		unlockEmailSent: rcache.NewWithTTL("account_lockout_email_sent", int(lockoutPeriod.Seconds())),
	}
}

func key(userID int32) string {
	return strconv.Itoa(int(userID))
}

func (s *lockoutStore) IsLockedOut(userID int32) (reason string, locked bool) {
	v, locked := s.lockouts.Get(key(userID))
	return string(v), locked
}

func (s *lockoutStore) IncreaseFailedAttempt(userID int32) {
	metricsAccountFailedSignInAttempts.Inc()

	key := key(userID)
	s.failedAttempts.Increase(key)

	// Get right after Increase should make the key always exist
	v, _ := s.failedAttempts.Get(key)
	count, _ := strconv.Atoi(string(v))
	if count >= s.failedThreshold {
		metricsAccountLockouts.Inc()
		s.lockouts.Set(key, []byte("too many failed attempts"))
	}
}

type unlockAccountClaims struct {
	UserID int32 `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *lockoutStore) GenerateUnlockAccountURL(userID int32) (string, string, error) {
	key := key(userID)
	ttl, exists := s.lockouts.KeyTTL(key)

	if !exists {
		return "", "", errors.Newf("user with id %d is not locked out, cannot generate unlock url")
	}

	signingKey := conf.SiteConfig().AuthUnlockAccountLinkSigningKey
	if signingKey == "" {
		return "", "", errors.Newf(`signing key not provided, cannot validate JWT on unlock account URL. Please add "auth.unlockAccountLinkSigningKey" to site configuration.`)
	}

	expiryTime := time.Now().Add(time.Second * time.Duration(ttl))

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, &unlockAccountClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    globals.ExternalURL().String(),
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			Subject:   strconv.FormatInt(int64(userID), 10),
		},
		UserID: userID,
	})

	// Sign and get the complete encoded token as a string using the secret
	decodedSigningKey, err := base64.StdEncoding.DecodeString(signingKey)
	if err != nil {
		return "", "", err
	}
	tokenString, err := token.SignedString(decodedSigningKey)
	if err != nil {
		return "", "", err
	}

	s.unlockToken.SetWithTTL(key, []byte(tokenString), ttl)

	path := fmt.Sprintf("/unlock-account/%s", tokenString)

	return globals.ExternalURL().ResolveReference(&url.URL{Path: path}).String(), tokenString, nil
}

func formatExpiryTime(ttl int) string {
	minutes := ttl / 60
	seconds := ttl

	if minutes < 1 {
		return fmt.Sprintf("%d seconds", seconds)
	}
	return fmt.Sprintf("%d minutes", minutes)
}

func (s *lockoutStore) SendUnlockAccountEmail(ctx context.Context, userID int32, recipientEmail string) error {
	key := key(userID)
	ttl, exists := s.lockouts.KeyTTL(key)

	if !exists || s.UnlockEmailSent(userID) {
		return nil
	}

	unlockUrl, _, err := s.GenerateUnlockAccountURL(userID)
	if err != nil {
		return err
	}

	emailF := txemail.Send
	if MockSendEmail != nil {
		emailF = MockSendEmail
	}
	err = emailF(ctx, "account_unlock", txemail.Message{
		To:       []string{recipientEmail},
		Template: emailTemplates,
		Data: struct {
			UnlockAccountUrl string
			ExpiryTime       string
		}{
			UnlockAccountUrl: unlockUrl,
			ExpiryTime:       formatExpiryTime(ttl),
		},
	})
	if err != nil {
		return err
	}

	s.unlockEmailSent.SetWithTTL(key, []byte("sent"), ttl)
	return nil
}

func (s *lockoutStore) UnlockEmailSent(userID int32) bool {
	_, locked := s.unlockEmailSent.Get(key(userID))
	return locked
}

func (s *lockoutStore) VerifyUnlockAccountTokenAndReset(urlToken string) (bool, error) {
	signingKey := conf.SiteConfig().AuthUnlockAccountLinkSigningKey

	if signingKey == "" {
		return false, errors.Newf("signing key not provided, cannot validate JWT on account reset URL. Please add AuthUnlockAccountLinkSigningKey to site configuration.")
	}

	token, err := jwt.ParseWithClaims(urlToken, &unlockAccountClaims{}, func(token *jwt.Token) (any, error) {
		return base64.StdEncoding.DecodeString(signingKey)
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Name}))

	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(*unlockAccountClaims); ok && token.Valid {
		userIdKey := key(claims.UserID)
		storedToken, found := s.unlockToken.Get(userIdKey)

		if !found || string(storedToken) != urlToken {
			return false, errors.Newf("No previously generated token exists for the specified user")
		}

		s.Reset(claims.UserID)
		return true, nil
	}

	return false, errors.Newf("provided token is invalid or expired")
}

func (s *lockoutStore) Reset(userID int32) {
	key := key(userID)
	s.lockouts.Delete(key)
	s.failedAttempts.Delete(key)
	s.unlockToken.Delete(key)
	s.unlockEmailSent.Delete(key)
}

var emailTemplates = txemail.MustValidate(txtypes.Templates{
	Subject: `Unlock your Sourcegraph Cloud account`,
	Text: `
You are receiving this email because your Sourcegraph account got locked after too many sign in attempts.

Please, visit this link in your browser to unlock the account and try to sign in again: {{.UnlockAccountUrl}}

This link will expire in {{.ExpiryTime}}.

To see our Terms of Service, please visit this link: https://about.sourcegraph.com/terms
To see our Privacy Policy, please visit this link: https://about.sourcegraph.com/privacy

Sourcegraph, 981 Mission St, San Francisco, CA 94103, USA
`,
	HTML: `
<html>
<head>
  <meta name="color-scheme" content="light">
  <meta name="supported-color-schemes" content="light">
  <style>
    body { color: #343a4d; background: #fff; padding: 20px; font-size: 16px; font-family: -apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica Neue,Arial,Noto Sans,sans-serif,Apple Color Emoji,Segoe UI Emoji,Segoe UI Symbol,Noto Color Emoji; }
    .logo { height: 34px; margin-bottom: 15px; }
    a { color: #0b70db; text-decoration: none; background-color: transparent; }
    a:hover { color: #0c7bf0; text-decoration: underline; }
    a.btn { display: inline-block; color: #fff; background-color: #0b70db; padding: 8px 16px; border-radius: 3px; font-weight: 600; }
    a.btn:hover { color: #fff; background-color: #0864c6; text-decoration:none; }
    .smaller { font-size: 14px; }
    small { color: #5e6e8c; font-size: 12px; }
    .mtm { margin-top: 10px; }
    .mtl { margin-top: 20px; }
    .mtxl { margin-top: 30px; }
  </style>
</head>
<body style="font-family: -apple-system,BlinkMacSystemFont,Segoe UI,Roboto,Helvetica Neue,Arial,Noto Sans,sans-serif,Apple Color Emoji,Segoe UI Emoji,Segoe UI Symbol,Noto Color Emoji;">
  <img class="logo" src="https://storage.googleapis.com/sourcegraph-assets/sourcegraph-logo-light-small.png" alt="Sourcegraph logo">
  <p>
  	You are receiving this email because your Sourcegraph account got locked after too many sign in attempts..
  </p>
  <p class="mtxl">
    Please, follow this link in your browser to unlock your account and try to sign in again: <a class="btn mtm" href="{{.UnlockAccountUrl}}">Unlock your Account</a>
  </p>
  <p class="smaller">Or visit this link in your browser: <a href="{{.UnlockAccountUrl}}">{{.UnlockAccountUrl}}</a></p>
  <small>
  <p class="mtl">
    This link will expire in {{.ExpiryTime}}.
  </p>
  <p class="mtl">
    <a href="https://about.sourcegraph.com/terms">Terms</a>&nbsp;&#8226;&nbsp;
    <a href="https://about.sourcegraph.com/privacy">Privacy</a>
  </p>
  <p>Sourcegraph, 981 Mission St, San Francisco, CA 94103, USA</p>
  </small>
</body>
</html>
`,
})
