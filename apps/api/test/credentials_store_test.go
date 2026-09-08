//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/lennardclaproth/ta11y/internal/account"
	"github.com/lennardclaproth/ta11y/internal/marketdata"
	"github.com/lennardclaproth/ta11y/internal/storage"
)

// TestMaskAPIKeyWithholdsTheSecret pins the masking rule: a hint must identify a key
// without narrowing it, so short keys are fully masked rather than mostly shown.
func TestMaskAPIKeyWithholdsTheSecret(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"", ""},
		{"abcd", "****"},
		{"abcdefgh", "********"},
		{"ms_live_abcdef1234567890XYZ", "****0XYZ"},
	}
	for _, tc := range cases {
		got := marketdata.MaskAPIKey(tc.key)
		if got != tc.want {
			t.Fatalf("MaskAPIKey(%q) = %q, want %q", tc.key, got, tc.want)
		}
		// Whatever the rule, a hint must never be the key itself.
		if tc.key != "" && got == tc.key {
			t.Fatalf("MaskAPIKey(%q) returned the key unmasked", tc.key)
		}
	}
}

// TestProviderCredentialsRoundTrip covers the credential lifecycle against the real
// migrated schema: a stored key is listed only as a hint, can be replaced, and a
// base-URI edit must not disturb the key.
func TestProviderCredentialsRoundTrip(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXMarketDataStore(db)
		credentials := marketdata.NewCredentials(store)

		provider, err := marketdata.NewAPIProviderWithAPIKey(
			marketdata.ProviderMarketStack, "https://api.marketstack.com/v2", "ms_original_key_1234")
		if err != nil {
			t.Fatalf("new provider: %v", err)
		}
		if err := store.CreateProvider(ctx, provider); err != nil {
			t.Fatalf("create provider: %v", err)
		}

		listed, err := credentials.List(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(listed) != 1 {
			t.Fatalf("got %d credentials, want 1", len(listed))
		}
		if !listed[0].HasAPIKey || listed[0].APIKeyHint != "****1234" {
			t.Fatalf("unexpected credential view: %+v", listed[0])
		}

		// Replacing the key must change the hint and the revealed value together.
		newKey := "ms_rotated_key_ABCD"
		updated, err := credentials.Update(ctx, provider.ID, &newKey, nil)
		if err != nil {
			t.Fatalf("update key: %v", err)
		}
		if updated.APIKeyHint != "****ABCD" {
			t.Fatalf("hint after rotation = %q, want ****ABCD", updated.APIKeyHint)
		}
		revealed, err := credentials.Reveal(ctx, provider.ID)
		if err != nil {
			t.Fatalf("reveal: %v", err)
		}
		if revealed != newKey {
			t.Fatalf("revealed %q, want %q", revealed, newKey)
		}

		// Editing only the base URI must leave the key untouched.
		newBase := "https://api.marketstack.com/v1"
		if _, err := credentials.Update(ctx, provider.ID, nil, &newBase); err != nil {
			t.Fatalf("update base uri: %v", err)
		}
		afterBase, err := credentials.Reveal(ctx, provider.ID)
		if err != nil {
			t.Fatalf("reveal after base uri change: %v", err)
		}
		if afterBase != newKey {
			t.Fatalf("base-uri edit changed the key: got %q, want %q", afterBase, newKey)
		}
	})
}

// TestProviderCredentialsRejectManualProviders checks that a provider ingesting
// uploaded files cannot be given an endpoint or key it would never use.
func TestProviderCredentialsRejectManualProviders(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXMarketDataStore(db)
		credentials := marketdata.NewCredentials(store)

		provider, err := marketdata.NewManualProvider(marketdata.ProviderBrandNewDay)
		if err != nil {
			t.Fatalf("new manual provider: %v", err)
		}
		if err := store.CreateProvider(ctx, provider); err != nil {
			t.Fatalf("create manual provider: %v", err)
		}

		key := "should-not-apply"
		if _, err := credentials.Update(ctx, provider.ID, &key, nil); err == nil {
			t.Fatal("expected a manual provider to reject credentials")
		} else if !strings.Contains(err.Error(), "manual") {
			t.Fatalf("unexpected error for manual provider: %v", err)
		}
	})
}

// TestAccountAdminFlagPersists covers the new column: the admin flag must survive a
// round trip, and ordinary accounts must default to non-admin.
func TestAccountAdminFlagPersists(t *testing.T) {
	eachDialect(t, func(t *testing.T, db *storage.DB) {
		ctx := context.Background()
		store := storage.NewSQLXAccountStore(db)

		admin, err := account.NewAccount("Owner", nil, nil, account.AsAdmin())
		if err != nil {
			t.Fatalf("new admin account: %v", err)
		}
		plain, err := account.NewAccount("Guest", nil, nil)
		if err != nil {
			t.Fatalf("new plain account: %v", err)
		}
		for _, acc := range []*account.Account{admin, plain} {
			if err := store.Create(ctx, acc); err != nil {
				t.Fatalf("create %s: %v", acc.Name, err)
			}
		}

		gotAdmin, err := store.GetByID(ctx, admin.ID)
		if err != nil {
			t.Fatalf("get admin: %v", err)
		}
		if !gotAdmin.Admin {
			t.Fatal("admin flag did not persist")
		}

		gotPlain, err := store.GetByID(ctx, plain.ID)
		if err != nil {
			t.Fatalf("get plain: %v", err)
		}
		if gotPlain.Admin {
			t.Fatal("an account created without AsAdmin should not be an admin")
		}

		// The list path feeds the web app's admin gating, so it must carry the flag too.
		listed, err := store.List(ctx)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		flags := map[string]bool{}
		for _, acc := range listed {
			flags[acc.Name] = acc.Admin
		}
		if !flags["Owner"] || flags["Guest"] {
			t.Fatalf("list did not carry admin flags: %+v", flags)
		}
	})
}
