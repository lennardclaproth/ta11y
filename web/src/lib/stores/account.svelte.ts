import { setUnauthorizedHandler } from '$lib/api/client';
import { DEMO_ACCOUNT_ID } from '$lib/api/config';
import { getSession, logout, type Session } from '$lib/services/auth';

/** How the session resolved, so callers can tell "signed out" from "request failed". */
export type AccountStatus = 'idle' | 'ready' | 'empty' | 'failed';

/**
 * Active-account context, resolved from the signed-in session.
 *
 * The account id now comes from `GET /auth/me` rather than `GET /accounts`: the API
 * derives every account-scoped request from the session cookie, so the client's job is
 * to know *who* is signed in, not to nominate an account. `GET /accounts` is
 * administrator-only and no longer part of this path.
 *
 * `status` still matters: `empty` means nobody is signed in, which is a redirect to the
 * sign-in page rather than an error, while `failed` means the API could not be reached
 * and should surface as one.
 */
let session = $state<Session | null>(null);
let status = $state<AccountStatus>('idle');
let loadPromise: Promise<void> | null = null;

/** Where to send the browser when a session is missing or has lapsed. */
const signInPath = '/login';

function redirectToSignIn(): void {
	if (typeof window === 'undefined') return;
	if (window.location.pathname === signInPath) return;
	window.location.assign(signInPath);
}

export const accountStore = {
	/** The signed-in account, or null when signed out. */
	get session(): Session | null {
		return session;
	},
	/** The id every account-scoped request is implicitly scoped to. */
	get activeId(): string {
		return session?.id ?? DEMO_ACCOUNT_ID;
	},
	get email(): string {
		return session?.email ?? '';
	},
	get status(): AccountStatus {
		return status;
	},
	get loaded(): boolean {
		return status !== 'idle';
	},
	/** True once a session has been resolved and account-scoped requests are safe. */
	get hasAccount(): boolean {
		return status === 'ready';
	},
	/** True when nobody is signed in — a redirect to sign-in, not an error. */
	get isEmpty(): boolean {
		return status === 'empty';
	},
	/** True when the session could not be fetched at all. */
	get failed(): boolean {
		return status === 'failed';
	},
	/**
	 * Whether the signed-in account may reach admin-only screens. The API enforces this
	 * independently; hiding the navigation is a convenience, not the control.
	 */
	get isAdmin(): boolean {
		return session?.admin === true;
	},
	/** Fetch the session once. Idempotent. */
	ensureLoaded(): Promise<void> {
		if (!loadPromise) {
			loadPromise = getSession()
				.then((resolved) => {
					if (!resolved) {
						session = null;
						status = 'empty';
						return;
					}
					session = resolved;
					status = 'ready';
				})
				.catch(() => {
					status = 'failed';
				});
		}
		return loadPromise;
	},
	/** Drop the cached session so the next `ensureLoaded` refetches it. */
	reset(): void {
		session = null;
		status = 'idle';
		loadPromise = null;
	},
	/** Revoke the session and return to the sign-in page. */
	async signOut(): Promise<void> {
		try {
			await logout();
		} finally {
			accountStore.reset();
			redirectToSignIn();
		}
	}
};

// A 401 from any request means the session lapsed mid-visit; send the browser to sign in
// rather than letting every screen render its own failure.
setUnauthorizedHandler(() => {
	accountStore.reset();
	status = 'empty';
	redirectToSignIn();
});
