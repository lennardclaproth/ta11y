package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

const configPath = "config.yaml"

type Config struct {
	Server      Server      `yaml:"server"`
	Database    Database    `yaml:"database"`
	Logging     Logging     `yaml:"logging"`
	APM         APMConfig   `yaml:"apm"`
	Auth        Auth        `yaml:"auth"`
	DiskStorage DiskStorage `yaml:"disk_storage"`
	Providers   Providers   `yaml:"providers"`
}

// Auth configures OpenID Connect sign-in. Client credentials are never read from
// config.yaml -- each provider's id and secret come from the environment, named
// <SLUG>_CLIENT_ID / <SLUG>_CLIENT_SECRET (e.g. GOOGLE_CLIENT_ID).
type Auth struct {
	// Enabled turns sign-in on. With it off the API boots without contacting any
	// provider, which keeps local development and the test suite self-contained.
	Enabled bool `yaml:"enabled"`
	// SessionTTL is how long a session stays valid. Zero means auth.DefaultSessionTTL.
	SessionTTL time.Duration `yaml:"session_ttl"`
	// FrontendURL is where the browser is sent once a login completes, and the default
	// origin for the session cookie. It must be an origin the frontend is served from.
	FrontendURL string `yaml:"frontend_url"`
	// CookieDomain scopes the session cookie. Empty leaves it host-only, which is
	// correct whenever the API and frontend share a host.
	CookieDomain string `yaml:"cookie_domain"`
	// CookieSecure forces the Secure attribute. It defaults to true outside development
	// and must stay true wherever the site is served over HTTPS.
	CookieSecure *bool `yaml:"cookie_secure"`
	// AllowedEmails restricts who may sign in. Empty means anyone the provider
	// authenticates, which for a public issuer such as Google means anyone at all --
	// so it is required outside development. Compared case-insensitively.
	AllowedEmails []string `yaml:"allowed_emails"`
	// BootstrapAdminEmail claims the seeded account on first sign-in, so pre-existing
	// data is adopted rather than orphaned behind a freshly provisioned account. Read
	// from AUTH_BOOTSTRAP_ADMIN_EMAIL when absent here.
	BootstrapAdminEmail string `yaml:"bootstrap_admin_email"`
	// Providers lists the identity providers to discover at start-up, keyed by the slug
	// that appears in /auth/{provider}/login.
	Providers map[string]AuthProvider `yaml:"providers"`
}

// AuthProvider is one OpenID Connect identity provider.
type AuthProvider struct {
	// Issuer is the provider's issuer URL; everything else is fetched by discovery.
	Issuer string `yaml:"issuer"`
	// RedirectURL must match the redirect registered with the provider exactly.
	RedirectURL string `yaml:"redirect_url"`
	// ClientID and ClientSecret are populated from the environment, not from yaml.
	ClientID     string `yaml:"-"`
	ClientSecret string `yaml:"-"`
}

// IsCookieSecure reports whether the session cookie carries the Secure attribute,
// defaulting to true anywhere but development so a misconfigured production never
// downgrades by omission.
func (a Auth) IsCookieSecure(environment string) bool {
	if a.CookieSecure != nil {
		return *a.CookieSecure
	}
	return !strings.EqualFold(strings.TrimSpace(environment), "development")
}

type DiskStorage struct {
	BasePath string `yaml:"base_path"`
}

type Logging struct {
	Level string `yaml:"level"`
}

type Server struct {
	Environment string `yaml:"environment"`
	Port        int    `yaml:"port"`
	// CORSAllowedOrigins lists the browser origins permitted to call the API
	// cross-origin (e.g. the SvelteKit dev server). Empty falls back to the
	// local dev origin; a single "*" entry allows any origin.
	CORSAllowedOrigins []string `yaml:"cors_allowed_origins"`
}

// AllowedOrigins returns the configured CORS origins, defaulting to the local
// SvelteKit dev server when none are set.
func (s Server) AllowedOrigins() []string {
	if len(s.CORSAllowedOrigins) == 0 {
		return []string{"http://localhost:5199"}
	}
	return s.CORSAllowedOrigins
}

type Database struct {
	ConnStr string `yaml:"connection_string"`
	Type    string `yaml:"type"`
}

type APMConfig struct {
	ServerURL             string  `yaml:"server_url"`
	ServiceName           string  `yaml:"service_name"`
	Environment           string  `yaml:"environment"`
	SecretToken           string  `yaml:"secret_token"`
	VerifyServerCert      bool    `yaml:"verify_server_cert"`
	LogLevel              string  `yaml:"log_level"`
	TransactionSampleRate float64 `yaml:"transaction_sample_rate"`
}

