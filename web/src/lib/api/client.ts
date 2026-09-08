import { apiBaseUrl } from './config';

/** A query value; arrays are joined with commas (matches the backend's comma-separated params). */
type QueryValue = string | number | boolean | string[] | undefined | null;

/**
 * Called when the API reports the caller is not signed in. The session store installs a
 * handler that sends the browser to the sign-in page; until then a 401 simply surfaces
 * as an ApiError.
 */
let onUnauthorized: (() => void) | null = null;

/** Register the callback invoked on the first 401 from any request. */
export function setUnauthorizedHandler(handler: (() => void) | null): void {
	onUnauthorized = handler;
}

/** Error thrown when the API responds with a non-2xx status. */
export class ApiError extends Error {
	readonly status: number;
	readonly body: unknown;

	constructor(status: number, message: string, body: unknown) {
		super(message);
		this.name = 'ApiError';
		this.status = status;
		this.body = body;
	}
}

function buildQuery(query?: Record<string, QueryValue>): string {
	if (!query) return '';
	const params = new URLSearchParams();
	for (const [key, value] of Object.entries(query)) {
		if (value === undefined || value === null || value === '') continue;
		params.set(key, Array.isArray(value) ? value.join(',') : String(value));
	}
	const qs = params.toString();
	return qs ? `?${qs}` : '';
}

async function parse<T>(res: Response): Promise<T> {
	const text = await res.text();
	const body = text ? (JSON.parse(text) as unknown) : null;
	if (!res.ok) {
		// A 401 means the session is gone, not that this particular request was bad, so
		// it is handled once centrally rather than by every caller.
		if (res.status === 401) {
			onUnauthorized?.();
		}
		throw new ApiError(res.status, `Request failed with status ${res.status}`, body);
	}
	return body as T;
}

/** GET `path` with optional query params, returning the decoded JSON body. */
export async function apiGet<T>(path: string, query?: Record<string, QueryValue>): Promise<T> {
	const res = await fetch(`${apiBaseUrl}${path}${buildQuery(query)}`, {
		credentials: 'include',
		headers: { Accept: 'application/json' }
	});
	return parse<T>(res);
}

/** Send a JSON body with the given method (POST/PUT/PATCH/DELETE), returning the decoded JSON body. */
export async function apiSend<T>(method: string, path: string, body?: unknown): Promise<T> {
	const res = await fetch(`${apiBaseUrl}${path}`, {
		method,
		credentials: 'include',
		headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
		body: body === undefined ? undefined : JSON.stringify(body)
	});
	return parse<T>(res);
}

/** `POST` with no body, returning nothing. Used by sign-out. */
export async function apiPostEmpty(path: string): Promise<void> {
	const res = await fetch(`${apiBaseUrl}${path}`, {
		method: 'POST',
		credentials: 'include',
		headers: { Accept: 'application/json' }
	});
	if (!res.ok && res.status !== 401) {
		throw new ApiError(res.status, `Request failed with status ${res.status}`, null);
	}
}

/**
 * POST a `multipart/form-data` body (file uploads), returning the decoded JSON body. The
 * `Content-Type` is intentionally left unset so the browser adds the multipart boundary.
 */
export async function apiUpload<T>(path: string, form: FormData): Promise<T> {
	const res = await fetch(`${apiBaseUrl}${path}`, {
		method: 'POST',
		credentials: 'include',
		headers: { Accept: 'application/json' },
		body: form
	});
	return parse<T>(res);
}
