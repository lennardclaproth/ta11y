import { numberToScaled as s } from '$lib/api/money';
import type {
	RecurringAmountPoint,
	RecurringItem,
	RecurringRhythm,
	RecurringSuggestion,
	RecurringTransaction
} from '$lib/api/types';

/**
 * Deterministic recurring items for fixture mode. Invented counterparties and round amounts —
 * nothing here is real financial data. "Today" in these fixtures is 2026-10-09, so the expected
 * dates read sensibly against the cashflow fixtures from the same period.
 * `amountCents` is the 1e6-scaled magnitude; direction carries the sign.
 */

const rhythmStepMonths: Record<RecurringRhythm, number> = {
	monthly: 1,
	quarterly: 3,
	yearly: 12
};

/**
 * Dates the amounts were observed on: the last one is `lastSeen`, the earlier ones step back by
 * the rhythm. In the real API these come from the linked transactions.
 */
function history(
	lastSeen: string,
	rhythm: RecurringRhythm,
	amounts: number[]
): RecurringAmountPoint[] {
	const step = rhythmStepMonths[rhythm];
	const end = new Date(`${lastSeen}T00:00:00Z`);
	return amounts.map((amount, index) => {
		const date = new Date(end);
		date.setUTCMonth(date.getUTCMonth() - (amounts.length - 1 - index) * step);
		return { date: date.toISOString(), amountCents: s(amount) };
	});
}

function item(
	id: string,
	name: string,
	direction: 'in' | 'out',
	rhythm: RecurringRhythm,
	amounts: number[],
	lastSeen: string,
	nextExpected: string | null,
	linked: number,
	endedFrom: string | null = null
): RecurringItem {
	return {
		id,
		name,
		direction,
		rhythm,
		lastAmountCents: s(amounts[amounts.length - 1]),
		history: history(lastSeen, rhythm, amounts),
		last_seen: `${lastSeen}T00:00:00Z`,
		next_expected: nextExpected ? `${nextExpected}T00:00:00Z` : null,
		linked_count: linked,
		ended_from: endedFrom
	};
}

const monthly = [1400, 1400, 1400, 1450, 1450, 1450];

export const recurringExpenses: RecurringItem[] = [
	item('rc-001', 'Harbour Rentals', 'out', 'monthly', monthly, '2026-10-01', '2026-11-01', 18),
	// Energy moves between charges without anyone announcing a price rise.
	item(
		'rc-002',
		'Northwind Energy',
		'out',
		'monthly',
		[165, 190, 175, 185, 170, 180],
		'2026-09-28',
		'2026-10-28',
		14
	),
	// An amount that never moved, so its line sits on the middle rather than the floor.
	item(
		'rc-003',
		'Fiber Collective',
		'out',
		'monthly',
		[45, 45, 45, 45, 45, 45],
		'2026-09-24',
		'2026-10-24',
		12
	),
	item(
		'rc-004',
		'Lighthouse Gym',
		'out',
		'monthly',
		[25, 25, 30, 30, 30, 30],
		'2026-10-01',
		'2026-11-01',
		11
	),
	item(
		'rc-005',
		'Mobile Mast',
		'out',
		'monthly',
		[18, 18, 18, 18, 18, 18],
		'2026-09-20',
		'2026-10-20',
		12
	),
	// Grew from €9 to €12 over six charges — the case the whole page exists for.
	item(
		'rc-006',
		'Pixel Stream',
		'out',
		'monthly',
		[9, 9, 10, 10, 12, 12],
		'2026-09-18',
		'2026-10-18',
		15
	),
	item(
		'rc-007',
		'Anchor Insurance',
		'out',
		'quarterly',
		[90, 90, 96, 96],
		'2026-09-01',
		'2026-12-01',
		6
	),
	item('rc-008', 'Meridian Hosting', 'out', 'yearly', [100, 110, 120], '2026-03-12', '2027-03-12', 3)
];

