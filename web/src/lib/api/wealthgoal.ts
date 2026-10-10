/**
 * Reading the monthly wealth goal [034] the way the backend means it.
 *
 * The API reports a month's income and contribution as 1e6-scaled amounts plus the goal
 * that applied in it; the share and the money still to go are derived here so every card
 * computes them the same way. Whether the month was met is the backend's verdict and is
 * never recomputed — it compares before rounding, and a displayed 30% can be 29.6%.
 */
import type { MonthResult, MonthStanding, TransactionPurpose } from './types';
import { scaledToNumber } from './money';

/** Share of marked income that went to wealth, rounded for display. */
export function sharePercent(month: MonthStanding): number {
	if (month.income_cents <= 0) return 0;
	return Math.round((month.contributed_cents / month.income_cents) * 100);
}

/** What the goal asks for in money, on the income marked in that month so far. */
export function goalAmount(month: MonthStanding): number {
	return (scaledToNumber(month.income_cents) * month.goal_percent) / 100;
}

/** What is still missing to reach the goal on the income marked so far. */
export function remainingAmount(month: MonthStanding): number {
	return Math.max(0, goalAmount(month) - scaledToNumber(month.contributed_cents));
}

/** "July 2026" for a "YYYY-MM-DD" first-of-month. */
export function monthLabel(month: string): string {
	return new Date(`${month}T00:00:00Z`).toLocaleDateString('en', {
		month: 'long',
		year: 'numeric',
		timeZone: 'UTC'
	});
}

/** "Jul 2026" — the short form the standing rows and the streak marks use. */
export function monthLabelShort(month: string): string {
	return new Date(`${month}T00:00:00Z`).toLocaleDateString('en', {
		month: 'short',
		year: 'numeric',
		timeZone: 'UTC'
	});
}

/** The result as a word, so it never depends on colour alone. */
export const monthResultLabel: Record<MonthResult, string> = {
	met: 'Met',
	missed: 'Missed',
	in_progress: 'In progress',
	not_scored: 'No goal yet'
};

/** The purpose as a word, for the ledger column and the bulk actions. */
export const purposeLabel: Record<TransactionPurpose, string> = {
	'': 'Not assigned',
	income: 'Income',
	wealth: 'To wealth'
};

/** The calendar month a "YYYY-MM-DD" first-of-month covers, as a date-range filter. */
export function monthRange(month: string): { from: string; to: string } {
	const start = new Date(`${month}T00:00:00Z`);
	const end = new Date(Date.UTC(start.getUTCFullYear(), start.getUTCMonth() + 1, 0));
	return { from: month, to: end.toISOString().slice(0, 10) };
}
