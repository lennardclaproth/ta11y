/**
 * Fake ledger rows for the issue #13 design prototypes. Invented descriptions and round amounts —
 * nothing here comes from a real account, statement or import.
 */
import { numberToScaled } from '$lib/api/money';
import type { CashflowTransaction } from '$lib/api/types';

/** The row the prototypes backdate: entered today, but it happened months ago. */
export const backdatedDate = '2026-07-14';

export const designTransactions: CashflowTransaction[] = [
	{
		id: 'design-1',
		description: 'Bike repair',
		note: 'Entered late',
		source: 'manual',
		amountCents: numberToScaled(120),
		direction: 'out',
		date: '2026-09-30T10:00:00Z',
		tag: 'household',
		ignored: false
	},
	{
		id: 'design-2',
		description: 'Monthly salary',
		note: 'Payroll',
		source: 'ing',
		amountCents: numberToScaled(3000),
		direction: 'in',
		date: '2026-09-25T08:00:00Z',
		tag: 'salary',
		ignored: false
	},
	{
		id: 'design-3',
		description: 'Coffee subscription',
		note: 'Monthly',
		source: 'manual',
		amountCents: numberToScaled(20),
		direction: 'out',
		date: '2026-09-21T09:00:00Z',
		tag: 'subscriptions',
		ignored: false
	},
	{
		id: 'design-4',
		description: 'Train ticket',
		note: 'Commute',
		source: 'ing',
		amountCents: numberToScaled(30),
		direction: 'out',
		date: '2026-09-18T07:00:00Z',
		tag: 'transport',
		ignored: false
	},
	{
		id: 'design-5',
		description: 'Rent',
		note: 'Apartment',
		source: 'ing',
		amountCents: numberToScaled(1200),
		direction: 'out',
		date: '2026-09-10T06:00:00Z',
		tag: 'rent',
		ignored: false
	},
	{
		id: 'design-6',
		description: 'Concert tickets',
		note: 'Split with friends',
		source: 'manual',
		amountCents: numberToScaled(90),
		direction: 'out',
		date: '2026-09-04T19:00:00Z',
		tag: 'leisure',
		ignored: false
	}
];

/** Only manually entered rows may have their date changed afterwards. */
export function isManual(row: CashflowTransaction): boolean {
	return row.source === 'manual';
}

/** Short label for the origin of a row. */
export function sourceLabel(row: CashflowTransaction): string {
	return isManual(row) ? 'Manual' : row.source.toUpperCase();
}

/** Fake portfolio rows, used only for the "portfolio is still rebuilding" page state. */
export type DesignPortfolioRow = {
	id: string;
	occurredAt: string;
	listing: string;
	side: 'Buy' | 'Sell';
	quantity: number;
	amount: number;
	origin: 'MANUAL' | 'IMPORT';
};

export const designPortfolioRows: DesignPortfolioRow[] = [
	{
		id: 'pf-1',
		occurredAt: '2026-07-14',
		listing: 'NORTHWIND INDEX FUND',
		side: 'Buy',
		quantity: 10,
		amount: 900,
		origin: 'MANUAL'
	},
	{
		id: 'pf-2',
		occurredAt: '2026-09-22',
		listing: 'HARBOUR WORLD ETF',
		side: 'Buy',
		quantity: 4,
		amount: 400,
		origin: 'IMPORT'
	},
	{
		id: 'pf-3',
		occurredAt: '2026-09-08',
		listing: 'MERIDIAN BOND FUND',
		side: 'Sell',
		quantity: 6,
		amount: 300,
		origin: 'IMPORT'
	},
	{
		id: 'pf-4',
		occurredAt: '2026-08-19',
		listing: 'NORTHWIND INDEX FUND',
		side: 'Buy',
		quantity: 5,
		amount: 450,
		origin: 'MANUAL'
	}
];

/** "78 days ago" style distance, used as a plain-language check on a backdated date. */
export function daysAgoLabel(iso: string, today: string): string {
	const from = Date.parse(`${iso}T00:00:00Z`);
	const to = Date.parse(`${today}T00:00:00Z`);
	if (Number.isNaN(from) || Number.isNaN(to)) return '';
	const days = Math.round((to - from) / 86_400_000);
	if (days <= 0) return 'Today';
	if (days === 1) return 'Yesterday';
	return `${days} days ago`;
}
