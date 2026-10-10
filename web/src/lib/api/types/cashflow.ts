import type { PaginatedResponse } from './common';

/** Cashflow direction. */
export type CashflowDirection = 'in' | 'out';

/**
 * One cashflow transaction. Mirrors `cashflow.CreateTransactionResponse`.
 * `amountCents` is 1e6-scaled (see api/money.ts), not hundredths.
 */
export interface CashflowTransaction {
	id: string;
	description: string;
	note: string;
	source: string;
	amountCents: number;
	direction: CashflowDirection;
	/** RFC3339 timestamp. */
	date: string;
	tag: string;
	ignored: boolean;
	/**
	 * The ignore rule that ignored this row, null when nobody's rule did. The rule's
	 * name is resolved from the account's rules rather than repeated on every row.
	 */
	ignored_by_rule_id?: string | null;
	/** The ignored state was decided by hand, so no rule will change it again. */
	ignore_overridden?: boolean;
	/** The import this row arrived with, null for a manual entry. */
	import_id?: string | null;
}

/** `GET /cashflow/transactions` — mirrors `cashflow.GetTransactionsResponse`. */
export type CashflowTransactionsResponse = PaginatedResponse<CashflowTransaction>;

/** One month of aggregated cashflow. Mirrors `cashflow.MonthlyAnalyticsPointResponse`. */
export interface CashflowMonthlyPoint {
	/** "YYYY-MM-DD" (first of month). */
	month: string;
	incoming_cents: number;
	outgoing_cents: number;
	net_cents: number;
}

/** `GET /cashflow/analytics/monthly` — mirrors `cashflow.CashflowMonthlyAnalyticsResponse`. */
export interface CashflowMonthlyAnalyticsResponse {
	data: CashflowMonthlyPoint[];
}

/** One tag total. Mirrors `cashflow.TagDistributionEntryResponse`. */
export interface TagDistributionEntry {
	tag: string;
	totalCents: number;
}

/** `GET /cashflow/analytics/tags` — mirrors `cashflow.TagDistributionResponse`. */
export interface TagDistributionResponse {
	combined: TagDistributionEntry[];
	incoming: TagDistributionEntry[];
	outgoing: TagDistributionEntry[];
}

/** One manual cashflow transaction to create. Mirrors `cashflow.CreateManualCashflowTransactionRequest`. */
export interface CreateManualCashflowTransaction {
	/** "YYYY-MM-DD". */
	date: string;
	/** Non-negative decimal string. */
	amount: string;
	/** Direction the backend accepts: `income`/`in` or `expense`/`out`. */
	type: 'income' | 'expense' | 'in' | 'out';
	description: string;
	note: string;
	tag: string;
	vendor?: string;
}

/** `POST /cashflow/transactions/manual` request — mirrors `cashflow.CreateTransactionsRequest`. */
export interface CreateCashflowTransactionsRequest {
	transactions: CreateManualCashflowTransaction[];
}

/** `POST /cashflow/transactions/manual` — mirrors `cashflow.TransactionsResponse`. */
export interface CreateCashflowTransactionsResponse {
	created_count: number;
	data: CashflowTransaction[];
}

/**
 * `POST /cashflow/transactions/date` request — mirrors `cashflow.ChangeTransactionDateRequest`.
 * Only manually entered transactions can be moved, and only to today or earlier.
 */
export interface ChangeCashflowTransactionDateRequest {
	id: string;
	/** "YYYY-MM-DD". */
	date: string;
}

/** `POST /cashflow/transactions/date` — mirrors `cashflow.ChangeTransactionDateResponse`. */
export interface ChangeCashflowTransactionDateResponse {
	id: string;
	/** RFC3339 timestamp. */
	date: string;
}

/** Shared bulk-mutation filter body. Mirrors `cashflow.TransactionFilters`. */
export interface CashflowTransactionFilters {
	q?: string;
	description?: string;
	note?: string;
	source?: string;
	direction?: string;
	/** comma-separated tags */
	tags?: string;
	untagged?: boolean;
	hide_ignored?: boolean;
	/** Only the rows one import brought in. */
	import_id?: string;
	/** Only the rows one ignore rule ignored. */
	ignored_by_rule?: string;
	from?: string;
	to?: string;
}

