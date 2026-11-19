import { apiGet, apiPostEmpty, ApiError } from '$lib/api/client';
import { apiBaseUrl, DEMO_ACCOUNT_ID, useMocks } from '$lib/api/config';
import { delay } from './_mock';

/** The signed-in account, as `GET /auth/me` returns it. */
export interface Session {
	id: string;
	email: string;
	/** Gates the admin-only screens. The API enforces it too. */
	admin: boolean;
}

/** The identity providers the API is configured with, from `GET /auth/providers`. */
export interface ProvidersResponse {
	providers: string[];
}

const mockSession: Session = {
	id: DEMO_ACCOUNT_ID,
	email: 'demo@example.com',
	admin: true
};

/**
 * `GET /auth/me` — the signed-in account, or null when there is no live session.
 *
 * A 401 is the ordinary "not signed in" answer rather than a failure, so it resolves to
 * null; anything else propagates, because a broken API should not look like a signed-out
 * user.
 */
export async function getSession(): Promise<Session | null> {
	if (useMocks) {
		await delay();
		return mockSession;
	}
	try {
		return await apiGet<Session>('/auth/me');
	} catch (error) {
		if (error instanceof ApiError && error.status === 401) return null;
		throw error;
	}
}

/** `GET /auth/providers` — the provider slugs the sign-in page renders a button for. */
export async function listProviders(): Promise<string[]> {
	if (useMocks) {
		await delay();
		return ['google'];
	}
	const res = await apiGet<ProvidersResponse>('/auth/providers');
	return res.providers ?? [];
}

/** `POST /auth/logout` — revokes the session and clears the cookie. */
export async function logout(): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiPostEmpty('/auth/logout');
}

/**
 * The URL that starts a sign-in. It is a full page navigation rather than a fetch: the
 * flow is a redirect to the identity provider and back, which XHR cannot follow.
 */
export function loginUrl(provider: string): string {
	return `${apiBaseUrl}/auth/${encodeURIComponent(provider)}/login`;
}

/** Human-readable text for the `error` code the callback redirects back with. */
export function signInErrorMessage(code: string | null): string | null {
	if (!code) return null;
	switch (code) {
		case 'not_allowed':
			return 'That account is not permitted to sign in to this application.';
		case 'email_unverified':
			return 'Your identity provider has not verified that email address.';
		case 'email_missing':
			return 'Your identity provider did not share an email address.';
		case 'flow_expired':
			return 'That sign-in attempt took too long. Please try again.';
		case 'flow_missing':
		case 'flow_invalid':
			return 'That sign-in attempt could not be verified. Please try again.';
		case 'unknown_provider':
			return 'That sign-in method is not available.';
		case 'provider_error':
			return 'Your identity provider declined the sign-in.';
		default:
			return 'Sign-in failed. Please try again.';
	}
}

/** Display name for a provider slug. */
export function providerLabel(provider: string): string {
	switch (provider) {
		case 'google':
			return 'Google';
		case 'entra':
		case 'entra-id':
			return 'Microsoft';
		default:
			return provider.charAt(0).toUpperCase() + provider.slice(1);
	}
}
