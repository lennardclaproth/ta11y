package marketdata

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// apiKeyHintLength is how many trailing characters of a key a hint keeps. Four is
// enough to tell two configured keys apart without meaningfully narrowing the secret.
const apiKeyHintLength = 4

// ProviderCredential describes a provider's connection record with the secret
// withheld. It is what read paths return: the stored key never leaves the process
// unless a caller explicitly asks to reveal it.
type ProviderCredential struct {
	ID            uuid.UUID
	Name          ProviderName
	IngestionMode ProviderIngestionMode
	BaseURI       *string
	// HasAPIKey reports whether a key is configured, which a hint alone cannot
	// distinguish from an empty one.
	HasAPIKey bool
	// APIKeyHint identifies the configured key without disclosing it.
	APIKeyHint string
	Remaining  int
	Used       int
	Total      int
	ResetsAt   *string
}

// MaskAPIKey renders a key as a hint that identifies it without disclosing it.
// Keys short enough that a suffix would give most of them away are fully masked.
func MaskAPIKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= apiKeyHintLength*2 {
		return strings.Repeat("*", len(trimmed))
	}
	return strings.Repeat("*", 4) + trimmed[len(trimmed)-apiKeyHintLength:]
}

// ToCredential projects a provider into its secret-free view.
func (p *Provider) ToCredential() *ProviderCredential {
	if p == nil {
		return nil
	}
	key := ""
	if p.ApiKey != nil {
		key = strings.TrimSpace(*p.ApiKey)
	}
	return &ProviderCredential{
		ID:            p.ID,
		Name:          p.Name,
		IngestionMode: p.IngestionMode,
		BaseURI:       p.BaseURI,
		HasAPIKey:     key != "",
		APIKeyHint:    MaskAPIKey(key),
		Remaining:     p.Remaining,
		Used:          p.Used,
		Total:         p.Total,
		ResetsAt:      p.ResetsAt,
	}
}

// CredentialStore reads and updates provider connection records.
type CredentialStore interface {
	ListProviders(ctx context.Context) ([]*Provider, error)
	GetProviderByID(ctx context.Context, id uuid.UUID) (*Provider, error)
	UpdateProviderCredentials(ctx context.Context, id uuid.UUID, apiKey, baseURI *string) error
}

// Credentials manages the API keys and base URIs of external market-data providers.
//
// Manual providers ingest from uploaded files and have no endpoint or key to store,
// so they are listed for completeness but reject credential updates.
type Credentials struct {
	cs CredentialStore
}

// NewCredentials creates the provider-credential use cases.
func NewCredentials(cs CredentialStore) *Credentials {
	return &Credentials{cs: cs}
}

// List returns every provider record with its secret withheld.
func (c *Credentials) List(ctx context.Context) ([]*ProviderCredential, error) {
	providers, err := c.cs.ListProviders(ctx)
	if err != nil {
		return nil, err
	}
	credentials := make([]*ProviderCredential, 0, len(providers))
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		credentials = append(credentials, provider.ToCredential())
	}
	return credentials, nil
}

// Update sets a provider's API key and/or base URI, returning the secret-free view.
// A nil field is left unchanged, so a base-URI edit cannot silently clear the key.
func (c *Credentials) Update(ctx context.Context, id uuid.UUID, apiKey, baseURI *string) (*ProviderCredential, error) {
	if apiKey == nil && baseURI == nil {
		return nil, ErrNoCredentialFieldsToUpdate
	}
	if apiKey != nil && strings.TrimSpace(*apiKey) == "" {
		return nil, ErrProviderAPIKeyEmpty
	}

	provider, err := c.cs.GetProviderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, ErrProviderNotFound
	}
	if provider.IsManualIngestion() {
		return nil, ErrCredentialsNotConfigurable
	}

	if err := c.cs.UpdateProviderCredentials(ctx, id, apiKey, baseURI); err != nil {
		return nil, err
	}

	updated, err := c.cs.GetProviderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrProviderNotFound
	}
	return updated.ToCredential(), nil
}

// Reveal returns a provider's stored API key in full.
//
// This is the one path that discloses a secret, so callers are expected to gate it
// and to record that it happened.
func (c *Credentials) Reveal(ctx context.Context, id uuid.UUID) (string, error) {
	provider, err := c.cs.GetProviderByID(ctx, id)
	if err != nil {
		return "", err
	}
	if provider == nil {
		return "", ErrProviderNotFound
	}
	if provider.ApiKey == nil || strings.TrimSpace(*provider.ApiKey) == "" {
		return "", ErrProviderAPIKeyEmpty
	}
	return *provider.ApiKey, nil
}
