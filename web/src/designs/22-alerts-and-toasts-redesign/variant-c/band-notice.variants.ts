import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';

/**
 * Variant C — "Masthead band". One full-width ruled band for every notification, in normal document
 * flow: under the masthead after an action, above the records when the problem stays, under the
 * dialog header when a save fails. Nothing floats over the page.
 */
export const bandBaseClasses =
	'flex w-full items-center gap-3 border-t border-b border-t-slate-300 border-b-slate-300 bg-white px-4 py-2.5';

/** Solid chip: the only place the intent colour is used at full strength. */
export const bandChipClasses = {
	info: 'bg-sky-700 text-white',
	success: 'bg-emerald-700 text-white',
	warning: 'bg-amber-500 text-slate-900',
	error: 'bg-red-700 text-white'
} satisfies Record<AlertIntent, string>;

export const bandLabels = {
	info: 'Information',
	success: 'Success',
	warning: 'Warning',
	error: 'Error'
} satisfies Record<AlertIntent, string>;

export const bandIcons = {
	info: 'heroicons:information-circle',
	success: 'heroicons:check-circle',
	warning: 'heroicons:exclamation-triangle',
	error: 'heroicons:exclamation-circle'
} satisfies Record<AlertIntent, string>;