/** Which transaction field an ignore rule reads. */
export type IgnoreRuleMatchField = 'description' | 'note';

/**
 * One ignore rule. It recognises transactions that should arrive ignored: text in the
 * description or note, optionally narrowed to a direction and to the bank a row was
 * imported from. Mirrors `cashflow.IgnoreRuleResponse`.
 */
export interface IgnoreRule {
	id: string;
	name: string;
	match_field: IgnoreRuleMatchField;
	contains: string;
	/** Empty for a rule that matches both directions. */
	direction: '' | CashflowDirection;
	/** The bank the rule is limited to, empty for every bank. */
	source: string;
	enabled: boolean;
	/** How many transactions this rule has ignored since it was made. */
	ignored_total: number;
	last_applied_at: string | null;
	created_at: string;
	updated_at: string;
}

/** `GET /cashflow/ignore-rules` — mirrors `cashflow.IgnoreRulesResponse`. */
export interface IgnoreRulesResponse {
	data: IgnoreRule[];
}

/** Body of a create, update, or preview of an ignore rule. */
export interface IgnoreRuleRequest {
	name: string;
	match_field: IgnoreRuleMatchField;
	contains: string;
	direction?: '' | CashflowDirection;
	source?: string;
	enabled?: boolean;
}

/** `POST /cashflow/ignore-rules/preview` — mirrors `cashflow.IgnoreRulePreviewResponse`. */
export interface IgnoreRulePreview {
	/** Every transaction in the ledger the rule matches. */
	matching: number;
	/** What applying it would ignore: matching, not ignored yet, not decided by hand. */
	not_yet_ignored: number;
	/** The size of the ledger the rule was held against. */
	scanned: number;
	sample: CashflowTransaction[];
}

/** `POST /cashflow/ignore-rules/{rule_id}/apply` — mirrors `cashflow.ApplyIgnoreRuleResponse`. */
export interface ApplyIgnoreRuleResponse {
	ignored_count: number;
	status: string;
}

/** The harvest of one rule within one import. Mirrors `cashflow.IgnoredRuleGroupResponse`. */
export interface IgnoredRuleGroup {
	rule: IgnoreRule;
	total: number;
	transactions: CashflowTransaction[];
}

/** `GET /cashflow/imports/{import_id}/ignored` — mirrors `cashflow.ImportIgnoredResponse`. */
export interface ImportIgnoredResponse {
	data: IgnoredRuleGroup[];
}

/** `POST /cashflow/transactions/tag` request (single). */
export interface TagTransactionRequest {
	id: string;
	tag: string;
}

/** `POST /cashflow/transactions/tag/selection` request. */
export interface TagTransactionsBySelectionRequest {
	tag: string;
	ids: string[];
}

/** `POST /cashflow/transactions/tag/filter` request. */
export interface TagTransactionsByFilterRequest {
	tag: string;
	filters: CashflowTransactionFilters;
}

/** `POST /cashflow/transactions/ignore/selection` request. */
export interface IgnoreTransactionsBySelectionRequest {
	/** Defaults to true on the backend when omitted. */
	ignored?: boolean;
	ids: string[];
}

/** `POST /cashflow/transactions/ignore/filter` request. */
export interface IgnoreTransactionsByFilterRequest {
	ignored?: boolean;
	filters: CashflowTransactionFilters;
}

/** Result of a bulk tag/ignore mutation. */
export interface CashflowBulkMutationResponse {
	updated_count: number;
	status: string;
}

/** Query filters for `GET /cashflow/transactions`. */
export interface CashflowTransactionsQuery {
	limit?: number;
	offset?: number;
	sort_by?: 'date' | 'description' | 'note' | 'tag' | 'source' | 'amount';
	sort_order?: 'asc' | 'desc';
	q?: string;
	description?: string;
	note?: string;
	source?: string;
	direction?: CashflowDirection;
	/** comma-separated tags */
	tags?: string;
	untagged?: boolean;
	hide_ignored?: boolean;
	/** Only the rows one import brought in. */
	import_id?: string;
	/** Only the rows one ignore rule ignored. */
	ignored_by_rule?: string;
	from?: string;
	to?: string;
}
