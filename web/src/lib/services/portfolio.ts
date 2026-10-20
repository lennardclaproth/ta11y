import { ApiError, apiGet, apiSend } from '$lib/api/client';
import { DEMO_ACCOUNT_ID, useMocks } from '$lib/api/config';
import type {
	CreateManualPortfolioTransactionRequest,
	DeletePortfolioTransactionResponse,
	ManualPortfolioTransactionResponse,
	PortfolioPositionsQuery,
	PortfolioPositionsResponse,
	PortfolioSnapshotsQuery,
	PortfolioSnapshotsResponse,
	PortfolioTransactionsQuery,
	PortfolioTransactionsResponse,
	UpdatePortfolioTransactionRequest
} from '$lib/api/types';
import {
	portfolioPositions,
	portfolioSnapshots,
	portfolioTransactions
} from '$lib/data/fixtures/portfolio';
import { clone, contains, delay } from './_mock';

/** `GET /portfolio/positions` */
export async function listPortfolioPositions(
	query: PortfolioPositionsQuery
): Promise<PortfolioPositionsResponse> {
	if (useMocks) {
		await delay();
		const data = clone(
			portfolioPositions.filter((p) => (query.include_closed ? true : !p.is_closed))
		);
		return { include_closed: Boolean(query.include_closed), data };
	}
	return apiGet<PortfolioPositionsResponse>('/portfolio/positions', { ...query });
}

/** `GET /portfolio/snapshots` */
export async function getPortfolioSnapshots(
	query: PortfolioSnapshotsQuery
): Promise<PortfolioSnapshotsResponse> {
	if (useMocks) {
		await delay();
		return clone(
			portfolioSnapshots.filter(
				(p) =>
					(!query.from || p.occurred_at.slice(0, 10) >= query.from) &&
					(!query.to || p.occurred_at.slice(0, 10) <= query.to)
			)
		);
	}
	return apiGet<PortfolioSnapshotsResponse>('/portfolio/snapshots', { ...query });
}

/** `GET /portfolio/transactions` */
export async function listPortfolioTransactions(
	query: PortfolioTransactionsQuery
): Promise<PortfolioTransactionsResponse> {
	if (useMocks) {
		await delay();
		let rows = portfolioTransactions.filter((tx) => {
			if (query.type && tx.type !== query.type) return false;
			if (query.origin && tx.origin !== query.origin) return false;
			if (!contains(tx.source, query.source)) return false;
			if (query.from && tx.occurred_at.slice(0, 10) < query.from) return false;
			if (query.to && tx.occurred_at.slice(0, 10) > query.to) return false;
			if (query.listing) {
				const l = query.listing.toLowerCase();
				const hit =
					(tx.symbol ?? '').toLowerCase().includes(l) || (tx.isin ?? '').toLowerCase().includes(l);
				if (!hit) return false;
			}
			if (query.q) {
				const q = query.q.toLowerCase();
				const hit =
					tx.description.toLowerCase().includes(q) ||
					tx.source.toLowerCase().includes(q) ||
					(tx.symbol ?? '').toLowerCase().includes(q) ||
					(tx.isin ?? '').toLowerCase().includes(q);
				if (!hit) return false;
			}
			return true;
		});

		const dir = query.sort_order === 'asc' ? 1 : -1;
		rows = rows.slice().sort((a, b) => a.occurred_at.localeCompare(b.occurred_at) * dir);

		const total = rows.length;
		const limit = query.limit ?? 25;
		const offset = query.offset ?? 0;
		const data = clone(rows.slice(offset, offset + limit));
		return { pagination: { limit, offset, count: data.length, total }, data };
	}
	return apiGet<PortfolioTransactionsResponse>('/portfolio/transactions', { ...query });
}

/** `POST /portfolio/transactions/manual` — also rebuilds, reporting whether that ran. */
export async function createManualPortfolioTransaction(
	body: CreateManualPortfolioTransactionRequest
): Promise<ManualPortfolioTransactionResponse> {
	if (useMocks) {
		await delay();
		const now = new Date().toISOString();
		return {
			id: crypto.randomUUID(),
			account_id: DEMO_ACCOUNT_ID,
			origin: 'MANUAL',
			source: 'manual',
			occurred_at: body.occurred_at,
			type: body.type,
			listing_id: body.listing_id ?? null,
			isin: null,
			symbol: null,
			description: body.description ?? '',
			amount: body.amount,
			quantity: body.quantity ?? '0',
			unit_price: '0',
			created_at: now,
			updated_at: now,
			rebuild: 'skipped'
		};
	}
	return apiSend<ManualPortfolioTransactionResponse>(
		'POST',
		'/portfolio/transactions/manual',
		body
	);
}

/** `PUT /portfolio/transactions/{id}` — manual rows only; also rebuilds, reporting whether that ran. */
export async function updatePortfolioTransaction(
	id: string,
	body: UpdatePortfolioTransactionRequest
): Promise<ManualPortfolioTransactionResponse> {
	if (useMocks) {
		await delay();
		const existing = portfolioTransactions.find((tx) => tx.id === id);
		if (!existing) {
			throw new ApiError(404, 'Request failed with status 404', {
				transaction_id: 'no transaction found with the given id'
			});
		}
		if (existing.origin !== 'MANUAL') {
			throw new ApiError(422, 'Request failed with status 422', {
				transaction: 'only manually entered transactions can be changed or deleted'
			});
		}
		const quantity = body.quantity ?? '0';
		const isTrade = body.type === 'BUY' || body.type === 'SELL';
		Object.assign(existing, {
			occurred_at: `${body.occurred_at}T00:00:00Z`,
			type: body.type,
			listing_id: body.listing_id ?? null,
			description: body.description ?? '',
			amount: body.amount,
			quantity,
			unit_price:
				isTrade && Number(quantity) > 0 ? String(Number(body.amount) / Number(quantity)) : '0',
			updated_at: new Date().toISOString()
		});
		return { ...clone(existing), rebuild: 'skipped' };
	}
	return apiSend<ManualPortfolioTransactionResponse>(
		'PUT',
		`/portfolio/transactions/${encodeURIComponent(id)}`,
		body
	);
}

/** `DELETE /portfolio/transactions/{id}` — manual rows only; also rebuilds, reporting whether that ran. */
export async function deletePortfolioTransaction(
	id: string
): Promise<DeletePortfolioTransactionResponse> {
	if (useMocks) {
		await delay();
		const index = portfolioTransactions.findIndex((tx) => tx.id === id);
		if (index === -1) {
			throw new ApiError(404, 'Request failed with status 404', {
				transaction_id: 'no transaction found with the given id'
			});
		}
		if (portfolioTransactions[index].origin !== 'MANUAL') {
			throw new ApiError(422, 'Request failed with status 422', {
				transaction: 'only manually entered transactions can be changed or deleted'
			});
		}
		portfolioTransactions.splice(index, 1);
		return { id, rebuild: 'skipped' };
	}
	return apiSend<DeletePortfolioTransactionResponse>(
		'DELETE',
		`/portfolio/transactions/${encodeURIComponent(id)}`
	);
}

/** `POST /portfolio/rebuild` — the account is taken from the session. */
export async function rebuildPortfolio(): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiSend<unknown>('POST', '/portfolio/rebuild', {});
}
