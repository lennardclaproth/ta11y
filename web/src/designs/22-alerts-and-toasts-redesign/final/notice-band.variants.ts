import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';

/**
 * Final design for #22 — "Masthead band" (variant C, chosen in round 1).
 *
 * One shape for every notification: a full-width band with a hairline rule above and below, always
 * in normal document flow. Under the masthead after an action, above the records when a problem
 * stays, under the dialog header when a save fails. Nothing floats over the page, so a notice can
 * never cover the navigation, the ledger actions or a table row.
 */
export const noticeBandBaseClasses =
	'flex w-full items-start gap-3 border-t border-b border-slate-300 py-2.5';

/**
 * Horizontal gutter. The band runs edge to edge, but its text lines up with the text of whatever
 * contains it: the page gutter under the masthead, the panel gutter above the records, the dialog
 * gutter under a modal header. Separate literals rather than a `class` override, because two
 * padding utilities in one class string are resolved by stylesheet order, not by position.
 */
export const noticeBandGutterClasses = {
	page: 'px-4 lg:px-6',
	panel: 'px-4',
	dialog: 'px-5'
} satisfies Record<NoticeBandGutter, string>;

export const noticeBandGutters = ['page', 'panel', 'dialog'] as const;
export type NoticeBandGutter = (typeof noticeBandGutters)[number];

/**
 * Surface. `paper` is a fresh white sheet on the taupe canvas and inside the taupe-50 panel;
 * `inset` is the warm sheet used inside the white dialog, where white on white would vanish.
 */
export const noticeBandSurfaceClasses = {
	paper: 'bg-white',
	inset: 'bg-taupe-50'
} satisfies Record<NoticeBandSurface, string>;

export const noticeBandSurfaces = ['paper', 'inset'] as const;
export type NoticeBandSurface = (typeof noticeBandSurfaces)[number];

/** Solid chip: the only place the intent colour is used at full strength. */
export const noticeBandChipClasses = {
	info: 'bg-sky-700 text-white',
	success: 'bg-emerald-700 text-white',
	warning: 'bg-amber-500 text-slate-900',
	error: 'bg-red-700 text-white'
} satisfies Record<AlertIntent, string>;

/**
 * Screen-reader-only intent name, so colour is never the only carrier of meaning. Not visible copy:
 * the pitch's no-go on new or rewritten notification texts stays intact.
 */
export const noticeBandLabels = {
	info: 'Information',
	success: 'Success',
	warning: 'Warning',
	error: 'Error'
} satisfies Record<AlertIntent, string>;

export const noticeBandIcons = {
	info: 'heroicons:information-circle',
	success: 'heroicons:check-circle',
	warning: 'heroicons:exclamation-triangle',
	error: 'heroicons:exclamation-circle'
} satisfies Record<AlertIntent, string>;
