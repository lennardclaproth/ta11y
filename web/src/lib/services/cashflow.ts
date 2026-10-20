import { apiGet, apiSend } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	CashflowBulkMutationResponse,
	CashflowDirection,
	ChangeCashflowTransactionDateRequest,
	ChangeCashflowTransactionDateResponse,
	CashflowMonthlyAnalyticsResponse,
	CashflowTransaction,
	CashflowTransactionsQuery,
	CashflowTransactionsResponse,
	CreateCashflowTransactionsRequest,
	CreateCashflowTransactionsResponse,
	IgnoreTransactionsByFilterRequest,
	IgnoreTransactionsBySelectionRequest,
	TagDistributionEntry,
	TagDistributionResponse,
	TagTransactionRequest,
	TagTransactionsByFilterRequest,
	TagTransactionsBySelectionRequest
} from '$lib/api/types';
import { cashflowMonthly, cashflowTransactions } from '$lib/data/fixtures/cashflow';
import { clone, contains, delay } from './_mock';

function compare(a: CashflowTransaction, b: CashflowTransaction, sortBy: string): number {
	switch (sortBy) {
		case 'amount':
			return a.amountCents - b.amountCents;
		case 'description':
			return a.description.localeCompare(b.description);
		case 'note':
			return a.note.localeCompare(b.note);
		case 'tag':
			return a.tag.localeCompare(b.tag);
		case 'source':
			return a.source.localeCompare(b.source);
		case 'date':
		default:
			return a.date.localeCompare(b.date);
	}
}

/**
 * The ledger fixture mode works against. It is a copy rather than the fixture itself
 * because ignoring and restoring have to stick for the session: a rule applied on the
 * rules page has to be visible on Cashflow, and a row put back has to stay back — exactly
 * what the feature is about. Nothing is persisted beyond a reload.
 */
let mockRows: CashflowTransaction[] = clone(cashflowTransactions);

/** The session's ledger, for the fixture branches of the ignore-rule service. */
export function mockCashflowRows(): CashflowTransaction[] {
	return mockRows;
}

/**
 * Attributes every row an ignore rule claimed to one import. Fixture rows carry no import
 * of their own, so the review the person opened stands in for it — without this the group
 * it shows could not be filtered back to, and "show the rest" would return the ledger.
 */
export function mutateMockImportId(importId: string): void {
	mockRows = mockRows.map((tx) =>
		tx.ignored_by_rule_id && !tx.import_id ? { ...tx, import_id: importId } : tx
	);
}

/** Sets the ignored state of every session row the predicate picks, and counts them. */
export function mutateMockIgnored(
	predicate: (tx: CashflowTransaction) => boolean,
	ignored: boolean,
	ruleId: string | null
): number {
	let updated = 0;
	mockRows = mockRows.map((tx) => {
		if (!predicate(tx)) return tx;
		updated += 1;
		return {
			...tx,
			ignored,
			// A rule's own apply keeps its attribution and leaves the row open to rules.
			// A by-hand change marks the row, and only an ignore clears the attribution:
			// restoring keeps it so the import review can still name the rule.
			ignored_by_rule_id: ruleId ?? (ignored ? null : (tx.ignored_by_rule_id ?? null)),
			ignore_overridden: ruleId === null ? true : (tx.ignore_overridden ?? false)
		};
	});
	return updated;
}

function mockTransactions(query: CashflowTransactionsQuery): CashflowTransactionsResponse {
	const tags = (query.tags ?? '')
		.split(',')
		.map((t) => t.trim())
		.filter(Boolean);

	let rows = mockRows.filter((tx) => {
		if (query.direction && tx.direction !== query.direction) return false;
		if (query.hide_ignored && tx.ignored) return false;
		if (query.untagged && tx.tag !== '') return false;
		if (tags.length > 0 && !tags.includes(tx.tag)) return false;
		if (!contains(tx.description, query.description)) return false;
		if (!contains(tx.note, query.note)) return false;
		if (!contains(tx.source, query.source)) return false;
		if (query.import_id && tx.import_id !== query.import_id) return false;
		if (query.ignored_by_rule && tx.ignored_by_rule_id !== query.ignored_by_rule) return false;
		if (query.from && tx.date.slice(0, 10) < query.from) return false;
		if (query.to && tx.date.slice(0, 10) > query.to) return false;
		if (query.q) {
			const q = query.q.toLowerCase();
			const hit =
				tx.description.toLowerCase().includes(q) ||
				tx.note.toLowerCase().includes(q) ||
				tx.tag.toLowerCase().includes(q);
			if (!hit) return false;
		}
		return true;
	});

	const sortBy = query.sort_by ?? 'date';
	const dir = query.sort_order === 'asc' ? 1 : -1;
	rows = rows.slice().sort((a, b) => compare(a, b, sortBy) * dir);

	const total = rows.length;
	const limit = query.limit ?? 100;
	const offset = query.offset ?? 0;
	const data = clone(rows.slice(offset, offset + limit));

	return { pagination: { limit, offset, count: data.length, total }, data };
}

/** `GET /cashflow/transactions` */
export async function listCashflowTransactions(
	query: CashflowTransactionsQuery = {}
): Promise<CashflowTransactionsResponse> {
	if (useMocks) {
		await delay();
		return mockTransactions(query);
	}
	return apiGet<CashflowTransactionsResponse>('/cashflow/transactions', { ...query });
}

