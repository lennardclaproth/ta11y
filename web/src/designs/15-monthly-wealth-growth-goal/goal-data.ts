/**
 * Obviously fake prototype data for design #15 (monthly wealth-growth goal).
 * Round numbers, invented names, no real financial data. Amounts are major EUR units
 * so the Money atom can render them directly; a real implementation converts the
 * 1e6-scaled API values with `$lib/api/money` first.
 */

export type GoalMonthStatus = 'met' | 'missed' | 'in-progress';

export interface GoalMonth {
	/** Calendar month, first of month. */
	month: string;
	label: string;
	/** Money you marked as income in that month. */
	income: number;
	/** Outgoing money you marked as a wealth contribution in that month. */
	contributed: number;
	/** The goal that applied in that month (a share of income, in percent). */
	goalPercent: number;
	/** Transactions in that month that have no purpose yet. */
	unassigned: number;
	status: GoalMonthStatus;
}

/** The goal that applies from the current month onwards. */
export const currentGoalPercent = 30;

/** Newest month first — the running month is never counted towards the streak. */
export const goalMonths: GoalMonth[] = [
	{
		month: '2026-07-01',
		label: 'July 2026',
		income: 5000,
		contributed: 900,
		goalPercent: 30,
		unassigned: 6,
		status: 'in-progress'
	},
	{
		month: '2026-06-01',
		label: 'June 2026',
		income: 5000,
		contributed: 1600,
		goalPercent: 30,
		unassigned: 0,
		status: 'met'
	},
	{
		month: '2026-05-01',
		label: 'May 2026',
		income: 5400,
		contributed: 1890,
		goalPercent: 30,
		unassigned: 2,
		status: 'met'
	},
	{
		month: '2026-04-01',
		label: 'April 2026',
		income: 5000,
		contributed: 1500,
		goalPercent: 30,
		unassigned: 0,
		status: 'met'
	},
	{
		month: '2026-03-01',
		label: 'March 2026',
		income: 5000,
		contributed: 1750,
		goalPercent: 30,
		unassigned: 0,
		status: 'met'
	},
	{
		month: '2026-02-01',
		label: 'February 2026',
		income: 4800,
		contributed: 960,
		goalPercent: 30,
		unassigned: 0,
		status: 'missed'
	},
	{
		month: '2026-01-01',
		label: 'January 2026',
		income: 4800,
		contributed: 1200,
		goalPercent: 25,
		unassigned: 0,
		status: 'met'
	}
];

export const currentStreak = 4;
export const bestStreak = 4;

/** Share of marked income that went to wealth, rounded to whole percent. */
export function sharePercent(month: GoalMonth): number {
	if (month.income <= 0) return 0;
	return Math.round((month.contributed / month.income) * 100);
}

/** What the goal asks for in money terms, for the month that is still running. */
export function goalAmount(month: GoalMonth): number {
	return Math.round((month.income * month.goalPercent) / 100);
}

export const monthStatusLabel: Record<GoalMonthStatus, string> = {
	met: 'Met',
	missed: 'Missed',
	'in-progress': 'In progress'
};

export type TransactionPurpose = 'income' | 'wealth' | null;

export interface GoalTransaction {
	id: string;
	date: string;
	description: string;
	tag: string;
	direction: 'in' | 'out';
	/** Magnitude in major EUR units; the sign lives in `direction`. */
	amount: number;
	purpose: TransactionPurpose;
}

/** A running month on Cashflow: two rows already pointed at, the rest still open. */
export const goalTransactions: GoalTransaction[] = [
	{
		id: 'px-01',
		date: '25 Jul 2026',
		description: 'Monthly salary',
		tag: 'salary',
		direction: 'in',
		amount: 5000,
		purpose: 'income'
	},
	{
		id: 'px-02',
		date: '26 Jul 2026',
		description: 'Transfer to savings',
		tag: 'savings',
		direction: 'out',
		amount: 600,
		purpose: 'wealth'
	},
	{
		id: 'px-03',
		date: '26 Jul 2026',
		description: 'Broker deposit',
		tag: 'investments',
		direction: 'out',
		amount: 300,
		purpose: 'wealth'
	},
	{
		id: 'px-04',
		date: '24 Jul 2026',
		description: 'Bullion purchase',
		tag: 'investments',
		direction: 'out',
		amount: 250,
		purpose: null
	},
	{
		id: 'px-05',
		date: '22 Jul 2026',
		description: 'Rent — July',
		tag: 'rent',
		direction: 'out',
		amount: 1450,
		purpose: null
	},
	{
		id: 'px-06',
		date: '20 Jul 2026',
		description: 'Supermarket',
		tag: 'groceries',
		direction: 'out',
		amount: 80,
		purpose: null
	},
	{
		id: 'px-07',
		date: '18 Jul 2026',
		description: 'Freelance invoice #51',
		tag: 'freelance',
		direction: 'in',
		amount: 1200,
		purpose: null
	},
	{
		id: 'px-08',
		date: '15 Jul 2026',
		description: 'Streaming service',
		tag: 'subscriptions',
		direction: 'out',
		amount: 10,
		purpose: null
	}
];

export const purposeLabel: Record<'income' | 'wealth', string> = {
	income: 'Income',
	wealth: 'To wealth'
};
