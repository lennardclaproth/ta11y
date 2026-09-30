/**
 * View-model for the #12 navigation and action-bar prototypes.
 *
 * Everything is derived from the committed mock fixtures, so the variants show the same
 * obviously-fake demo numbers the app shows in fixture mode. No new data is invented here.
 */
import { decimalStringToNumber, scaledToNumber } from '$lib/api/money';
import {
	assetSnapshots,
	cashflowMonthly,
	cashflowTagDistribution,
	cashflowTransactions,
	listings,
	portfolioPositions,
	portfolioSnapshots
} from '$lib/data/fixtures';
import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';
import type { NavItem } from '$lib/components/molecules/nav-menu/nav-menu.types';

export const account = { name: 'Demo account', email: 'demo@ta11y.example' };

/** The one period that lives in the overview and follows you between pages. */
export const period = {
	from: '2026-01-01',
	to: '2026-06-30',
	label: 'Jan 1 – Jun 30, 2026',
	shortLabel: 'Jan 1 – Jun 30'
};

const worthSeries = assetSnapshots.map((point) => decimalStringToNumber(point.total_worth));
const firstWorth = worthSeries[0];
const lastWorth = worthSeries[worthSeries.length - 1];

export const netWorth = lastWorth;
export const netWorthChange = lastWorth - firstWorth;
export const netWorthChangePct = Number((((lastWorth - firstWorth) / firstWorth) * 100).toFixed(2));
export const netWorthTrend = worthSeries;

export const navItems: NavItem[] = [
	{ label: 'Cashflow', href: '/cashflow', icon: 'heroicons:banknotes' },
	{ label: 'Assets', href: '/assets', icon: 'heroicons:building-library' },
	{ label: 'Portfolio', href: '/portfolio', icon: 'heroicons:chart-pie' },
	{ label: 'Listings', href: '/admin/listings', icon: 'heroicons:cog-6-tooth', divider: true },
	{ label: 'Dailies', href: '/admin/dailies', icon: 'heroicons:calendar-days' },
	{ label: 'Credentials', href: '/admin/credentials', icon: 'heroicons:key' }
];

export const euro = (value: number) =>
	`€${value.toLocaleString('en', { maximumFractionDigits: 0 })}`;

export const monthShort = (iso: string) =>
	new Date(`${iso.slice(0, 10)}T00:00:00Z`).toLocaleDateString('en', {
		month: 'short',
		timeZone: 'UTC'
	});

/* ---------- Cashflow ---------- */

export const cashflowLabels = cashflowMonthly.map((point) => point.month);
export const cashflowNet = cashflowMonthly.map((point) => scaledToNumber(point.net_cents));

const toDonut = (entries: { tag: string; totalCents: number }[]) =>
	entries.map((entry) => ({
		label: entry.tag || 'Untagged',
		value: scaledToNumber(entry.totalCents)
	}));

export const incomingByTag = toDonut(cashflowTagDistribution.incoming);
export const outgoingByTag = toDonut(cashflowTagDistribution.outgoing);

export const transactions = cashflowTransactions.filter((row) => !row.ignored).slice(0, 7);
export const transactionTotal = cashflowTransactions.filter((row) => !row.ignored).length;

export const tagOptions = cashflowTagDistribution.combined
	.map((entry) => entry.tag)
	.filter((tag) => tag !== '')
	.map((tag) => ({ value: tag, label: tag }));

/** Two preselected rows so the selection action is visible in the prototypes. */
export const selectedTransactionIds = transactions.slice(1, 3).map((row) => row.id);

/* ---------- Portfolio ---------- */

export const positions = portfolioPositions.filter((row) => !row.is_closed);

/** Only the snapshots inside the selected period, so the chart matches what the overview states. */
const periodSnapshots = portfolioSnapshots.filter((point) => {
	const day = point.occurred_at.slice(0, 10);
	return day >= period.from && day <= period.to;
});

export const portfolioLabels = periodSnapshots.map((point) => point.occurred_at.slice(0, 10));
export const portfolioMarketValue = periodSnapshots.map((point) =>
	scaledToNumber(point.market_value)
);
export const portfolioCostBasis = periodSnapshots.map((point) =>
	scaledToNumber(point.total_cost_basis)
);

const latestSnapshot = periodSnapshots[periodSnapshots.length - 1];

/* ---------- Admin ---------- */

/** Listings for the admin story: no ISIN column, the layout is what is being shown. */
export const adminListings = listings.slice(0, 6);

export const portfolioKpis: KpiItem[] = [
	{
		label: 'Market value',
		amount: scaledToNumber(latestSnapshot.market_value),
		currency: 'EUR',
		change: latestSnapshot.total_pnl_pct
	},
	{
		label: 'Total P&L',
		amount: scaledToNumber(latestSnapshot.total_pnl),
		currency: 'EUR',
		change: latestSnapshot.return_vs_cost_basis_pct
	},
	{
		label: 'Cost basis',
		amount: scaledToNumber(latestSnapshot.total_cost_basis),
		currency: 'EUR'
	}
];
