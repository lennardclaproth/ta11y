/** Minimal account identity. Mirrors `account.AccountResponse`. */
export interface Account {
	id: string;
	name: string;
	/**
	 * Whether this account may reach admin-only screens. The API records this but does
	 * not enforce it, so clients use it to hide screens, not to protect them.
	 */
	admin: boolean;
	external_id?: string | null;
}

/** `GET /accounts` returns a bare array of accounts. */
export type AccountsResponse = Account[];

/** `POST /accounts` request — mirrors `account.CreateAccountRequest`. */
export interface CreateAccountRequest {
	name: string;
	external_id?: string;
}

/** `POST /accounts` — mirrors `account.CreateAccountResponse`. */
export interface CreateAccountResponse {
	id: string;
}
