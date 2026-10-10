import type { CashflowDirection } from './cashflow';

/**
 * Recurring cashflow items: subscriptions, fixed costs and recurring income.
 * Money fields are 1e6-scaled like the rest of cashflow (see api/money.ts).
 */

/** How often an item is expected. Mirrors `cashflow.Rhythm`. */
export type RecurringRhythm = 'monthly' | 'quarterly' | 'yearly';

/** One observed amount and the day it landed. Mirrors `cashflow.RecurringAmountPointResponse`. */
export interface RecurringAmountPoint {
	/** RFC3339 timestamp. */
	date: string;
	amountCents: number;
}

/** One recurring item. Mirrors `cashflow.RecurringItemResponse`. */
export interface RecurringItem {
	id: string;
	/** The name the user chose or confirmed — never derived silently. */
	name: string;
	direction: CashflowDirection;
	rhythm: RecurringRhythm;
	lastAmountCents: number;
	/** Observed amounts oldest to newest, shown as they are. */
	history: RecurringAmountPoint[];
	/** RFC3339 timestamp of the most recent linked transaction. */
	last_seen: string | null;
	/** RFC3339 timestamp; null once the item is ended or nothing is linked yet. */
	next_expected: string | null;
	linked_count: number;
	/** "YYYY-MM" the item was ended from; null while it is running. */
	ended_from: string | null;
}

/** The monthly-equivalent total of one month. Mirrors `cashflow.RecurringMonthPointResponse`. */
export interface RecurringMonthPoint {
	/** RFC3339 timestamp (first of month). */
	month: string;
	expenseCents: number;
	incomeCents: number;
}

/** `GET /cashflow/recurring` — mirrors `cashflow.RecurringOverviewResponse`. */
export interface RecurringOverviewResponse {
	expenses: RecurringItem[];
	income: RecurringItem[];
	ended: RecurringItem[];
	/** Quarterly amounts count as a third and yearly ones as a twelfth. */
	monthlyExpenseCents: number;
	monthlyIncomeCents: number;
	series: RecurringMonthPoint[];
}

/** One transaction linked to an item. Mirrors `cashflow.RecurringTransactionResponse`. */
export interface RecurringTransaction {
	id: string;
	/** RFC3339 timestamp. */
	date: string;
	description: string;
	amountCents: number;
	source: string;
}

/** `GET /cashflow/recurring/{item_id}` — mirrors `cashflow.RecurringItemDetailResponse`. */
export interface RecurringItemDetail extends RecurringItem {
	/** Newest first. */
	transactions: RecurringTransaction[];
}

/** One pattern found in the account's history. Mirrors `cashflow.RecurringSuggestionResponse`. */
export interface RecurringSuggestion {
	/** The fingerprint the pattern was grouped on; confirming or dismissing quotes it back. */
	match_key: string;
	name: string;
	direction: CashflowDirection;
	rhythm: RecurringRhythm;
	amountCents: number;
	matches: number;
	/** RFC3339 timestamp of the oldest matched transaction. */
	since: string;
	/** The statement text the suggested name was read off. */
	sample: string;
}

/** `GET /cashflow/recurring/suggestions` — mirrors `cashflow.RecurringSuggestionsResponse`. */
export interface RecurringSuggestionsResponse {
	data: RecurringSuggestion[];
}

/** `POST /cashflow/recurring` request — mirrors `cashflow.CreateRecurringItemRequest`. */
export interface CreateRecurringItemRequest {
	name: string;
	direction: CashflowDirection;
	rhythm: RecurringRhythm;
	ids: string[];
}

/** `PATCH /cashflow/recurring/{item_id}` request — mirrors `cashflow.UpdateRecurringItemRequest`. */
export interface UpdateRecurringItemRequest {
	name: string;
	rhythm: RecurringRhythm;
}

/** `POST /cashflow/recurring/{item_id}/transactions` request. */
export interface LinkRecurringTransactionsRequest {
	ids: string[];
}

/** `POST /cashflow/recurring/{item_id}/end` request. */
export interface EndRecurringItemRequest {
	/** "YYYY-MM". */
	from: string;
}

/** `POST /cashflow/recurring/suggestions/confirm` request. */
export interface ConfirmRecurringSuggestionRequest {
	match_key: string;
	name: string;
	direction: CashflowDirection;
	rhythm: RecurringRhythm;
}

/** `POST /cashflow/recurring/suggestions/dismiss` request. */
export interface DismissRecurringSuggestionRequest {
	match_key: string;
	direction: CashflowDirection;
}

/** Result of a link mutation. Mirrors `cashflow.RecurringMutationResponse`. */
export interface RecurringMutationResponse {
	linked_count: number;
}
