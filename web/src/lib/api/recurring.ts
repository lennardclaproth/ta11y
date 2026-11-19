/**
 * Display conventions for recurring items (mirrors the Go backend's `cashflow.Rhythm`
 * and the "YYYY-MM" month an item is ended from).
 *
 * ta11y reports amounts as they were charged and never calls a rise expensive, so
 * nothing here turns a number into a verdict.
 */
import {
	addMonthsUTC,
	formatDisplayDate,
	monthLabel as calendarMonthLabel,
	startOfMonthUTC,
	todayISO
} from '$lib/components/molecules/calendar/calendar.utils';
import type { RecurringRhythm } from './types';

export const rhythmLabels: Record<RecurringRhythm, string> = {
	monthly: 'Monthly',
	quarterly: 'Quarterly',
	yearly: 'Yearly'
};

/** "Expense" / "Income" — the direction is always a word, never only a colour. */
export function recurringDirectionLabel(direction: 'in' | 'out'): string {
	return direction === 'in' ? 'Income' : 'Expense';
}

/** Whole days between today and an RFC3339 timestamp; negative once the day has passed. */
function daysUntil(timestamp: string): number {
	const target = Date.parse(timestamp.slice(0, 10));
	const today = Date.parse(todayISO());
	return Math.round((target - today) / 86_400_000);
}

/**
 * "18 Oct 2026 · in 9 days" — the date first, the distance as a reading aid. An item
 * with nothing linked yet, or one that has been ended, has no expectation at all.
 */
export function nextExpectedLabel(timestamp: string | null): string {
	if (!timestamp) return 'No expectation';
	const days = daysUntil(timestamp);
	const distance =
		days === 0
			? 'today'
			: days === 1
				? 'tomorrow'
				: days < 0
					? `${Math.abs(days)} days ago`
					: `in ${days} days`;
	return `${formatDisplayDate(timestamp.slice(0, 10))} · ${distance}`;
}

/** "October 2026" from a "YYYY-MM" month. */
export function endedFromLabel(month: string): string {
	const parsed = new Date(`${month}-01T00:00:00Z`);
	return Number.isNaN(parsed.getTime()) ? month : calendarMonthLabel(parsed);
}

/** The months an item can be ended from, starting at the current month. */
export function endMonthOptions(count = 6): { value: string; label: string }[] {
	const start = startOfMonthUTC(new Date(`${todayISO()}T00:00:00Z`));
	return Array.from({ length: count }, (_, index) => {
		const month = addMonthsUTC(start, index);
		return { value: month.toISOString().slice(0, 7), label: calendarMonthLabel(month) };
	});
}

/** "Jun '26" — short enough for a chart tick, unambiguous across years. */
export function recurringTickLabel(timestamp: string): string {
	const date = new Date(timestamp);
	if (Number.isNaN(date.getTime())) return timestamp;
	const month = date.toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });
	return `${month} '${String(date.getUTCFullYear()).slice(2)}`;
}

/** "Oct" — the month alone, for the overview chart where the year does not change. */
export function recurringMonthTick(timestamp: string): string {
	const date = new Date(timestamp);
	if (Number.isNaN(date.getTime())) return timestamp;
	return date.toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });
}