type Providers struct {
	MarketStack  ProviderConfig `yaml:"marketstack"`
	AlphaVantage ProviderConfig `yaml:"alphavantage"`
}

type ProviderConfig struct {
	BaseURI string   `yaml:"base_uri"`
	APIKeys []string `yaml:"-"`
}

func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}
	if c.Database.ConnStr == "" {
		return fmt.Errorf("database connection string cannot be empty")
	}
	if c.Database.Type != "sqlite3" && c.Database.Type != "postgres" {
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}
	if c.Logging.Level != "debug" && c.Logging.Level != "info" && c.Logging.Level != "warn" && c.Logging.Level != "error" {
		return fmt.Errorf("invalid logging level: %s", c.Logging.Level)
	}
	if c.APM.ServerURL == "" {
		return fmt.Errorf("APM server URL cannot be empty")
	}
	if c.APM.ServiceName == "" {
		return fmt.Errorf("APM service name cannot be empty")
	}
	if c.APM.LogLevel != "debug" && c.APM.LogLevel != "info" && c.APM.LogLevel != "warn" && c.APM.LogLevel != "error" {
		return fmt.Errorf("invalid APM log level: %s", c.APM.LogLevel)
	}
	if c.APM.TransactionSampleRate < 0 || c.APM.TransactionSampleRate > 1 {
		return fmt.Errorf("APM transaction sample rate must be between 0 and 1")
	}
	if c.DiskStorage.BasePath == "" {
		return fmt.Errorf("disk storage base path cannot be empty")
	}
	if err := c.Auth.validate(); err != nil {
		return err
	}
	if err := c.validateAuthPosture(); err != nil {
		return err
	}
	return nil
}

// validateAuthPosture refuses the two configurations that would quietly expose data
// outside development: running with sign-in off, and running with sign-in on but no
// allowlist, which for a public issuer lets anyone with an account at that provider in.
func (c *Config) validateAuthPosture() error {
	if isDevelopment(c.Server.Environment) {
		return nil
	}
	if !c.Auth.Enabled {
		return fmt.Errorf("auth must be enabled outside development: every request would run as the bootstrapped account")
	}
	if len(c.Auth.AllowedEmails) == 0 {
		return fmt.Errorf("auth.allowed_emails must list at least one address outside development: an empty allowlist lets anyone with an account at the identity provider sign in")
	}
	return nil
}

// isDevelopment reports whether an environment may run without authentication. The
// list is an allowlist rather than a "not production" check: an unfamiliar name such as
// "staging" is treated as production-like, so a new environment is strict by default
// instead of silently unauthenticated.
func isDevelopment(environment string) bool {
	switch strings.ToLower(strings.TrimSpace(environment)) {
	case "", "development", "dev", "local", "test", "ci":
		return true
	default:
		return false
	}
}

// IsEmailAllowed reports whether an address may sign in. An empty allowlist admits
// everyone, which validate() permits only in development.
func (a Auth) IsEmailAllowed(email string) bool {
	if len(a.AllowedEmails) == 0 {
		return true
	}
	email = strings.ToLower(strings.TrimSpace(email))
	for _, allowed := range a.AllowedEmails {
		if strings.EqualFold(strings.TrimSpace(allowed), email) {
			return true
		}
	}
	return false
}

func (a Auth) validate() error {
	if !a.Enabled {
		return nil
	}
	if len(a.Providers) == 0 {
		return fmt.Errorf("auth is enabled but no identity providers are configured")
	}
	if strings.TrimSpace(a.FrontendURL) == "" {
		return fmt.Errorf("auth frontend_url is required when auth is enabled")
	}
	for slug, provider := range a.Providers {
		if strings.TrimSpace(provider.Issuer) == "" {
			return fmt.Errorf("auth provider %q is missing an issuer", slug)
		}
		if strings.TrimSpace(provider.RedirectURL) == "" {
			return fmt.Errorf("auth provider %q is missing a redirect_url", slug)
		}
		if provider.ClientID == "" || provider.ClientSecret == "" {
			return fmt.Errorf(
				"auth provider %q is missing credentials: set %s_CLIENT_ID and %s_CLIENT_SECRET",
				slug, envPrefix(slug), envPrefix(slug),
			)
		}
	}
	return nil
}

