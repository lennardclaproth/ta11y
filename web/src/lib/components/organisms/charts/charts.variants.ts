/**
 * Chart frame heights for the overview cards of the scrolling pages. A chart fills what is left
 * of the screen under the masthead, the page heading and its card title, so the cards read as
 * the page and the table follows when you scroll. Whole literals, as the Tailwind 4 scanner needs.
 */
export const chartHeights = {
	/** A chart in an overview card that has the screen to itself. */
	tall: 'h-56 lg:h-[max(16rem,calc(100dvh-19rem))]',
	/** Under a one-line legend. */
	tallUnderLegend: 'h-52 lg:h-[max(15rem,calc(100dvh-21rem))]',
	/** Between a one-line legend and a one-line caption. */
	tallBetweenLegendAndCaption: 'h-48 lg:h-[max(14rem,calc(100dvh-23rem))]',
	/** The same, for a chart that shares the screen with a KPI row above it. */
	tallUnderKpis: 'h-56 lg:h-[max(16rem,calc(100dvh-27rem))]'
} as const;
