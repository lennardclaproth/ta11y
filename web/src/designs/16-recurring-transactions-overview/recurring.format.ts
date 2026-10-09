import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';

/** Fixed "today" for the prototypes, so screenshots stay deterministic. */
export const today = '2026-10-09';

const shortMonths = [
	'Jan',
	'Feb',
	'Mar',
	'Apr',
	'May',
	'Jun',
	'Jul',
	'Aug',
	'Sep',
	'Oct',
	'Nov',
	'Dec'
];

function utc(iso: string): Date {
	return new Date(`${iso}T00:00:00Z`);
}

/** Whole days between `today` and an ISO date; negative when the date has passed. */
export function daysUntil(iso: string): number {
	return Math.round((utc(iso).getTime() - utc(today).getTime()) / 86_400_000);
}

/** "18 Oct 2026 · in 9 days" — the date first, the distance as a reading aid. */
export function nextExpectedLabel(iso: string | null): string {
	if (!iso) return 'No expectation';
	const days = daysUntil(iso);
	const distance = days === 0 ? 'today' : days === 1 ? 'tomorrow' : `in ${days} days`;
	return `${formatDisplayDate(iso)} · ${distance}`;
}

/** "October 2026" */
export function monthLabel(iso: string): string {
	const date = utc(iso);
	return `${
		[
			'January',
			'February',
			'March',
			'April',
			'May',
			'June',
			'July',
			'August',
			'September',
			'October',
			'November',
			'December'
		][date.getUTCMonth()]
	} ${date.getUTCFullYear()}`;
}

/** "2026-10" */
export function monthKey(iso: string): string {
	return iso.slice(0, 7);
}

/** Day number of an ISO date, e.g. 18. */
export function dayOfMonth(iso: string): number {
	return utc(iso).getUTCDate();
}

/** "Oct" */
export function shortMonth(iso: string): string {
	return shortMonths[utc(iso).getUTCMonth()];
}

/** "Aug 2026" — used to date the start of an amount history. */
export function monthYearShort(iso: string): string {
	const date = utc(iso);
	return `${shortMonths[date.getUTCMonth()]} ${date.getUTCFullYear()}`;
}
