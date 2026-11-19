package marketdata

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeCredentialStore struct {
	providers []*Provider
}

func (s *fakeCredentialStore) ListProviders(context.Context) ([]*Provider, error) {
	return s.providers, nil
}

func (s *fakeCredentialStore) GetProviderByID(_ context.Context, id uuid.UUID) (*Provider, error) {
	for _, p := range s.providers {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (s *fakeCredentialStore) CreateProvider(_ context.Context, provider *Provider) error {
	s.providers = append(s.providers, provider)
	return nil
}

func (s *fakeCredentialStore) UpdateProviderCredentials(context.Context, uuid.UUID, *string, *string) error {
	return nil
}

func newTestCredentials() (*Credentials, *fakeCredentialStore) {
	store := &fakeCredentialStore{}
	return NewCredentials(store, []APIProviderDefault{
		{Name: ProviderMarketStack, BaseURI: "https://api.marketstack.com/v2"},
		{Name: ProviderAlphaVantage, BaseURI: "https://www.alphavantage.co"},
	}), store
}

func TestCredentialsAddUsesDefaultBaseURI(t *testing.T) {
	credentials, store := newTestCredentials()

	got, err := credentials.Add(context.Background(), ProviderAlphaVantage, "  av_key_abcdef12  ", nil)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if got.BaseURI == nil || *got.BaseURI != "https://www.alphavantage.co" {
		t.Fatalf("base uri = %v, want the default", got.BaseURI)
	}
	if got.APIKeyHint != "****ef12" || !got.HasAPIKey {
		t.Fatalf("unexpected credential view: %+v", got)
	}
	if len(store.providers) != 1 || *store.providers[0].ApiKey != "av_key_abcdef12" {
		t.Fatalf("stored key was not trimmed: %+v", store.providers)
	}
}

func TestCredentialsAddKeepsExistingKeys(t *testing.T) {
	credentials, store := newTestCredentials()
	ctx := context.Background()

	if _, err := credentials.Add(ctx, ProviderMarketStack, "ms_key_one_1111", nil); err != nil {
		t.Fatalf("add first: %v", err)
	}
	custom := "https://api.marketstack.com/v1"
	second, err := credentials.Add(ctx, ProviderMarketStack, "ms_key_two_2222", &custom)
	if err != nil {
		t.Fatalf("add second: %v", err)
	}
	if *second.BaseURI != custom {
		t.Fatalf("base uri = %q, want %q", *second.BaseURI, custom)
	}
	if len(store.providers) != 2 {
		t.Fatalf("got %d providers, want 2", len(store.providers))
	}
}

func TestCredentialsAddRejects(t *testing.T) {
	ctx := context.Background()
	invalid := "http://localhost:8080"

	cases := []struct {
		name    string
		seed    string
		target  ProviderName
		key     string
		baseURI *string
		want    error
	}{
		{name: "duplicate key", seed: "ms_dup_key_9999", target: ProviderMarketStack, key: "ms_dup_key_9999", want: ErrProviderAPIKeyExists},
		{name: "manual provider", target: ProviderBrandNewDay, key: "anything_1234", want: ErrProviderKeyNotSupported},
		{name: "empty key", target: ProviderMarketStack, key: "   ", want: ErrProviderAPIKeyEmpty},
		{name: "invalid base uri", target: ProviderMarketStack, key: "ms_key_5555", baseURI: &invalid, want: ErrProviderBaseURIInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			credentials, store := newTestCredentials()
			if tc.seed != "" {
				if _, err := credentials.Add(ctx, tc.target, tc.seed, nil); err != nil {
					t.Fatalf("seed: %v", err)
				}
			}
			before := len(store.providers)
			if _, err := credentials.Add(ctx, tc.target, tc.key, tc.baseURI); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if len(store.providers) != before {
				t.Fatal("a rejected key was stored")
			}
		})
	}
}

func TestCredentialsAPIProvidersListsAllDefaults(t *testing.T) {
	credentials, _ := newTestCredentials()
	got := credentials.APIProviders()
	if len(got) != 2 || got[0].Name != ProviderMarketStack || got[1].Name != ProviderAlphaVantage {
		t.Fatalf("unexpected providers: %+v", got)
	}
}
