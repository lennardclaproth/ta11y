import { apiGet, apiSend } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import { numberToScaled } from '$lib/api/money';
import type {
	ConfirmRecurringSuggestionRequest,
	CreateRecurringItemRequest,
	DismissRecurringSuggestionRequest,
	EndRecurringItemRequest,
	LinkRecurringTransactionsRequest,
	RecurringItem,
	RecurringItemDetail,
	RecurringMonthPoint,
	RecurringMutationResponse,
	RecurringOverviewResponse,
	RecurringSuggestionsResponse,
	UpdateRecurringItemRequest
} from '$lib/api/types';
import {
	recurringEnded,
	recurringExpenses,
	recurringIncome,
	recurringLinkedTransactions,
	recurringSeriesExpenses,
	recurringSeriesIncome,
	recurringSeriesMonths,
	recurringSuggestions
} from '$lib/data/fixtures/recurring';
import { clone, delay, mockId } from './_mock';

function matchesSearch(item: RecurringItem, q: string | undefined): boolean {
	if (!q) return true;
	return item.name.toLowerCase().includes(q.toLowerCase());
}

function mockSeries(): RecurringMonthPoint[] {
	return recurringSeriesMonths.map((month, index) => ({
		month: `${month}T00:00:00Z`,
		expenseCents: numberToScaled(recurringSeriesExpenses[index]),
		incomeCents: numberToScaled(recurringSeriesIncome[index])
	}));
}

/** `GET /cashflow/recurring` */
export async function getRecurringOverview(
	query: { q?: string } = {}
): Promise<RecurringOverviewResponse> {
	if (useMocks) {
		await delay();
		return clone({
			expenses: recurringExpenses.filter((item) => matchesSearch(item, query.q)),
			income: recurringIncome.filter((item) => matchesSearch(item, query.q)),
			ended: recurringEnded.filter((item) => matchesSearch(item, query.q)),
			// The totals describe the account, so a search does not shrink them.
			monthlyExpenseCents: numberToScaled(
				recurringSeriesExpenses[recurringSeriesExpenses.length - 1]
			),
			monthlyIncomeCents: numberToScaled(recurringSeriesIncome[recurringSeriesIncome.length - 1]),
			series: mockSeries()
		});
	}
	return apiGet<RecurringOverviewResponse>('/cashflow/recurring', { ...query });
}

/** `GET /cashflow/recurring/{item_id}` */
export async function getRecurringItem(id: string): Promise<RecurringItemDetail> {
	if (useMocks) {
		await delay();
		const all = [...recurringExpenses, ...recurringIncome, ...recurringEnded];
		const item = all.find((candidate) => candidate.id === id) ?? all[0];
		return clone({ ...item, transactions: recurringLinkedTransactions[item.id] ?? [] });
	}
	return apiGet<RecurringItemDetail>(`/cashflow/recurring/${id}`);
}

/** `GET /cashflow/recurring/suggestions` */
export async function getRecurringSuggestions(): Promise<RecurringSuggestionsResponse> {
	if (useMocks) {
		await delay();
		return clone({ data: recurringSuggestions });
	}
	return apiGet<RecurringSuggestionsResponse>('/cashflow/recurring/suggestions');
}

/** `POST /cashflow/recurring` */
export async function createRecurringItem(
	body: CreateRecurringItemRequest
): Promise<RecurringItem> {
	if (useMocks) {
		await delay();
		return {
			id: mockId(),
			name: body.name,
			direction: body.direction,
			rhythm: body.rhythm,
			lastAmountCents: 0,
			history: [],
			last_seen: null,
			next_expected: null,
			linked_count: body.ids.length,
			ended_from: null
		};
	}
	return apiSend<RecurringItem>('POST', '/cashflow/recurring', body);
}

/** `PATCH /cashflow/recurring/{item_id}` */
export async function updateRecurringItem(
	id: string,
	body: UpdateRecurringItemRequest
): Promise<RecurringItem> {
	if (useMocks) {
		await delay();
		const all = [...recurringExpenses, ...recurringIncome, ...recurringEnded];
		const item = all.find((candidate) => candidate.id === id) ?? all[0];
		return clone({ ...item, name: body.name, rhythm: body.rhythm });
	}
	return apiSend<RecurringItem>('PATCH', `/cashflow/recurring/${id}`, body);
}

/** `POST /cashflow/recurring/{item_id}/transactions` */
export async function linkRecurringTransactions(
	id: string,
	body: LinkRecurringTransactionsRequest
): Promise<RecurringMutationResponse> {
	if (useMocks) {
		await delay();
		return { linked_count: body.ids.length };
	}
	return apiSend<RecurringMutationResponse>('POST', `/cashflow/recurring/${id}/transactions`, body);
}

/** `DELETE /cashflow/recurring/{item_id}/transactions/{transaction_id}` */
export async function unlinkRecurringTransaction(id: string, transactionId: string): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiSend<unknown>('DELETE', `/cashflow/recurring/${id}/transactions/${transactionId}`);
}

/** `POST /cashflow/recurring/{item_id}/end` */
export async function endRecurringItem(
	id: string,
	body: EndRecurringItemRequest
): Promise<RecurringItem> {
	if (useMocks) {
		await delay();
		const all = [...recurringExpenses, ...recurringIncome, ...recurringEnded];
		const item = all.find((candidate) => candidate.id === id) ?? all[0];
		return clone({ ...item, next_expected: null, ended_from: body.from });
	}
	return apiSend<RecurringItem>('POST', `/cashflow/recurring/${id}/end`, body);
}

/** `POST /cashflow/recurring/suggestions/confirm` */
export async function confirmRecurringSuggestion(
	body: ConfirmRecurringSuggestionRequest
): Promise<RecurringItem> {
	if (useMocks) {
		await delay();
		const suggestion = recurringSuggestions.find((entry) => entry.match_key === body.match_key);
		return {
			id: mockId(),
			name: body.name,
			direction: body.direction,
			rhythm: body.rhythm,
			lastAmountCents: suggestion?.amountCents ?? 0,
			history: [],
			last_seen: null,
			next_expected: null,
			linked_count: suggestion?.matches ?? 0,
			ended_from: null
		};
	}
	return apiSend<RecurringItem>('POST', '/cashflow/recurring/suggestions/confirm', body);
}

/** `POST /cashflow/recurring/suggestions/dismiss` */
export async function dismissRecurringSuggestion(
	body: DismissRecurringSuggestionRequest
): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiSend<unknown>('POST', '/cashflow/recurring/suggestions/dismiss', body);
}
