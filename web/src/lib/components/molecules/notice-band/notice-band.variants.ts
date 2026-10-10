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
 */
export const noticeBandBaseClasses =
	'flex w-full items-start gap-3 border-t border-b border-slate-300 py-2.5';

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

export const noticeBandIcons = {
	info: 'heroicons:information-circle',
	success: 'heroicons:check-circle',
	warning: 'heroicons:exclamation-triangle',
	error: 'heroicons:exclamation-circle'
} satisfies Record<NoticeIntent, string>;
