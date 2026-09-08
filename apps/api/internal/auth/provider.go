package auth

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// scopeEmail is the only profile scope requested. `profile` is deliberately omitted:
// the application has no use for a name or picture, so it does not ask for them.
var requestedScopes = []string{oidc.ScopeOpenID, "email"}

// ProviderConfig describes one OpenID Connect identity provider. Every provider is
// resolved through discovery, so adding one is configuration rather than code.
type ProviderConfig struct {
	// Slug identifies the provider in URLs, e.g. "google" in /auth/google/login.
	Slug string
	// Issuer is the provider's issuer URL, used for discovery and token validation.
	Issuer string
	// ClientID and ClientSecret come from the environment, never from config.yaml.
	ClientID     string
	ClientSecret string
	// RedirectURL must match the redirect registered with the provider exactly.
	RedirectURL string
}

// Provider is a configured, discovered identity provider ready to run a login.
type Provider struct {
	slug     string
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    *oauth2.Config
}

// Slug returns the provider's URL identifier.
func (p *Provider) Slug() string { return p.slug }

// Registry holds the configured identity providers by slug.
type Registry struct {
	providers map[string]*Provider
}

// NewRegistry discovers each configured provider and returns a registry. Discovery
// reaches the network, so this belongs in start-up rather than in a request path; a
// provider that cannot be discovered fails the boot rather than failing at first login.
func NewRegistry(ctx context.Context, configs []ProviderConfig) (*Registry, error) {
	registry := &Registry{providers: make(map[string]*Provider, len(configs))}

	for _, cfg := range configs {
		slug := strings.ToLower(strings.TrimSpace(cfg.Slug))
		if slug == "" {
			return nil, fmt.Errorf("auth: identity provider is missing a slug")
		}
		if cfg.ClientID == "" || cfg.ClientSecret == "" {
			return nil, fmt.Errorf("auth: identity provider %q is missing client credentials", slug)
		}

		provider, err := oidc.NewProvider(ctx, cfg.Issuer)
		if err != nil {
			return nil, fmt.Errorf("auth: discover identity provider %q: %w", slug, err)
		}

		registry.providers[slug] = &Provider{
			slug:     slug,
			provider: provider,
			verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
			oauth: &oauth2.Config{
				ClientID:     cfg.ClientID,
				ClientSecret: cfg.ClientSecret,
				Endpoint:     provider.Endpoint(),
				RedirectURL:  cfg.RedirectURL,
				Scopes:       requestedScopes,
			},
		}
	}

	return registry, nil
}

// NewEmptyRegistry returns a registry with no providers, which every lookup answers
// with ErrProviderUnknown. It is what running with authentication disabled uses.
func NewEmptyRegistry() *Registry {
	return &Registry{providers: make(map[string]*Provider)}
}

// Get returns the provider registered under slug, or ErrProviderUnknown.
func (r *Registry) Get(slug string) (*Provider, error) {
	provider, ok := r.providers[strings.ToLower(strings.TrimSpace(slug))]
	if !ok {
		return nil, ErrProviderUnknown
	}
	return provider, nil
}

// Slugs returns the configured provider slugs in a stable order, so clients can render
// a sign-in button per provider without hard-coding the list.
func (r *Registry) Slugs() []string {
	slugs := make([]string, 0, len(r.providers))
	for slug := range r.providers {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	return slugs
}

// AuthCodeURL builds the provider's authorization URL for a login attempt, binding it
// to the flow's state, nonce and PKCE challenge.
func (p *Provider) AuthCodeURL(flow *Flow) string {
	return p.oauth.AuthCodeURL(
		flow.State,
		oidc.Nonce(flow.Nonce),
		oauth2.S256ChallengeOption(flow.CodeVerifier),
	)
}

// Exchange trades an authorization code for tokens and verifies the returned ID token,
// checking its signature, issuer, audience, expiry and nonce. Only the claims this
// application uses are returned; the tokens themselves are discarded here and never
// leave this function.
func (p *Provider) Exchange(ctx context.Context, code string, flow *Flow) (*Claims, error) {
	token, err := p.oauth.Exchange(ctx, code, oauth2.VerifierOption(flow.CodeVerifier))
	if err != nil {
		return nil, fmt.Errorf("auth: exchange authorization code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("auth: provider %q returned no id_token", p.slug)
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("auth: verify id token: %w", err)
	}
	if idToken.Nonce != flow.Nonce {
		return nil, ErrFlowStateMismatch
	}

	var payload struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&payload); err != nil {
		return nil, fmt.Errorf("auth: decode id token claims: %w", err)
	}

	return &Claims{
		Issuer:        idToken.Issuer,
		Subject:       idToken.Subject,
		Email:         NormalizeEmail(payload.Email),
		EmailVerified: payload.EmailVerified,
	}, nil
}
