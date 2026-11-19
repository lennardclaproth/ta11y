package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// FlowTTL bounds how long a login may sit half-finished before the flow cookie lapses.
const FlowTTL = 10 * time.Minute

// flowSecretBytes is the entropy behind each of the flow's three secrets.
const flowSecretBytes = 32

// Flow is the short-lived state of one login attempt. It lives in an HttpOnly cookie
// between the redirect to the provider and the callback.
//
// The cookie is not signed. Its three values are only ever compared against what the
// provider echoes back, so tampering with one's own cookie breaks one's own login and
// nothing else; keeping the nonce and verifier out of reach of other origins is what
// matters, and HttpOnly plus SameSite=Lax does that.
type Flow struct {
	// Provider is the slug the login started at, checked against the callback route so
	// a flow cannot be completed at a different provider's endpoint.
	Provider string `json:"p"`
	// State ties the callback to this login attempt (CSRF defence on the callback).
	State string `json:"s"`
	// Nonce ties the returned ID token to this login attempt (replay defence).
	Nonce string `json:"n"`
	// CodeVerifier is the PKCE secret proving the code is redeemed by whoever asked for it.
	CodeVerifier string `json:"v"`
	// Redirect is where to send the browser once the login completes.
	Redirect string `json:"r,omitempty"`
	// IssuedAt bounds the flow's lifetime independently of the cookie's own expiry.
	IssuedAt time.Time `json:"i"`
}

// NewFlow starts a login attempt for the given provider.
func NewFlow(provider, redirect string) (*Flow, error) {
	state, err := randomSecret()
	if err != nil {
		return nil, err
	}
	nonce, err := randomSecret()
	if err != nil {
		return nil, err
	}
	verifier, err := randomSecret()
	if err != nil {
		return nil, err
	}
	return &Flow{
		Provider:     provider,
		State:        state,
		Nonce:        nonce,
		CodeVerifier: verifier,
		Redirect:     redirect,
		IssuedAt:     time.Now().UTC(),
	}, nil
}

// Encode serialises the flow for transport in a cookie.
func (f *Flow) Encode() (string, error) {
	raw, err := json.Marshal(f)
	if err != nil {
		return "", fmt.Errorf("auth: encode flow: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// DecodeFlow parses a flow cookie value.
func DecodeFlow(encoded string) (*Flow, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, ErrFlowStateMismatch
	}
	var flow Flow
	if err := json.Unmarshal(raw, &flow); err != nil {
		return nil, ErrFlowStateMismatch
	}
	return &flow, nil
}

// Validate checks a callback against the flow that started it: the provider must be the
// one the login began at, the state must match, and the flow must not have lapsed.
// State is compared in constant time.
func (f *Flow) Validate(provider, state string) error {
	if time.Since(f.IssuedAt) > FlowTTL {
		return ErrFlowExpired
	}
	if f.Provider != provider {
		return ErrFlowStateMismatch
	}
	if subtle.ConstantTimeCompare([]byte(f.State), []byte(state)) != 1 {
		return ErrFlowStateMismatch
	}
	return nil
}

func randomSecret() (string, error) {
	buf := make([]byte, flowSecretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("auth: generate secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
