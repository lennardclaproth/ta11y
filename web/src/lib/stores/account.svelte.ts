import { DEMO_ACCOUNT_ID } from '$lib/api/config';
import { listAccounts } from '$lib/services/account';
import type { Account } from '$lib/api/types';

/** How the account list resolved, so callers can tell "none yet" from "request failed". */
export type AccountStatus = 'idle' | 'ready' | 'empty' | 'failed';

/**
 * Active-account context. The portal is single-account today: on first use it fetches `GET /accounts`
 * and adopts the first account as active, so account-scoped screens (portfolio, assets, realtime)
 * send a real backend id instead of a hard-coded one. In mock mode the demo account is returned.
 *
 * `status` matters: an account list that comes back empty is a legitimate empty state, not a
 * failure. Callers must check `hasAccount` before making account-scoped requests — firing them
 * with the placeholder id would 4xx and surface as a data-load error for what is really just an
 * account that has not been created yet.
 */
let accounts = $state<Account[]>([]);
let activeId = $state<string>(DEMO_ACCOUNT_ID);
let status = $state<AccountStatus>('idle');
let loadPromise: Promise<void> | null = null;

export const accountStore = {
	get accounts(): Account[] {
		return accounts;
	},
	/** The id every account-scoped request should use. Only meaningful once `hasAccount`. */
	get activeId(): string {
		return activeId;
	},
	get status(): AccountStatus {
		return status;
	},
	get loaded(): boolean {
		return status !== 'idle';
	},
	/** True once a real account has been resolved and account-scoped requests are safe. */
	get hasAccount(): boolean {
		return status === 'ready';
	},
	/** True when the account list came back but held nothing — an empty state, not an error. */
	get isEmpty(): boolean {
		return status === 'empty';
	},
	/** True when the account list could not be fetched at all. */
	get failed(): boolean {
		return status === 'failed';
	},
	/**
	 * Whether the active account may reach admin-only screens. Mirrors `accounts.admin` from
	 * the API, which records intent only — the API does not enforce it, so this hides screens
	 * rather than protecting them.
	 */
	get isAdmin(): boolean {
		return accounts.some((account) => account.id === activeId && account.admin);
	},
	/** Fetch the account list once and adopt the first account as active. Idempotent. */
	ensureLoaded(): Promise<void> {
		if (!loadPromise) {
			loadPromise = listAccounts()
				.then((list) => {
					const resolved = list ?? [];
					if (resolved.length === 0) {
						status = 'empty';
						return;
					}
					accounts = resolved;
					activeId = resolved[0].id;
					status = 'ready';
				})
				.catch(() => {
					status = 'failed';
				});
		}
		return loadPromise;
	}
};
