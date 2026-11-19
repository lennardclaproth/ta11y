/** Shared chart data shapes consumed by the chart organisms. */

/** One series in a time-series (line/bar/area) chart. */
export interface SeriesDataset {
	label: string;
	/** Y values, aligned 1:1 with the chart's `labels`. */
	data: number[];
	/** Render as a line (default) or bars. */
	type?: 'line' | 'bar';
	/** Series color (hex from charts/theme). */
	color: string;
	/** Fill the area under a line with a top→bottom gradient. */
	fill?: boolean;
	/** Dashed line (e.g. a projected/secondary series). */
	dashed?: boolean;
	/** Flip the line segment to the negative color when it dips below zero. */
	signed?: boolean;
	/** Shade the gap to the previous dataset: gain colour above it, loss colour below. */
	gapFill?: boolean;
	/** One colour per bar, aligned with `data`; overrides `color` for the bars. */
	barColors?: string[];
}

/** One slice of a donut chart. */
export interface DonutDatum {
	label: string;
	value: number;
	/** Optional explicit slice color; otherwise the chart assigns from its ramp. */
	color?: string;
}
