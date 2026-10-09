/**
 * Fake data for the #16 design prototypes only. Invented vendors, round amounts, no real
 * financial data. Amounts are major units (EUR) so the Money atom can render them directly.
 * "Today" in these fixtures is 2026-10-09.
 */

export type RecurringRhythm = 'monthly' | 'quarterly' | 'yearly';
export type RecurringDirection = 'in' | 'out';

export interface RecurringItem {
	id: string;
	/** Who you pay, or who pays you. Chosen or confirmed by the user, never derived silently. */
	name: string;
	direction: RecurringDirection;
	rhythm: RecurringRhythm;
	/** Most recent observed amount, major units. */
	amount: number;
	/** Observed amounts, oldest to newest — shown as-is, never labelled "more expensive". */
	history: number[];
	/** ISO date of the next expected transaction; null when the rhythm is irregular. */
	nextExpected: string | null;
	/** ISO date of the most recent linked transaction. */
	lastSeen: string;
	/** Number of linked transactions. */
	linked: number;
	/** Month the item was ended from, "YYYY-MM"; absent while running. */
	endedFrom?: string;
}

export interface RecurringSuggestion {
	id: string;
	name: string;
	direction: RecurringDirection;
	rhythm: RecurringRhythm;
	amount: number;
	/** Transactions ta11y matched for this suggestion. */
	matches: number;
	/** ISO date of the oldest matched transaction. */
	since: string;
	/** Statement text the suggested name came from. */
	sample: string;
}

export const rhythmLabels: Record<RecurringRhythm, string> = {
	monthly: 'Monthly',
	quarterly: 'Quarterly',
	yearly: 'Yearly'
};

export const recurringExpenses: RecurringItem[] = [
	{
		id: 'rc-001',
		name: 'Harbour Rentals',
		direction: 'out',
		rhythm: 'monthly',
		amount: 1450,
		history: [1400, 1400, 1400, 1450, 1450, 1450],
		nextExpected: '2026-11-01',
		lastSeen: '2026-10-01',
		linked: 18
	},
	{
		id: 'rc-002',
		name: 'Northwind Energy',
		direction: 'out',
		rhythm: 'monthly',
		amount: 180,
		history: [165, 190, 175, 185, 170, 180],
		nextExpected: '2026-10-28',
		lastSeen: '2026-09-28',
		linked: 14
	},
	{
		id: 'rc-003',
		name: 'Fiber Collective',
		direction: 'out',
		rhythm: 'monthly',
		amount: 45,
		history: [45, 45, 45, 45, 45, 45],
		nextExpected: '2026-10-24',
		lastSeen: '2026-09-24',
		linked: 12
	},
	{
		id: 'rc-004',
		name: 'Lighthouse Gym',
		direction: 'out',
		rhythm: 'monthly',
		amount: 30,
		history: [25, 25, 30, 30, 30, 30],
		nextExpected: '2026-11-01',
		lastSeen: '2026-10-01',
		linked: 11
	},
	{
		id: 'rc-005',
		name: 'Mobile Mast',
		direction: 'out',
		rhythm: 'monthly',
		amount: 18,
		history: [18, 18, 18, 18, 18, 18],
		nextExpected: '2026-10-20',
		lastSeen: '2026-09-20',
		linked: 12
	},
	{
		id: 'rc-006',
		name: 'Pixel Stream',
		direction: 'out',
		rhythm: 'monthly',
		amount: 12,
		history: [9, 9, 10, 10, 12, 12],
		nextExpected: '2026-10-18',
		lastSeen: '2026-09-18',
		linked: 15
	},
	{
		id: 'rc-007',
		name: 'Anchor Insurance',
		direction: 'out',
		rhythm: 'quarterly',
		amount: 96,
		history: [90, 90, 96, 96],
		nextExpected: '2026-12-01',
		lastSeen: '2026-09-01',
		linked: 6
	},
	{
		id: 'rc-008',
		name: 'Meridian Hosting',
		direction: 'out',
		rhythm: 'yearly',
		amount: 120,
		history: [100, 110, 120],
		nextExpected: '2027-03-12',
		lastSeen: '2026-03-12',
		linked: 3
	}
];

