package conf

import (
	"context"
	"encoding/base64"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/cronexpr"

	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/dotcomuser"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	srccli "github.com/sourcegraph/sourcegraph/internal/src-cli"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	deployType := deploy.Type()
	if !deploy.IsValidDeployType(deployType) {
		log.Fatalf("The 'DEPLOY_TYPE' environment variable is invalid. Expected one of: %q, %q, %q, %q, %q, %q, %q. Got: %q", deploy.Kubernetes, deploy.DockerCompose, deploy.PureDocker, deploy.SingleDocker, deploy.Dev, deploy.Helm, deploy.App, deployType)
	}

	confdefaults.Default = defaultConfigForDeployment()
}

func defaultConfigForDeployment() conftypes.RawUnified {
	deployType := deploy.Type()
	switch {
	case deploy.IsDev(deployType):
		return confdefaults.DevAndTesting
	case deploy.IsDeployTypeSingleDockerContainer(deployType):
		return confdefaults.DockerContainer
	case deploy.IsDeployTypeKubernetes(deployType), deploy.IsDeployTypeDockerCompose(deployType), deploy.IsDeployTypePureDocker(deployType):
		return confdefaults.KubernetesOrDockerComposeOrPureDocker
	case deploy.IsDeployTypeApp(deployType):
		return confdefaults.App
	default:
		panic("deploy type did not register default configuration")
	}
}

func ExecutorsAccessToken() string {
	if deploy.IsApp() {
		return confdefaults.AppInMemoryExecutorPassword
	}
	return Get().ExecutorsAccessToken
}

