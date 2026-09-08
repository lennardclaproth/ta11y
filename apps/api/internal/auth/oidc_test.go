package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/lennardclaproth/ta11y/internal/auth"
)

func TestNewFlowGeneratesDistinctSecrets(t *testing.T) {
	flow, err := auth.NewFlow("google", "http://localhost:5199")
	if err != nil {
		t.Fatalf("new flow: %v", err)
	}
	if flow.State == "" || flow.Nonce == "" || flow.CodeVerifier == "" {
		t.Fatal("state, nonce and verifier must all be populated")
	}
	if flow.State == flow.Nonce || flow.State == flow.CodeVerifier || flow.Nonce == flow.CodeVerifier {
		t.Fatal("state, nonce and verifier must be independent secrets")
	}
}

func TestFlowRoundTripsThroughCookieEncoding(t *testing.T) {
	flow, err := auth.NewFlow("google", "http://localhost:5199")
	if err != nil {
		t.Fatalf("new flow: %v", err)
	}

	encoded, err := flow.Encode()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := auth.DecodeFlow(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.State != flow.State || decoded.Nonce != flow.Nonce || decoded.CodeVerifier != flow.CodeVerifier {
		t.Fatal("decoded flow lost a secret")
	}
	if decoded.Provider != "google" || decoded.Redirect != "http://localhost:5199" {
		t.Fatalf("decoded flow = %+v, want provider google and the login redirect", decoded)
	}
}

func TestDecodeFlowRejectsGarbage(t *testing.T) {
	for _, input := range []string{"not base64!", "", "YWJj"} {
		if _, err := auth.DecodeFlow(input); !errors.Is(err, auth.ErrFlowStateMismatch) {
			t.Fatalf("DecodeFlow(%q) error = %v, want ErrFlowStateMismatch", input, err)
		}
	}
}

func TestFlowValidate(t *testing.T) {
	flow, err := auth.NewFlow("google", "")
	if err != nil {
		t.Fatalf("new flow: %v", err)
	}

	t.Run("accepts its own provider and state", func(t *testing.T) {
		if err := flow.Validate("google", flow.State); err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
	})

	t.Run("rejects a mismatched state", func(t *testing.T) {
		if err := flow.Validate("google", "somebody-elses-state"); !errors.Is(err, auth.ErrFlowStateMismatch) {
			t.Fatalf("Validate() = %v, want ErrFlowStateMismatch", err)
		}
	})

	t.Run("rejects an empty state", func(t *testing.T) {
		if err := flow.Validate("google", ""); !errors.Is(err, auth.ErrFlowStateMismatch) {
			t.Fatalf("Validate() = %v, want ErrFlowStateMismatch", err)
		}
	})

	// A flow completed at a different provider's callback would let a login started
	// against one issuer be finished against another.
	t.Run("rejects a different provider", func(t *testing.T) {
		if err := flow.Validate("entra", flow.State); !errors.Is(err, auth.ErrFlowStateMismatch) {
			t.Fatalf("Validate() = %v, want ErrFlowStateMismatch", err)
		}
	})

	t.Run("rejects a lapsed flow", func(t *testing.T) {
		stale := *flow
		stale.IssuedAt = time.Now().UTC().Add(-auth.FlowTTL - time.Minute)
		if err := stale.Validate("google", stale.State); !errors.Is(err, auth.ErrFlowExpired) {
			t.Fatalf("Validate() = %v, want ErrFlowExpired", err)
		}
	})
}

func TestEmptyRegistryRejectsEveryProvider(t *testing.T) {
	registry := auth.NewEmptyRegistry()
	if _, err := registry.Get("google"); !errors.Is(err, auth.ErrProviderUnknown) {
		t.Fatalf("Get() = %v, want ErrProviderUnknown", err)
	}
	if got := registry.Slugs(); len(got) != 0 {
		t.Fatalf("Slugs() = %v, want empty", got)
	}
}
