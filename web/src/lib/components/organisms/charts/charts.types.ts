/** `md`: the ring beside its legend. `lg`: a larger ring above it, for a tall card. */
export const donutChartSizes = ['md', 'lg'] as const;
export type DonutChartSize = (typeof donutChartSizes)[number];

/** One row of a ranked bar chart. */
export interface RankedBarDatum {
	/** Stable identity, handed back on select (a tag, a class id). */
	key: string;
	label: string;
	value: number;
	/** Explicit bar colour; otherwise the chart's ramp by rank. */
	color?: string;
}
