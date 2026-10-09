/**
 * Obviously fake data for the #23 ignore-rules design prototypes. Round numbers, invented
 * descriptions, no real account identifiers. Never promote this file out of designs/.
 */
import { numberToScaled as s } from '$lib/api/money';
import type { CashflowTransaction } from '$lib/api/types';

/** Which transaction field a rule reads. */
export type RuleField = 'description' | 'note';

/** Direction a rule limits itself to. `any` matches both. */
export type RuleDirection = 'any' | 'in' | 'out';

export interface IgnoreRule {
	id: string;
	name: string;
	field: RuleField;
	contains: string;
	direction: RuleDirection;
	/**
	 * Which bank the rule is limited to, or null for every one. A cashflow transaction has
	 * `source` (the vendor it was imported from), not a separate account, so that is what a
	 * rule can honestly scope on today — see the open question in the design comment.
	 */
	source: string | null;
	enabled: boolean;
	/** How many transactions this rule has ignored since it was made. */
	ignoredTotal: number;
	/** How many it caught in the most recent import. */
	ignoredLastImport: number;
	lastUsed: string;
}

export const sourceLabels = ['ING', 'N26'] as const;

export const ignoreRules: IgnoreRule[] = [
	{
		id: 'rule-1',
		name: 'Transfer to savings',
		field: 'description',
		contains: 'Transfer to savings',
		direction: 'out',
		source: null,
		enabled: true,
		ignoredTotal: 48,
		ignoredLastImport: 12,
		lastUsed: '25 Jun 2026'
	},
	{
		id: 'rule-2',
		name: 'Credit card payment',
		field: 'description',
		contains: 'Credit card',
		direction: 'out',
		source: 'ING',
		enabled: true,
		ignoredTotal: 19,
		ignoredLastImport: 9,
		lastUsed: '25 Jun 2026'
	},
	{
		id: 'rule-3',
		name: 'Own transfer to N26',
		field: 'description',
		contains: 'To N26 Everyday',
		direction: 'out',
		source: 'ING',
		enabled: true,
		ignoredTotal: 31,
		ignoredLastImport: 10,
		lastUsed: '25 Jun 2026'
	},
	{
		id: 'rule-4',
		name: 'Incoming own transfer',
		field: 'note',
		contains: 'Own account',
		direction: 'in',
		source: null,
		enabled: true,
		ignoredTotal: 27,
		ignoredLastImport: 6,
		lastUsed: '25 Jun 2026'
	},
	{
		id: 'rule-5',
		name: 'Round-up savings',
		field: 'description',
		contains: 'Round-up',
		direction: 'out',
		source: 'N26',
		enabled: false,
		ignoredTotal: 12,
		ignoredLastImport: 0,
		lastUsed: '20 May 2026'
	}
];

export function ruleById(id: string): IgnoreRule {
	return ignoreRules.find((rule) => rule.id === id) ?? ignoreRules[0];
}

/** Human-readable summary of what a rule matches on, for a single ruled line. */
export function ruleSummary(rule: IgnoreRule): string {
	const field = rule.field === 'description' ? 'Description' : 'Note';
	const direction =
		rule.direction === 'any' ? 'In and out' : rule.direction === 'in' ? 'Incoming' : 'Outgoing';
	return `${field} contains “${rule.contains}” · ${direction} · ${rule.source ?? 'All banks'}`;
}

/**
 * A transaction plus the rule that ignored it, or `null` when it was ignored by hand.
 * `restored` marks a row the person put back themselves; rules leave it alone from then on.
 */
export type IgnoredTransaction = CashflowTransaction & {
	ruleId: string | null;
	restored?: boolean;
};

