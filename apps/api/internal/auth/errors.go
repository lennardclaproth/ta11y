package auth

import "errors"

var (
	// ErrProviderUnknown indicates that no identity provider is configured under the requested slug.
	ErrProviderUnknown = errors.New("identity provider is not configured")
	// ErrIdentityNotFound indicates that no local identity matches the provider's
	// (issuer, subject) pair, meaning this is a first login through that provider.
	ErrIdentityNotFound = errors.New("identity not found")
	// ErrIdentityExists indicates that the provider's (issuer, subject) pair is already
	// linked, which in practice means two first logins raced.
	ErrIdentityExists = errors.New("identity already exists")
	// ErrSessionNotFound indicates that no session matches the presented token.
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionExpired indicates that the presented session has lapsed or been revoked.
	ErrSessionExpired = errors.New("session expired")
	// ErrEmailMissing indicates that the identity provider returned no email claim, leaving
	// nothing to identify the account by.
	ErrEmailMissing = errors.New("identity provider returned no email")
	// ErrEmailUnverified indicates that the provider did not assert the email as verified.
	// Linking on an unverified email would let a provider that permits arbitrary addresses
	// take over an existing account, so provisioning refuses it.
	ErrEmailUnverified = errors.New("identity provider did not verify the email")
	// ErrEmailNotAllowed indicates that the address authenticated correctly but is not
	// on the sign-in allowlist.
	ErrEmailNotAllowed = errors.New("email is not permitted to sign in")
	// ErrFlowStateMismatch indicates that the callback's state did not match the one issued
	// at login, which means the response cannot be tied to a login this API started.
	ErrFlowStateMismatch = errors.New("authentication state mismatch")
	// ErrFlowExpired indicates that the login flow took longer than the flow cookie's lifetime.
	ErrFlowExpired = errors.New("authentication flow expired")
)