export const recurringIncome: RecurringItem[] = [
	{
		id: 'rc-101',
		name: 'ACME Corp payroll',
		direction: 'in',
		rhythm: 'monthly',
		amount: 4200,
		history: [4000, 4000, 4200, 4200, 4200, 4200],
		nextExpected: '2026-10-25',
		lastSeen: '2026-09-25',
		linked: 18
	},
	{
		id: 'rc-102',
		name: 'Tenant — Canal Street',
		direction: 'in',
		rhythm: 'monthly',
		amount: 850,
		history: [820, 820, 850, 850, 850, 850],
		nextExpected: '2026-11-01',
		lastSeen: '2026-10-01',
		linked: 16
	},
	{
		id: 'rc-103',
		name: 'Northwind retainer',
		direction: 'in',
		rhythm: 'quarterly',
		amount: 1500,
		history: [1200, 1500, 1500],
		nextExpected: '2027-01-05',
		lastSeen: '2026-10-05',
		linked: 3
	}
];

export const recurringEnded: RecurringItem[] = [
	{
		id: 'rc-201',
		name: 'Quarterly Review (magazine)',
		direction: 'out',
		rhythm: 'monthly',
		amount: 8,
		history: [6, 6, 8, 8],
		nextExpected: null,
		lastSeen: '2026-07-14',
		linked: 9,
		endedFrom: '2026-08'
	},
	{
		id: 'rc-202',
		name: 'Vault Backup',
		direction: 'out',
		rhythm: 'yearly',
		amount: 60,
		history: [50, 60],
		nextExpected: null,
		lastSeen: '2026-04-02',
		linked: 2,
		endedFrom: '2026-05'
	}
];

export const recurringSuggestions: RecurringSuggestion[] = [
	{
		id: 'sg-001',
		name: 'Cloud Locker',
		direction: 'out',
		rhythm: 'monthly',
		amount: 4,
		matches: 7,
		since: '2026-03-06',
		sample: 'CLOUDLOCKER*SUB 0312'
	},
	{
		id: 'sg-002',
		name: 'City Water',
		direction: 'out',
		rhythm: 'quarterly',
		amount: 64,
		matches: 4,
		since: '2025-10-15',
		sample: 'CITY WATER BOARD INCASSO'
	},
	{
		id: 'sg-003',
		name: 'Harbour Foundation',
		direction: 'out',
		rhythm: 'monthly',
		amount: 10,
		matches: 11,
		since: '2025-11-02',
		sample: 'ST. HARBOUR FND DONATIE'
	}
];

/** Monthly-equivalent totals of the running items (quarterly ÷ 3, yearly ÷ 12). */
export const monthlyExpenseTotal = 1777;
export const monthlyIncomeTotal = 5550;

/** Total of the running recurring expenses per month, oldest to newest (May – Oct 2026). */
export const expenseTrendMonths = [
	'2026-05-01',
	'2026-06-01',
	'2026-07-01',
	'2026-08-01',
	'2026-09-01',
	'2026-10-01'
];
export const expenseTrendTotals = [1700, 1725, 1740, 1752, 1765, 1777];

/** Transactions of one item, for the "open a post" detail in the prototypes. */
export const sampleItemTransactions = [
	{ id: 'tx-1', date: '2026-09-18', description: 'PIXEL STREAM MONTHLY', amount: 12 },
	{ id: 'tx-2', date: '2026-08-18', description: 'PIXEL STREAM MONTHLY', amount: 12 },
	{ id: 'tx-3', date: '2026-07-18', description: 'PIXEL STREAM MONTHLY', amount: 10 },
	{ id: 'tx-4', date: '2026-06-18', description: 'PIXEL STREAM MONTHLY', amount: 10 }
];