export const ignoredTransactions: IgnoredTransaction[] = [
	{
		id: 'ig-01',
		description: 'Transfer to savings',
		note: 'Own account',
		source: 'ing',
		amountCents: s(500),
		direction: 'out',
		date: '2026-06-25T08:00:00Z',
		tag: 'savings',
		ignored: true,
		ruleId: 'rule-1'
	},
	{
		id: 'ig-02',
		description: 'Credit card payment',
		note: 'Monthly settlement',
		source: 'ing',
		amountCents: s(820),
		direction: 'out',
		date: '2026-06-24T06:00:00Z',
		tag: '',
		ignored: true,
		ruleId: 'rule-2'
	},
	{
		id: 'ig-03',
		description: 'To N26 Everyday',
		note: 'Own account',
		source: 'ing',
		amountCents: s(300),
		direction: 'out',
		date: '2026-06-23T09:30:00Z',
		tag: '',
		ignored: true,
		ruleId: 'rule-3'
	},
	{
		id: 'ig-04',
		description: 'From ING Current',
		note: 'Own account',
		source: 'n26',
		amountCents: s(300),
		direction: 'in',
		date: '2026-06-23T09:31:00Z',
		tag: '',
		ignored: true,
		ruleId: 'rule-4'
	},
	{
		id: 'ig-05',
		description: 'Credit card annual fee',
		note: 'Card costs',
		source: 'ing',
		amountCents: s(30),
		direction: 'out',
		date: '2026-06-22T06:00:00Z',
		tag: 'fees',
		ignored: false,
		ruleId: 'rule-2',
		restored: true
	},
	{
		id: 'ig-06',
		description: 'Transfer to savings',
		note: 'Own account',
		source: 'ing',
		amountCents: s(250),
		direction: 'out',
		date: '2026-06-20T08:00:00Z',
		tag: 'savings',
		ignored: true,
		ruleId: 'rule-1'
	},
	{
		id: 'ig-07',
		description: 'To N26 Everyday',
		note: 'Own account',
		source: 'ing',
		amountCents: s(150),
		direction: 'out',
		date: '2026-06-18T09:30:00Z',
		tag: '',
		ignored: true,
		ruleId: 'rule-3'
	},
	{
		id: 'ig-08',
		description: 'Annual club membership',
		note: 'Not an own transfer',
		source: 'n26',
		amountCents: s(120),
		direction: 'out',
		date: '2026-05-12T11:00:00Z',
		tag: 'health',
		ignored: true,
		ruleId: null
	}
];

/** Rows the Cashflow ledger shows while "ignored rows" are switched on. */
export const ledgerRows: IgnoredTransaction[] = [
	{
		id: 'tx-01',
		description: 'Salary June',
		note: '',
		source: 'ing',
		amountCents: s(3200),
		direction: 'in',
		date: '2026-06-25T07:00:00Z',
		tag: 'salary',
		ignored: false,
		ruleId: null
	},
	...ignoredTransactions.slice(0, 4),
	{
		id: 'tx-02',
		description: 'Groceries week 26',
		note: '',
		source: 'n26',
		amountCents: s(96),
		direction: 'out',
		date: '2026-06-22T17:20:00Z',
		tag: 'groceries',
		ignored: false,
		ruleId: null
	},
	ignoredTransactions[4],
	ignoredTransactions[5]
];

/** Counts of the most recent cashflow import. */
export const lastImport = {
	file: 'statement-june.csv',
	bank: 'ING',
	finishedAgo: '2 minutes ago',
	totalRows: 207,
	imported: 128,
	duplicates: 42,
	autoIgnored: 37
};

/** Rows the editor shows as a sample of what a rule would catch. */
export const previewMatches: CashflowTransaction[] = [
	{
		id: 'pv-01',
		description: 'Credit card payment',
		note: 'Monthly settlement',
		source: 'ing',
		amountCents: s(820),
		direction: 'out',
		date: '2026-06-24T06:00:00Z',
		tag: '',
		ignored: false
	},
	{
		id: 'pv-02',
		description: 'Credit card payment',
		note: 'Monthly settlement',
		source: 'ing',
		amountCents: s(640),
		direction: 'out',
		date: '2026-05-24T06:00:00Z',
		tag: '',
		ignored: false
	},
	{
		id: 'pv-03',
		description: 'Credit card payment',
		note: 'Monthly settlement',
		source: 'ing',
		amountCents: s(710),
		direction: 'out',
		date: '2026-04-24T06:00:00Z',
		tag: '',
		ignored: false
	},
	{
		id: 'pv-04',
		description: 'Credit card annual fee',
		note: 'Card costs',
		source: 'ing',
		amountCents: s(30),
		direction: 'out',
		date: '2026-04-02T06:00:00Z',
		tag: 'fees',
		ignored: false
	}
];

/** What the editor reports above the preview. */
export const previewSummary = { matching: 23, scanned: 1284, notYetIgnored: 23 };