func BitbucketServerConfigs(ctx context.Context) ([]*schema.BitbucketServerConnection, error) {
	var config []*schema.BitbucketServerConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindBitbucketServer, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitHubConfigs(ctx context.Context) ([]*schema.GitHubConnection, error) {
	var config []*schema.GitHubConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindGitHub, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitLabConfigs(ctx context.Context) ([]*schema.GitLabConnection, error) {
	var config []*schema.GitLabConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindGitLab, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitoliteConfigs(ctx context.Context) ([]*schema.GitoliteConnection, error) {
	var config []*schema.GitoliteConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindGitolite, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func PhabricatorConfigs(ctx context.Context) ([]*schema.PhabricatorConnection, error) {
	var config []*schema.PhabricatorConnection
	if err := internalapi.Client.ExternalServiceConfigs(ctx, extsvc.KindPhabricator, &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GitHubAppEnabled() bool {
	cfg, _ := GitHubAppConfig()
	return cfg.Configured()
}

type GitHubAppConfiguration struct {
	PrivateKey   []byte
	AppID        string
	Slug         string
	ClientID     string
	ClientSecret string
}

func (c GitHubAppConfiguration) Configured() bool {
	return c.AppID != "" && len(c.PrivateKey) != 0 && c.Slug != "" && c.ClientID != "" && c.ClientSecret != ""
}

func GitHubAppConfig() (config GitHubAppConfiguration, err error) {
	cfg := Get().GitHubApp
	if cfg == nil {
		return GitHubAppConfiguration{}, nil
	}

	privateKey, err := base64.StdEncoding.DecodeString(cfg.PrivateKey)
	if err != nil {
		return GitHubAppConfiguration{}, errors.Wrap(err, "decoding GitHub app private key failed")
	}
	return GitHubAppConfiguration{
		PrivateKey:   privateKey,
		AppID:        cfg.AppID,
		Slug:         cfg.Slug,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
	}, nil
}

type AccessTokenAllow string

const (
	AccessTokensNone  AccessTokenAllow = "none"
	AccessTokensAll   AccessTokenAllow = "all-users-create"
	AccessTokensAdmin AccessTokenAllow = "site-admin-create"
)

// AccessTokensAllow returns whether access tokens are enabled, disabled, or
// restricted creation to only site admins.
func AccessTokensAllow() AccessTokenAllow {
	cfg := Get().AuthAccessTokens
	if cfg == nil || cfg.Allow == "" {
		return AccessTokensAll
	}
	v := AccessTokenAllow(cfg.Allow)
	switch v {
	case AccessTokensAll, AccessTokensAdmin:
		return v
	default:
		return AccessTokensNone
	}
}

// EmailVerificationRequired returns whether users must verify an email address before they
// can perform most actions on this site.
//
// It's false for sites that do not have an email sending API key set up.
func EmailVerificationRequired() bool {
	return CanSendEmail()
}

// CanSendEmail returns whether the site can send emails (e.g., to reset a password or
// invite a user to an org).
//
// It's false for sites that do not have an email sending API key set up.
func CanSendEmail() bool {
	return Get().EmailSmtp != nil
}

// UpdateChannel tells the update channel. Default is "release".
func UpdateChannel() string {
	channel := Get().UpdateChannel
	if channel == "" {
		return "release"
	}
	return channel
}

func BatchChangesEnabled() bool {
	if enabled := Get().BatchChangesEnabled; enabled != nil {
		return *enabled
	}
	return true
}

func BatchChangesRestrictedToAdmins() bool {
	if restricted := Get().BatchChangesRestrictToAdmins; restricted != nil {
		return *restricted
	}
	return false
}

// CodyEnabled returns whether Cody is enabled on this instance.
//
// If `cody.enabled` is not set or set to false, it's not enabled.
// If `cody.enabled` is true but `completions` aren't set, it returns false.
//
// Legacy-support for `completions.enabled`:
// If `cody.enabled` is NOT set, but `completions.enabled` is set, then cody is enabled.
// If `cody.enabled` is set, but `completions.enabled` is set, then cody is enabled based on value of `cody.enabled`.
func CodyEnabled() bool {
	enabled := Get().CodyEnabled
	completions := Get().Completions

	// Support for Legacy configurations in which `completions` is set to
	// `enabled`, but `cody.enabled` is not set.
	if enabled == nil && completions != nil && completions.Enabled {
		return true
	}

	if enabled == nil {
		return false
	}

	return *enabled
}

func CodyRestrictUsersFeatureFlag() bool {
	if restrict := Get().CodyRestrictUsersFeatureFlag; restrict != nil {
		return *restrict
	}
	return false
}

func ExecutorsEnabled() bool {
	return Get().ExecutorsAccessToken != ""
}

func ExecutorsFrontendURL() string {
	current := Get()
	if current.ExecutorsFrontendURL != "" {
		return current.ExecutorsFrontendURL
	}

	return current.ExternalURL
}

func ExecutorsSrcCLIImage() string {
	current := Get()
	if current.ExecutorsSrcCLIImage != "" {
		return current.ExecutorsSrcCLIImage
	}

	return "sourcegraph/src-cli"
}

func ExecutorsSrcCLIImageTag() string {
	current := Get()
	if current.ExecutorsSrcCLIImageTag != "" {
		return current.ExecutorsSrcCLIImageTag
	}

	return srccli.MinimumVersion
}

func ExecutorsLsifGoImage() string {
	current := Get()
	if current.ExecutorsLsifGoImage != "" {
		return current.ExecutorsLsifGoImage
	}
	return "sourcegraph/lsif-go"
}

func ExecutorsBatcheshelperImage() string {
	current := Get()
	if current.ExecutorsBatcheshelperImage != "" {
		return current.ExecutorsBatcheshelperImage
	}

	return "sourcegraph/batcheshelper"
}

func ExecutorsBatcheshelperImageTag() string {
	current := Get()
	if current.ExecutorsBatcheshelperImageTag != "" {
		return current.ExecutorsBatcheshelperImageTag
	}

	if version.IsDev(version.Version()) {
		return "insiders"
	}

	return version.Version()
}

func CodeIntelAutoIndexingEnabled() bool {
	if enabled := Get().CodeIntelAutoIndexingEnabled; enabled != nil {
		return *enabled
	}
	return false
}

func CodeIntelAutoIndexingAllowGlobalPolicies() bool {
	if enabled := Get().CodeIntelAutoIndexingAllowGlobalPolicies; enabled != nil {
		return *enabled
	}
	return false
}

func CodeIntelAutoIndexingPolicyRepositoryMatchLimit() int {
	val := Get().CodeIntelAutoIndexingPolicyRepositoryMatchLimit
	if val == nil || *val < -1 {
		return -1
	}

	return *val
}

func CodeIntelRankingDocumentReferenceCountsEnabled() bool {
	if enabled := Get().CodeIntelRankingDocumentReferenceCountsEnabled; enabled != nil {
		return *enabled
	}
	return false
}

func CodeIntelRankingDocumentReferenceCountsCronExpression() (*cronexpr.Expression, error) {
	if cronExpression := Get().CodeIntelRankingDocumentReferenceCountsCronExpression; cronExpression != nil {
		return cronexpr.Parse(*cronExpression)
	}

	return cronexpr.Parse("@weekly")
}

func CodeIntelRankingDocumentReferenceCountsGraphKey() string {
	if val := Get().CodeIntelRankingDocumentReferenceCountsGraphKey; val != "" {
		return val
	}
	return "dev"
}

func CodeIntelRankingDocumentReferenceCountsDerivativeGraphKeyPrefix() string {
	if val := Get().CodeIntelRankingDocumentReferenceCountsDerivativeGraphKeyPrefix; val != "" {
		return val
	}
	return ""
}

func CodeIntelRankingStaleResultAge() time.Duration {
	if val := Get().CodeIntelRankingStaleResultsAge; val > 0 {
		return time.Duration(val) * time.Hour
	}
	return 24 * time.Hour
}

func EmbeddingsEnabled() bool {
	return GetEmbeddingsConfig(Get().SiteConfiguration) != nil
}

func ProductResearchPageEnabled() bool {
	if enabled := Get().ProductResearchPageEnabled; enabled != nil {
		return *enabled
	}
	return true
}

func ExternalURL() string {
	return Get().ExternalURL
}

func UsingExternalURL() bool {
	url := Get().ExternalURL
	return !(url == "" || strings.HasPrefix(url, "http://localhost") || strings.HasPrefix(url, "https://localhost") || strings.HasPrefix(url, "http://127.0.0.1") || strings.HasPrefix(url, "https://127.0.0.1")) // CI:LOCALHOST_OK
}

func IsExternalURLSecure() bool {
	return strings.HasPrefix(Get().ExternalURL, "https:")
}

func IsBuiltinSignupAllowed() bool {
	provs := Get().AuthProviders
	for _, prov := range provs {
		if prov.Builtin != nil {
			return prov.Builtin.AllowSignup
		}
	}
	return false
}

// IsAccessRequestEnabled returns whether request access experimental feature is enabled or not.
func IsAccessRequestEnabled() bool {
	authAccessRequest := Get().AuthAccessRequest
	return authAccessRequest == nil || authAccessRequest.Enabled == nil || *authAccessRequest.Enabled
}

// AuthPrimaryLoginProvidersCount returns the number of primary login providers
// configured, or 3 (the default) if not explicitly configured.
// This is only used for the UI
func AuthPrimaryLoginProvidersCount() int {
	c := Get().AuthPrimaryLoginProvidersCount
	if c == 0 {
		return 3 // default to 3
	}
	return c
}

// SearchSymbolsParallelism returns 20, or the site config
// "debug.search.symbolsParallelism" value if configured.
func SearchSymbolsParallelism() int {
	val := Get().DebugSearchSymbolsParallelism
	if val == 0 {
		return 20
	}
	return val
}

func EventLoggingEnabled() bool {
	val := ExperimentalFeatures().EventLogging
	if val == "" {
		return true
	}
	return val == "enabled"
}

func StructuralSearchEnabled() bool {
	val := ExperimentalFeatures().StructuralSearch
	if val == "" {
		return true
	}
	return val == "enabled"
}

// SearchDocumentRanksWeight controls the impact of document ranks on the final ranking when
// SearchOptions.UseDocumentRanks is enabled. The default is 0.5 * 9000 (half the zoekt default),
// to match existing behavior where ranks are given half the priority as existing scoring signals.
// We plan to eventually remove this, once we experiment on real data to find a good default.
func SearchDocumentRanksWeight() float64 {
	ranking := ExperimentalFeatures().Ranking
	if ranking != nil && ranking.DocumentRanksWeight != nil {
		return *ranking.DocumentRanksWeight
	} else {
		return 4500
	}
}

// SearchFlushWallTime controls the amount of time that Zoekt shards collect and rank results. For
// larger codebases, it can be helpful to increase this to improve the ranking stability and quality.
func SearchFlushWallTime() time.Duration {
	ranking := ExperimentalFeatures().Ranking
	if ranking != nil && ranking.FlushWallTimeMS > 0 {
		return time.Duration(ranking.FlushWallTimeMS) * time.Millisecond
	} else {
		return 500 * time.Millisecond
	}
}

func ExperimentalFeatures() schema.ExperimentalFeatures {
	val := Get().ExperimentalFeatures
	if val == nil {
		return schema.ExperimentalFeatures{}
	}
	return *val
}

func Tracer() string {
	ot := Get().ObservabilityTracing
	if ot == nil {
		return ""
	}
	return ot.Type
}

// AuthMinPasswordLength returns the value of minimum password length requirement.
// If not set, it returns the default value 12.
func AuthMinPasswordLength() int {
	val := Get().AuthMinPasswordLength
	if val <= 0 {
		return 12
	}
	return val
}

// GenericPasswordPolicy is a generic password policy that defines password requirements.
type GenericPasswordPolicy struct {
	Enabled                   bool
	MinimumLength             int
	NumberOfSpecialCharacters int
	RequireAtLeastOneNumber   bool
	RequireUpperandLowerCase  bool
}

// AuthPasswordPolicy returns a GenericPasswordPolicy for password validation
func AuthPasswordPolicy() GenericPasswordPolicy {
	ml := Get().AuthMinPasswordLength

	if p := Get().AuthPasswordPolicy; p != nil {
		return GenericPasswordPolicy{
			Enabled:                   p.Enabled,
			MinimumLength:             ml,
			NumberOfSpecialCharacters: p.NumberOfSpecialCharacters,
			RequireAtLeastOneNumber:   p.RequireAtLeastOneNumber,
			RequireUpperandLowerCase:  p.RequireUpperandLowerCase,
		}
	}

	if ep := ExperimentalFeatures().PasswordPolicy; ep != nil {
		return GenericPasswordPolicy{
			Enabled:                   ep.Enabled,
			MinimumLength:             ml,
			NumberOfSpecialCharacters: ep.NumberOfSpecialCharacters,
			RequireAtLeastOneNumber:   ep.RequireAtLeastOneNumber,
			RequireUpperandLowerCase:  ep.RequireUpperandLowerCase,
		}
	}

	return GenericPasswordPolicy{
		Enabled:                   false,
		MinimumLength:             0,
		NumberOfSpecialCharacters: 0,
		RequireAtLeastOneNumber:   false,
		RequireUpperandLowerCase:  false,
	}
}

func PasswordPolicyEnabled() bool {
	pc := AuthPasswordPolicy()
	return pc.Enabled
}

// By default, password reset links are valid for 4 hours.
const defaultPasswordLinkExpiry = 14400

// AuthPasswordResetLinkExpiry returns the time (in seconds) indicating how long password
// reset links are considered valid. If not set, it returns the default value.
func AuthPasswordResetLinkExpiry() int {
	val := Get().AuthPasswordResetLinkExpiry
	if val <= 0 {
		return defaultPasswordLinkExpiry
	}
	return val
}

// AuthLockout populates and returns the *schema.AuthLockout with default values
// for fields that are not initialized.
func AuthLockout() *schema.AuthLockout {
	val := Get().AuthLockout
	if val == nil {
		return &schema.AuthLockout{
			ConsecutivePeriod:      3600,
			FailedAttemptThreshold: 5,
			LockoutPeriod:          1800,
		}
	}

	if val.ConsecutivePeriod <= 0 {
		val.ConsecutivePeriod = 3600
	}
	if val.FailedAttemptThreshold <= 0 {
		val.FailedAttemptThreshold = 5
	}
	if val.LockoutPeriod <= 0 {
		val.LockoutPeriod = 1800
	}
	return val
}

type ExternalServiceMode int

const (
	ExternalServiceModeDisabled ExternalServiceMode = 0
	ExternalServiceModePublic   ExternalServiceMode = 1
	ExternalServiceModeAll      ExternalServiceMode = 2
)

func (e ExternalServiceMode) String() string {
	switch e {
	case ExternalServiceModeDisabled:
		return "disabled"
	case ExternalServiceModePublic:
		return "public"
	case ExternalServiceModeAll:
		return "all"
	default:
		return "unknown"
	}
}

// ExternalServiceUserMode returns the site level mode describing if users are
// allowed to add external services for public and private repositories. It does
// NOT take into account permissions granted to the current user.
func ExternalServiceUserMode() ExternalServiceMode {
	switch Get().ExternalServiceUserMode {
	case "public":
		return ExternalServiceModePublic
	case "all":
		return ExternalServiceModeAll
	default:
		return ExternalServiceModeDisabled
	}
}

const defaultGitLongCommandTimeout = time.Hour

// GitLongCommandTimeout returns the maximum amount of time in seconds that a
// long Git command (e.g. clone or remote update) is allowed to execute. If not
// set, it returns the default value.
//
// In general, Git commands that are expected to take a long time should be
// executed in the background in a non-blocking fashion.
func GitLongCommandTimeout() time.Duration {
	val := Get().GitLongCommandTimeout
	if val < 1 {
		return defaultGitLongCommandTimeout
	}
	return time.Duration(val) * time.Second
}

// GitMaxCodehostRequestsPerSecond returns maximum number of remote code host
// git operations to be run per second per gitserver. If not set, it returns the
// default value -1.
func GitMaxCodehostRequestsPerSecond() int {
	val := Get().GitMaxCodehostRequestsPerSecond
	if val == nil || *val < -1 {
		return -1
	}
	return *val
}

func GitMaxConcurrentClones() int {
	v := Get().GitMaxConcurrentClones
	if v <= 0 {
		return 5
	}
	return v
}

// GetCompletionsConfig evaluates a complete completions configuration based on
// site configuration. The configuration may be nil if completions is disabled.
func GetCompletionsConfig(siteConfig schema.SiteConfiguration) (c *conftypes.CompletionsConfig) {
	codyEnabled := siteConfig.CodyEnabled
	// If the cody.enabled flag is not set, or is explicitly false, no completions
	// should be used.
	if codyEnabled == nil || !*codyEnabled {
		// if siteConfig.Completions != nil &&
		return nil
	}
	// completionsConfig := siteConfig.Completions

	c.AccessToken = getSourcegraphProviderAccessToken(siteConfig.Completions.AccessToken, config)
	// If we weren't able to generate an access token of some sort, authing with
	// Cody Gateway is not possible and we cannot use completions.
	if c.AccessToken == "" {
		return nil
	}

	// If App is running and there wasn't a completions config
	// use a provider that sends the request to dotcom
	if deploy.IsApp() {
		// If someone explicitly turned Cody off, no config
		if codyEnabled != nil && !*codyEnabled {
			return nil
		}

		// If Cody is on or not explicitly turned off and no config, assume default
		if completionsConfig == nil {
			appConfig := Get().App
			if appConfig == nil {
				return nil
			}
			// Only the Provider, Access Token and Enabled required to forward the request to dotcom
			return &CompletionsConfig{
				AccessToken: appConfig.DotcomAuthToken,
				Provider:    CompletionsProviderNameDotcom,
				// TODO: These are not required right now as upstream overwrites this,
				// but should we switch to Cody Gateway they will be.
				ChatModel:       "claude-v1",
				FastChatModel:   "claude-instant-v1",
				CompletionModel: "claude-instant-v1",

				// Irrelevant for this provider.
				Endpoint:                         "",
				PerUserDailyLimit:                0,
				PerUserCodeCompletionsDailyLimit: 0,
			}
		}
	}

	// If `cody.enabled` is used but no completions config, we assume defaults
	if codyEnabled != nil && *codyEnabled {
		if siteConfig.Completions == nil {
			c = &CompletionsConfig{
				Provider:        CompletionsProviderNameSourcegraph,
				ChatModel:       "anthropic/claude-v1",
				FastChatModel:   "anthropic/claude-instant-v1",
				CompletionModel: "anthropic/claude-instant-v1",
				// TODO: Check for 0 length string.
				AccessToken:                      licensing.GenerateLicenseKeyBasedAccessToken(siteConfig.LicenseKey),
				Endpoint:                         "https://cody-gateway.sourcegraph.com",
				PerUserDailyLimit:                0,
				PerUserCodeCompletionsDailyLimit: 0,
			}
		}
	}

	// DEPRECATED: If the config is explicitly disabled, we return nil.
	if siteConfig.Completions != nil && !siteConfig.Completions.Enabled {
		return nil
	}

	// If the completions config is set, provider, ChatModel, CompletionModel, and AccessToken are required:
	// if siteConfig.Completions.pro

	// If a provider is not set, or if the provider is Cody Gateway, set up
	// magic defaults. Note that we do NOT enable completions for the user -
	// that still needs to be explicitly configured.
	if completionsConfig.Provider == "" || completionsConfig.Provider == string(CompletionsProviderNameSourcegraph) {
		// Set provider to Cody Gateway in case it's empty.
		completionsConfig.Provider = string(CompletionsProviderNameSourcegraph)

		// Configure accessToken. We don't validate the license here because
		// Cody Gateway will check and reject the request.
		if completionsConfig.AccessToken == "" && siteConfig.LicenseKey != "" {
			completionsConfig.AccessToken = licensing.GenerateLicenseKeyBasedAccessToken(siteConfig.LicenseKey)
		}

		// Configure endpoint
		if completionsConfig.Endpoint == "" {
			completionsConfig.Endpoint = codygatewayDefaultEndpoint
		}
		// Configure chatModel
		if completionsConfig.ChatModel == "" {
			completionsConfig.ChatModel = "anthropic/claude-v1"
		}
		// Configure completionModel
		if completionsConfig.CompletionModel == "" {
			completionsConfig.CompletionModel = "anthropic/claude-instant-v1"
		}
		// Configure fastChatModel
		if completionsConfig.FastChatModel == "" {
			completionsConfig.FastChatModel = completionsConfig.CompletionModel
		}

		// NOTE: We explicitly aren't adding back-compat for completions.model
		// because Cody Gateway disallows the use of most chat models for
		// code completions, so in most cases the back-compat wouldn't work
		// anyway.

		return completionsConfig
	}

	if completionsConfig.ChatModel == "" {
		// If no model for chat is configured, nothing we can do.
		if completionsConfig.Model == "" {
			return nil
		}
		completionsConfig.ChatModel = completionsConfig.Model
	}

	// TODO: Temporary workaround to fix instances where no completion model is set.
	if completionsConfig.CompletionModel == "" {
		completionsConfig.CompletionModel = "claude-instant-v1"
	}

	if completionsConfig.Endpoint == "" && completionsConfig.Provider == string(CompletionsProviderNameOpenAI) {
		completionsConfig.Endpoint = "https://api.openai.com/v1/chat/completions"
	}
	if completionsConfig.Endpoint == "" && completionsConfig.Provider == string(CompletionsProviderNameAnthropic) {
		completionsConfig.Endpoint = "https://api.anthropic.com/v1/complete"
	}

	if completionsConfig.FastChatModel == "" {
		completionsConfig.FastChatModel = completionsConfig.CompletionModel
	}

	return completionsConfig
}

const codygatewayDefaultEndpoint = "https://cody-gateway.sourcegraph.com"

// GetEmbeddingsConfig evaluates a complete embeddings configuration based on
// site configuration. The configuration may be nil if completions is disabled.
func GetEmbeddingsConfig(siteConfig schema.SiteConfiguration) *conftypes.EmbeddingsConfig {
	// If cody is explicitly disabled, don't use embeddings.
	if siteConfig.CodyEnabled != nil && !*siteConfig.CodyEnabled {
		return nil
	}

	// Additionally Embeddings in App are disabled if there is no dotcom auth token
	// and the user hasn't provided their own api token.
	if deploy.IsApp() {
		if (siteConfig.App == nil || len(siteConfig.App.DotcomAuthToken) == 0) && (siteConfig.Embeddings == nil || siteConfig.Embeddings.AccessToken == "") {
			return nil
		}
	}

	// If embeddings are explicitly disabled (legacy flag, TODO: remove after 5.1),
	// don't use embeddings either.
	if siteConfig.Embeddings != nil && siteConfig.Embeddings.Enabled != nil && !*siteConfig.Embeddings.Enabled {
		return nil
	}

	embeddingsConfig := siteConfig.Embeddings
	// If no embeddings configuration is set at all, but cody is enabled, assume
	// a default configuration.
	if embeddingsConfig == nil {
		embeddingsConfig = &schema.Embeddings{
			Dimensions:      1536,
			Enabled:         pointify(true),
			Provider:        string(conftypes.EmbeddingsProviderNameSourcegraph),
			Incremental:     pointify(true),
			Model:           "openai/text-embedding-ada-002",
			MinimumInterval: "24h",
		}
	}

	if embeddingsConfig.Provider == string(conftypes.EmbeddingsProviderNameSourcegraph) || embeddingsConfig.Provider == "" {
		embeddingsConfig.Provider = string(conftypes.EmbeddingsProviderNameSourcegraph)

		// Fallback to URL, it's the previous name of the setting.
		if embeddingsConfig.Endpoint == "" {
			embeddingsConfig.Endpoint = embeddingsConfig.Url
		}
		// If that is also not set, use a sensible default.
		if embeddingsConfig.Endpoint == "" {
			embeddingsConfig.Endpoint = "https://cody-gateway.sourcegraph.com/v1/embeddings"
		}

		embeddingsConfig.AccessToken = getSourcegraphProviderAccessToken(config.Embeddings.AccessToken, config)
		// If we weren't able to generate an access token of some sort, authing with
		// Cody Gateway is not possible and we cannot use embeddings.
		if embeddingsConfig.AccessToken == "" {
			return nil
		}

		if embeddingsConfig.Model == "" {
			embeddingsConfig.Model = "openai/text-embedding-ada-002"
		}
		if embeddingsConfig.Dimensions == 0 && embeddingsConfig.Model == "openai/text-embedding-ada-002" {
			embeddingsConfig.Dimensions = 1536
		}
	} else if embeddingsConfig.Provider == string(conftypes.EmbeddingsProviderNameOpenAI) {
		// Fallback to URL, it's the previous name of the setting.
		if embeddingsConfig.Endpoint == "" {
			embeddingsConfig.Endpoint = embeddingsConfig.Url
		}
		// If that is also not set, use a sensible default.
		if embeddingsConfig.Endpoint == "" {
			embeddingsConfig.Endpoint = "https://api.openai.com/v1/embeddings"
		}

		if embeddingsConfig.Model == "" {
			embeddingsConfig.Model = "text-embedding-ada-002"
		}
		if embeddingsConfig.Dimensions == 0 && embeddingsConfig.Model == "text-embedding-ada-002" {
			embeddingsConfig.Dimensions = 1536
		}
	}

	defaultMinimumInterval := 24 * time.Hour
	minimumIntervalString := embeddingsConfig.MinimumInterval
	d, err := time.ParseDuration(minimumIntervalString)
	if err != nil {
		embeddingsConfig.MinimumInterval = defaultMinimumInterval
	} else {
		embeddingsConfig.MinimumInterval = d
	}

	defaultMaxCodeEmbeddingsPerRepo := 3_072_000
	defaultMaxTextEmbeddingsPerRepo := 512_000
	embeddingsConfig.MaxCodeEmbeddingsPerRepo = defaultTo(embeddingsConfig.MaxCodeEmbeddingsPerRepo, defaultMaxCodeEmbeddingsPerRepo)
	embeddingsConfig.MaxCodeEmbeddingsPerRepo = defaultTo(embeddingsConfig.MaxTextEmbeddingsPerRepo, defaultMaxTextEmbeddingsPerRepo)
	embeddingsConfig.PolicyRepositoryMatchLimit = embeddingsConfig.PolicyRepositoryMatchLimit
	if embeddingsConfig.Model == "" {
		embeddingsConfig.Model = "text-embedding-ada-002"
	}
	embeddingsConfig.Model = strings.ToLower(embeddingsConfig.Model)

	return embeddingsConfig
}

func getSourcegraphProviderAccessToken(accessToken string, config *schema.SiteConfiguration) string {
	// If an access token is configured, use it.
	if accessToken != "" {
		return accessToken
	}
	// App generates a token from the api token the user used to connect app to dotcom.
	if deploy.IsApp() && config.App != nil {
		if config.App.DotcomAuthToken == "" {
			return ""
		}
		return dotcomuser.GenerateDotcomUserGatewayAccessToken(config.App.DotcomAuthToken)
	}
	// Otherwise, use the current license key to compute an access token.
	if config.LicenseKey == "" {
		return ""
	}
	return licensing.GenerateLicenseKeyBasedAccessToken(config.LicenseKey)
}

func defaultTo(val, def int) int {
	if val == 0 {
		return def
	}
	return val
}

func pointify[T any](v T) *T {
	return &v
}