export const recurringIncome: RecurringItem[] = [
	item(
		'rc-101',
		'ACME Corp payroll',
		'in',
		'monthly',
		[4000, 4000, 4200, 4200, 4200, 4200],
		'2026-09-25',
		'2026-10-25',
		18
	),
	item(
		'rc-102',
		'Tenant — Canal Street',
		'in',
		'monthly',
		[820, 820, 850, 850, 850, 850],
		'2026-10-01',
		'2026-11-01',
		16
	),
	item(
		'rc-103',
		'Northwind retainer',
		'in',
		'quarterly',
		[1200, 1500, 1500],
		'2026-10-05',
		'2027-01-05',
		3
	)
];

export const recurringEnded: RecurringItem[] = [
	item(
		'rc-201',
		'Quarterly Review (magazine)',
		'out',
		'monthly',
		[6, 6, 8, 8],
		'2026-07-14',
		null,
		9,
		'2026-08'
	),
	item('rc-202', 'Vault Backup', 'out', 'yearly', [50, 60], '2026-04-02', null, 2, '2026-05')
];

export const recurringSuggestions: RecurringSuggestion[] = [
	{
		match_key: 'cloudlocker sub',
		name: 'Cloudlocker Sub',
		direction: 'out',
		rhythm: 'monthly',
		amountCents: s(4),
		matches: 7,
		since: '2026-03-06T00:00:00Z',
		sample: 'CLOUDLOCKER*SUB 0312'
	},
	{
		match_key: 'city water board',
		name: 'City Water Board',
		direction: 'out',
		rhythm: 'quarterly',
		amountCents: s(64),
		matches: 4,
		since: '2025-10-15T00:00:00Z',
		sample: 'CITY WATER BOARD INCASSO'
	},
	{
		match_key: 'harbour fnd donatie',
		name: 'Harbour Fnd Donatie',
		direction: 'out',
		rhythm: 'monthly',
		amountCents: s(10),
		matches: 11,
		since: '2025-11-02T00:00:00Z',
		sample: 'ST. HARBOUR FND DONATIE'
	}
];

/**
 * Transactions of one item. The pair on 18 July is the same payment twice: an overlapping import
 * can produce that, and unlinking is how it is corrected.
 */
export const recurringLinkedTransactions: Record<string, RecurringTransaction[]> = {
	'rc-006': [
		{
			id: 'rx-1',
			date: '2026-09-18T00:00:00Z',
			description: 'PIXEL STREAM MONTHLY',
			amountCents: s(12),
			source: 'ing'
		},
		{
			id: 'rx-2',
			date: '2026-08-18T00:00:00Z',
			description: 'PIXEL STREAM MONTHLY',
			amountCents: s(12),
			source: 'ing'
		},
		{
			id: 'rx-3',
			date: '2026-07-18T00:00:00Z',
			description: 'PIXEL STREAM MONTHLY',
			amountCents: s(10),
			source: 'ing'
		},
		{
			id: 'rx-4',
			date: '2026-07-18T00:00:00Z',
			description: 'PIXEL STREAM MONTHLY',
			amountCents: s(10),
			source: 'ing'
		},
		{
			id: 'rx-5',
			date: '2026-06-18T00:00:00Z',
			description: 'PIXEL STREAM MONTHLY',
			amountCents: s(10),
			source: 'ing'
		},
		{
			id: 'rx-6',
			date: '2026-05-18T00:00:00Z',
			description: 'Pixel Stream',
			amountCents: s(9),
			source: 'manual'
		}
	]
};

/** Months the overview chart covers, oldest first. */
export const recurringSeriesMonths = [
	'2026-05-01',
	'2026-06-01',
	'2026-07-01',
	'2026-08-01',
	'2026-09-01',
	'2026-10-01'
];

/** Monthly-equivalent totals of the running items, matching the months above. */
export const recurringSeriesExpenses = [1700, 1725, 1740, 1752, 1765, 1777];
export const recurringSeriesIncome = [5220, 5220, 5220, 5550, 5550, 5550];