// hydrateAuthEnv fills each provider's credentials from the environment and applies the
// bootstrap admin email, keeping secrets out of config.yaml entirely.
func (c *Config) hydrateAuthEnv() {
	if c.Auth.BootstrapAdminEmail == "" {
		c.Auth.BootstrapAdminEmail = strings.TrimSpace(os.Getenv("AUTH_BOOTSTRAP_ADMIN_EMAIL"))
	}
	for slug, provider := range c.Auth.Providers {
		prefix := envPrefix(slug)
		provider.ClientID = strings.TrimSpace(os.Getenv(prefix + "_CLIENT_ID"))
		provider.ClientSecret = strings.TrimSpace(os.Getenv(prefix + "_CLIENT_SECRET"))
		c.Auth.Providers[slug] = provider
	}
}

// envPrefix maps a provider slug to its environment-variable prefix ("google" ->
// "GOOGLE", "entra-id" -> "ENTRA_ID").
func envPrefix(slug string) string {
	return strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(strings.TrimSpace(slug)))
}

func ReadConfig() (*Config, error) {
	f, err := os.ReadFile(configPath)

	if err != nil {
		return nil, fmt.Errorf("config: error opening config file at %s: %w", configPath, err)
	}

	var cfg Config

	if err := yaml.Unmarshal(f, &cfg); err != nil {
		return nil, fmt.Errorf("config: error decoding config: %w", err)
	}

	cfg.hydrateProviderEnv()
	cfg.hydrateAuthEnv()
	cfg.applyAPMDefaults()

	apmEnv := []struct {
		key   string
		value string
	}{
		{"ELASTIC_APM_SERVER_URL", cfg.APM.ServerURL},
		{"ELASTIC_APM_SERVICE_NAME", cfg.APM.ServiceName},
		{"ELASTIC_APM_ENVIRONMENT", cfg.APM.Environment},
		{"ELASTIC_APM_SECRET_TOKEN", cfg.APM.SecretToken},
		{"ELASTIC_APM_VERIFY_SERVER_CERT", strconv.FormatBool(cfg.APM.VerifyServerCert)},
		{"ELASTIC_APM_LOG_LEVEL", cfg.APM.LogLevel},
		{"ELASTIC_APM_TRANSACTION_SAMPLE_RATE", strconv.FormatFloat(cfg.APM.TransactionSampleRate, 'f', 2, 64)},
	}
	for _, env := range apmEnv {
		if err := os.Setenv(env.key, env.value); err != nil {
			return nil, fmt.Errorf("config: failed setting %s: %w", env.key, err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) hydrateProviderEnv() {
	marketStackBaseURI := strings.TrimSpace(c.Providers.MarketStack.BaseURI)
	if marketStackBaseURI == "" {
		marketStackBaseURI = strings.TrimSpace(os.Getenv("MARKETSTACK_BASE_URI"))
	}
	if marketStackBaseURI == "" {
		marketStackBaseURI = "https://api.marketstack.com/v2"
	}
	c.Providers.MarketStack.BaseURI = marketStackBaseURI
	c.Providers.MarketStack.APIKeys = splitAndDedupeCommaValues(os.Getenv("MARKETSTACK_API_KEY"))

	alphaVantageBaseURI := strings.TrimSpace(c.Providers.AlphaVantage.BaseURI)
	if alphaVantageBaseURI == "" {
		alphaVantageBaseURI = strings.TrimSpace(os.Getenv("ALPHA_VANTAGE_BASE_URI"))
	}
	if alphaVantageBaseURI == "" {
		alphaVantageBaseURI = "https://www.alphavantage.co"
	}
	c.Providers.AlphaVantage.BaseURI = alphaVantageBaseURI
	c.Providers.AlphaVantage.APIKeys = splitAndDedupeCommaValues(os.Getenv("ALPHA_VANTAGE_API_KEY"))
}

func splitAndDedupeCommaValues(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))

	for _, part := range parts {
		v := strings.TrimSpace(part)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	return out
}

func (c *Config) applyAPMDefaults() {
	if raw := strings.TrimSpace(os.Getenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE")); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			c.APM.TransactionSampleRate = parsed
			return
		}
	}

	if c.APM.TransactionSampleRate > 0 {
		return
	}

	environment := strings.ToLower(strings.TrimSpace(c.Server.Environment))
	if environment == "" {
		environment = strings.ToLower(strings.TrimSpace(c.APM.Environment))
	}

	switch environment {
	case "prod", "production":
		c.APM.TransactionSampleRate = 0.2
	default:
		c.APM.TransactionSampleRate = 1.0
	}
}

func (l *Logging) GetLogLevel() slog.Leveler {
	switch l.Level {
	case "debug":
		return slog.LevelDebug // Debug
	case "info":
		return slog.LevelInfo // Info
	case "warn":
		return slog.LevelWarn // Warn
	case "error":
		return slog.LevelError // Error
	default:
		return slog.LevelInfo // Default to Info
	}
}
