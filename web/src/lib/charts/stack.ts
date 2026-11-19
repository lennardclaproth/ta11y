// Stacking series that each change on their own days: one shared, sorted set of dates, and every
// series carried forward onto it. Pure, so the stacked chart's numbers do not depend on a page.

/**
 * Every date any series moved on, plus the window's own ends, limited to `[from, end]` and
 * sorted. Including the ends keeps the chart spanning the whole window even where nothing changed.
 */
export function unionDates(
	dates: readonly (readonly string[])[],
	from: string,
	end: string
): string[] {
	return [...new Set([from, ...dates.flat(), end])]
		.filter((date) => !!date && date >= from && date <= end)
		.sort();
}

/**
 * The series' value on each of `dates`: its last known value on or before that date, and 0 before
 * its first. `points` and `dates` are both oldest first.
 */
export function forwardFill(
	points: readonly { date: string; value: number }[],
	dates: readonly string[]
): number[] {
	let last = 0;
	let i = 0;
	return dates.map((date) => {
		while (i < points.length && points[i].date <= date) {
			last = points[i].value;
			i += 1;
		}
		return last;
	});
}
