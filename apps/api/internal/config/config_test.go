package config

import (
	"testing"
)

func TestApplyAPMDefaults_ProductionDefault(t *testing.T) {
	t.Setenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", "")

	cfg := &Config{
		Server: Server{Environment: "production"},
		APM:    APMConfig{TransactionSampleRate: 0},
	}
	cfg.applyAPMDefaults()
	if cfg.APM.TransactionSampleRate != 0.2 {
		t.Fatalf("expected production sample rate 0.2, got %v", cfg.APM.TransactionSampleRate)
	}
}

func TestApplyAPMDefaults_DevDefault(t *testing.T) {
	t.Setenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", "")

	cfg := &Config{
		Server: Server{Environment: "development"},
		APM:    APMConfig{TransactionSampleRate: 0},
	}
	cfg.applyAPMDefaults()
	if cfg.APM.TransactionSampleRate != 1.0 {
		t.Fatalf("expected development sample rate 1.0, got %v", cfg.APM.TransactionSampleRate)
	}
}

func TestApplyAPMDefaults_EnvOverride(t *testing.T) {
	t.Setenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", "0.42")

	cfg := &Config{
		Server: Server{Environment: "production"},
		APM:    APMConfig{TransactionSampleRate: 0},
	}
	cfg.applyAPMDefaults()
	if cfg.APM.TransactionSampleRate != 0.42 {
		t.Fatalf("expected env override sample rate 0.42, got %v", cfg.APM.TransactionSampleRate)
	}
}

func TestApplyAPMDefaults_ExistingConfigValueWins(t *testing.T) {
	t.Setenv("ELASTIC_APM_TRANSACTION_SAMPLE_RATE", "")

	cfg := &Config{
		Server: Server{Environment: "production"},
		APM:    APMConfig{TransactionSampleRate: 0.75},
	}
	cfg.applyAPMDefaults()
	if cfg.APM.TransactionSampleRate != 0.75 {
		t.Fatalf("expected existing config value 0.75, got %v", cfg.APM.TransactionSampleRate)
	}
}

func validConfigForValidation() Config {
	return Config{
		Server: Server{
			Environment: "development",
			Port:        6060,
		},
		Database: Database{
			ConnStr: "sqlite://tmp.db",
			Type:    "sqlite3",
		},
		Logging: Logging{
			Level: "info",
		},
		APM: APMConfig{
			ServerURL:             "http://localhost:8200",
			ServiceName:           "ta11y",
			Environment:           "development",
			LogLevel:              "info",
			TransactionSampleRate: 1,
		},
		DiskStorage: DiskStorage{
			BasePath: "C:\\ta11y",
		},
	}
}

func TestValidate_AcceptsValidConfig(t *testing.T) {
	cfg := validConfigForValidation()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config to pass validation, got %v", err)
	}
}

func TestServerAddr(t *testing.T) {
	cases := []struct {
		host string
		want string
	}{
		{host: "", want: ":6060"},
		{host: "  ", want: ":6060"},
		{host: "127.0.0.1", want: "127.0.0.1:6060"},
		{host: "localhost", want: "localhost:6060"},
		{host: "::1", want: "[::1]:6060"},
	}
	for _, tc := range cases {
		got := Server{Host: tc.host, Port: 6060}.Addr()
		if got != tc.want {
			t.Errorf("Addr() with host %q = %q, want %q", tc.host, got, tc.want)
		}
	}
}

func TestValidate_ServerHost(t *testing.T) {
	for _, host := range []string{"", "127.0.0.1", "0.0.0.0", "::1", "localhost"} {
		cfg := validConfigForValidation()
		cfg.Server.Host = host
		if err := cfg.Validate(); err != nil {
			t.Errorf("host %q: expected valid, got %v", host, err)
		}
	}
	for _, host := range []string{"example.com", "127.0.0.1:6060", "not a host"} {
		cfg := validConfigForValidation()
		cfg.Server.Host = host
		if err := cfg.Validate(); err == nil {
			t.Errorf("host %q: expected a validation error", host)
		}
	}
}

func TestApplyEnvOverrides_EnvWinsOverFile(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://env/ta11y")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com, https://app.example.com,")
	t.Setenv("AUTH_ALLOWED_EMAILS", "a@example.com,b@example.com")
	t.Setenv("GOOGLE_REDIRECT_URL", "https://api.example.com/auth/google/callback")
	t.Setenv("MARKETDATA_MAX_SEED_PAGES", "10")
	t.Setenv("LOG_LEVEL", "")

	cfg := validConfigForValidation()
	cfg.Auth.Providers = map[string]AuthProvider{"google": {RedirectURL: "http://localhost:6060/auth/google/callback"}}
	if err := cfg.applyEnvOverrides(); err != nil {
		t.Fatalf("applyEnvOverrides: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Database.ConnStr != "postgres://env/ta11y" {
		t.Errorf("conn str = %q", cfg.Database.ConnStr)
	}
	if len(cfg.Server.CORSAllowedOrigins) != 1 || cfg.Server.CORSAllowedOrigins[0] != "https://app.example.com" {
		t.Errorf("cors origins = %v", cfg.Server.CORSAllowedOrigins)
	}
	if len(cfg.Auth.AllowedEmails) != 2 {
		t.Errorf("allowed emails = %v", cfg.Auth.AllowedEmails)
	}
	if got := cfg.Auth.Providers["google"].RedirectURL; got != "https://api.example.com/auth/google/callback" {
		t.Errorf("redirect url = %q", got)
	}
	if cfg.MarketData.MaxSeedPages != 10 {
		t.Errorf("max seed pages = %d, want 10", cfg.MarketData.MaxSeedPages)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("blank LOG_LEVEL must keep the file value, got %q", cfg.Logging.Level)
	}
}

func TestApplyEnvOverrides_RejectsNonNumericPort(t *testing.T) {
	t.Setenv("PORT", "eighty")

	cfg := validConfigForValidation()
	if err := cfg.applyEnvOverrides(); err == nil {
		t.Fatal("expected an error for a non-numeric PORT")
	}
}

func TestValidateAuthPosture(t *testing.T) {
	strict := func(env string, origins ...string) *Config {
		return &Config{
			Server: Server{Environment: env, CORSAllowedOrigins: origins},
			Auth:   Auth{Enabled: true, AllowedEmails: []string{"owner@example.com"}},
		}
	}

	if err := strict("production", "https://app.example.com").validateAuthPosture(); err != nil {
		t.Fatalf("production with an explicit origin: expected valid, got %v", err)
	}
	if err := strict("production", "*").validateAuthPosture(); err == nil {
		t.Fatal(`production with "*" origins: expected an error`)
	}
	if err := strict("development", "*").validateAuthPosture(); err != nil {
		t.Fatalf(`development with "*" origins: expected valid, got %v`, err)
	}

	unset := strict("")
	unset.Auth.Enabled = false
	if err := unset.validateAuthPosture(); err == nil {
		t.Fatal("unset environment with auth off: expected an error")
	}
}
