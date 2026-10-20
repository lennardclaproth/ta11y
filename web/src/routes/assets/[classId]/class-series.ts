import { decimalStringToNumber, scaledToNumber } from '$lib/api/money';
import { addDaysISO } from '$lib/components/molecules/calendar/calendar.utils';
import type { AssetHolding, AssetMutation, PortfolioSnapshotPoint } from '$lib/api/types';

/** One day of worth against what had been paid in by then. */
export type WorthPoint = { date: string; value: number; paid: number };

/** The figures every item and every class reports for the selected period. */
export type WorthFigures = {
	worth: number;
	paid: number;
	unrealized: number;
	unrealizedPct: number | null;
	/** Change in worth over the period, without the money that was added or taken out. */
	growth: number;
	growthPct: number | null;
};

/**
 * The worth history of an item whose worth is set by hand. The opening worth counts as paid in,
 * money added or withdrawn moves paid in with it, and a later "set worth" is the item growing:
 * savings that earn interest, a house that is revalued.
 */
export function manualItemPoints(mutations: AssetMutation[], itemId: string): WorthPoint[] {
	const rows = mutations
		.filter((m) => m.item_id === itemId)
		.sort(
			(a, b) =>
				a.effective_date.localeCompare(b.effective_date) || a.created_at.localeCompare(b.created_at)
		);
	const points: WorthPoint[] = [];
	let paid = 0;
	for (const [index, row] of rows.entries()) {
		const value = decimalStringToNumber(row.new_worth);
		if (index === 0) {
			paid = value;
		} else if (row.change_type.toUpperCase() === 'ADJUST') {
			const amount = decimalStringToNumber(row.amount);
			paid += row.direction?.toUpperCase() === 'DECREASE' ? -amount : amount;
		}
		points.push({ date: row.effective_date.slice(0, 10), value, paid });
	}
	return points;
}

export function holdingPoints(holding: AssetHolding): WorthPoint[] {
	return holding.series.map((p) => ({
		date: p.date,
		value: decimalStringToNumber(p.value),
		paid: decimalStringToNumber(p.paid)
	}));
}

export function portfolioPoints(snapshots: PortfolioSnapshotPoint[]): WorthPoint[] {
	return snapshots.map((s) => ({
		date: s.occurred_at.slice(0, 10),
		value: scaledToNumber(s.market_value),
		paid: scaledToNumber(s.total_cost_basis)
	}));
}

/**
 * One point per day from `from` to `to`. Worth only changes on the days something happened, so
 * each day carries the last known point forward; a day before the first point is worth nothing.
 */
export function dailyPoints(points: WorthPoint[], from: string, to: string): WorthPoint[] {
	const sorted = [...points].sort((a, b) => a.date.localeCompare(b.date));
	const out: WorthPoint[] = [];
	let last: WorthPoint | null = null;
	let next = 0;
	for (let day = from; day <= to; day = addDaysISO(day, 1)) {
		while (next < sorted.length && sorted[next].date <= day) last = sorted[next++];
		out.push({ date: day, value: last?.value ?? 0, paid: last?.paid ?? 0 });
	}
	return out;
}

/** Adds day-aligned series from `dailyPoints` into one. */
export function sumPoints(series: WorthPoint[][]): WorthPoint[] {
	if (series.length === 0) return [];
	return series[0].map((point, i) => ({
		date: point.date,
		value: series.reduce((total, s) => total + s[i].value, 0),
		paid: series.reduce((total, s) => total + s[i].paid, 0)
	}));
}

/** Drops the empty days before anything was held, so a chart starts where the item does. */
export function trimLeading(points: WorthPoint[]): WorthPoint[] {
	const first = points.findIndex((p) => p.value !== 0 || p.paid !== 0);
	return first === -1 ? [] : points.slice(first);
}

export function figuresFor(points: WorthPoint[]): WorthFigures | null {
	if (points.length === 0) return null;
	const start = points[0];
	const end = points[points.length - 1];
	const unrealized = end.value - end.paid;
	const added = end.paid - start.paid;
	const growth = end.value - start.value - added;
	// What was at work during the period: the worth it started with plus anything added since.
	const base = start.value + Math.max(0, added);
	return {
		worth: end.value,
		paid: end.paid,
		unrealized,
		unrealizedPct: end.paid > 0 ? (unrealized / end.paid) * 100 : null,
		growth,
		growthPct: base > 0 ? (growth / base) * 100 : null
	};
}
