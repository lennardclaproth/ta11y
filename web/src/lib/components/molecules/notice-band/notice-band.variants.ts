import type {
	NoticeBandGutter,
	NoticeBandSurface,
	NoticeIntent
} from './notice-band.types';

/**
 * "Masthead band": one shape for every notification — a full-width band with a hairline rule above
 * and below, always in normal document flow. Under the masthead after an action, above the records
 * when a problem stays, under the dialog header when a save fails. Nothing floats over the page, so
 * a notice can never cover the navigation, the ledger actions or a table row.
 *
 * No `w-full`: callers bleed the band to their container's edges with negative `-mx-*`, and only an
 * auto width grows with those margins — a fixed 100% would stop short of the right edge.
 */
export const noticeBandBaseClasses =
	'flex items-start gap-3 border-t border-b border-slate-300 py-2.5';

/**
 * Separate literals rather than a `class` override, because two padding utilities in one class
 * string are resolved by stylesheet order, not by position.
 */
export const noticeBandGutterClasses = {
	page: 'px-4 lg:px-8',
	panel: 'px-4',
	dialog: 'px-5',
	none: 'px-0'
} satisfies Record<NoticeBandGutter, string>;

export const noticeBandSurfaceClasses = {
	paper: 'bg-white',
	inset: 'bg-taupe-50'
} satisfies Record<NoticeBandSurface, string>;

/** Solid chip: the only place the intent colour is used at full strength. */
export const noticeBandChipClasses = {
	info: 'bg-sky-700 text-white',
	success: 'bg-emerald-700 text-white',
	warning: 'bg-amber-500 text-slate-900',
	error: 'bg-red-700 text-white'
} satisfies Record<NoticeIntent, string>;

/**
 * Screen-reader-only intent name, so colour is never the only carrier of meaning. Not visible copy,
 * so it adds no notification text to the app.
 */
export const noticeBandLabels = {
	info: 'Information',
	success: 'Success',
	warning: 'Warning',
	error: 'Error'
} satisfies Record<NoticeIntent, string>;

/**
 * Solid glyphs: inside the solid chip an outline icon's thin ring swallows the "i" and "!", so the
 * mark is cut out of a filled shape instead.
 */
export const noticeBandIcons = {
	info: 'heroicons:information-circle-20-solid',
	success: 'heroicons:check-circle-20-solid',
	warning: 'heroicons:exclamation-triangle-20-solid',
	error: 'heroicons:exclamation-circle-20-solid'
} satisfies Record<NoticeIntent, string>;