/** `GET /cashflow/analytics/monthly` */
export async function getCashflowMonthly(
	query: { from?: string; to?: string; include_ignored?: boolean } = {}
): Promise<CashflowMonthlyAnalyticsResponse> {
	if (useMocks) {
		await delay();
		// Month keys are first-of-month, so compare on the YYYY-MM prefix to keep months that
		// overlap an arbitrary range bound (e.g. from="2026-06-08" must still include June).
		const fromMonth = query.from?.slice(0, 7);
		const toMonth = query.to?.slice(0, 7);
		const data = clone(
			cashflowMonthly.filter((p) => {
				const m = p.month.slice(0, 7);
				return (!fromMonth || m >= fromMonth) && (!toMonth || m <= toMonth);
			})
		);
		return { data };
	}
	return apiGet<CashflowMonthlyAnalyticsResponse>('/cashflow/analytics/monthly', { ...query });
}

/** Sum the (1e6-scaled) magnitudes per tag for one direction, largest first. */
function tagTotals(rows: CashflowTransaction[], direction: CashflowDirection): TagDistributionEntry[] {
	const totals = new Map<string, number>();
	for (const tx of rows) {
		if (tx.direction !== direction) continue;
		totals.set(tx.tag, (totals.get(tx.tag) ?? 0) + tx.amountCents);
	}
	return [...totals.entries()]
		.map(([tag, totalCents]) => ({ tag, totalCents }))
		.sort((a, b) => b.totalCents - a.totalCents);
}

/** `GET /cashflow/analytics/tags` */
export async function getCashflowTagDistribution(
	query: { from?: string; to?: string; include_ignored?: boolean } = {}
): Promise<TagDistributionResponse> {
	if (useMocks) {
		await delay();
		// Derive the distribution from the dated transactions so the donuts react to the range.
		const rows = cashflowTransactions.filter((tx) => {
			if (!query.include_ignored && tx.ignored) return false;
			if (query.from && tx.date.slice(0, 10) < query.from) return false;
			if (query.to && tx.date.slice(0, 10) > query.to) return false;
			return true;
		});
		const incoming = tagTotals(rows, 'in');
		const outgoing = tagTotals(rows, 'out');
		const combined = [...incoming, ...outgoing].sort((a, b) => b.totalCents - a.totalCents);
		return clone({ incoming, outgoing, combined });
	}
	return apiGet<TagDistributionResponse>('/cashflow/analytics/tags', { ...query });
}

/** `POST /cashflow/transactions/manual` */
export async function createCashflowTransactions(
	body: CreateCashflowTransactionsRequest
): Promise<CreateCashflowTransactionsResponse> {
	if (useMocks) {
		await delay();
		return { created_count: body.transactions.length, data: [] };
	}
	return apiSend<CreateCashflowTransactionsResponse>(
		'POST',
		'/cashflow/transactions/manual',
		body
	);
}

/** `POST /cashflow/transactions/date` */
export async function changeCashflowTransactionDate(
	body: ChangeCashflowTransactionDateRequest
): Promise<ChangeCashflowTransactionDateResponse> {
	if (useMocks) {
		await delay();
		return { id: body.id, date: `${body.date}T00:00:00Z` };
	}
	return apiSend<ChangeCashflowTransactionDateResponse>('POST', '/cashflow/transactions/date', body);
}

/** `POST /cashflow/transactions/tag` (single transaction) */
export async function tagCashflowTransaction(body: TagTransactionRequest): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiSend<unknown>('POST', '/cashflow/transactions/tag', body);
}

/** `POST /cashflow/transactions/tag/selection` */
export async function tagCashflowTransactionsBySelection(
	body: TagTransactionsBySelectionRequest
): Promise<CashflowBulkMutationResponse> {
	if (useMocks) {
		await delay();
		return { updated_count: body.ids.length, status: 'ok' };
	}
	return apiSend<CashflowBulkMutationResponse>(
		'POST',
		'/cashflow/transactions/tag/selection',
		body
	);
}

/** `POST /cashflow/transactions/tag/filter` */
export async function tagCashflowTransactionsByFilter(
	body: TagTransactionsByFilterRequest
): Promise<CashflowBulkMutationResponse> {
	if (useMocks) {
		await delay();
		return { updated_count: 0, status: 'ok' };
	}
	return apiSend<CashflowBulkMutationResponse>('POST', '/cashflow/transactions/tag/filter', body);
}

/** `POST /cashflow/transactions/ignore/selection` */
export async function ignoreCashflowTransactionsBySelection(
	body: IgnoreTransactionsBySelectionRequest
): Promise<CashflowBulkMutationResponse> {
	if (useMocks) {
		await delay();
		const updated = mutateMockIgnored(
			(tx) => body.ids.includes(tx.id),
			body.ignored ?? true,
			null
		);
		return { updated_count: updated, status: 'ok' };
	}
	return apiSend<CashflowBulkMutationResponse>(
		'POST',
		'/cashflow/transactions/ignore/selection',
		body
	);
}

/** `POST /cashflow/transactions/ignore/filter` */
export async function ignoreCashflowTransactionsByFilter(
	body: IgnoreTransactionsByFilterRequest
): Promise<CashflowBulkMutationResponse> {
	if (useMocks) {
		await delay();
		const f = body.filters;
		const updated = mutateMockIgnored(
			(tx) =>
				(!f.import_id || tx.import_id === f.import_id) &&
				(!f.ignored_by_rule || tx.ignored_by_rule_id === f.ignored_by_rule) &&
				(!f.direction || tx.direction === f.direction) &&
				contains(tx.description, f.description) &&
				contains(tx.note, f.note) &&
				contains(tx.source, f.source),
			body.ignored ?? true,
			null
		);
		return { updated_count: updated, status: 'ok' };
	}
	return apiSend<CashflowBulkMutationResponse>(
		'POST',
		'/cashflow/transactions/ignore/filter',
		body
	);
}
