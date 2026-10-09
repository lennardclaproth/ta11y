import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';

/**
 * Variant A — "Ruled notice". No filled surface: the intent is carried by a 2px rule and a small
 * uppercase kicker, the way the portal already rules its KPI sections and ledger headers.
 */
export const ruledSurfaceClasses = {
	/** In a content panel, a form or a dialog: rules only, no box. */
	inline: 'border-t-2 border-b border-b-slate-200 py-3',
	/** Floating above content: the same notice on a square sheet of paper. */
	paper:
		'border-t-2 border-r border-b border-l border-r-slate-200 border-b-slate-200 border-l-slate-200 bg-white px-4 py-3 shadow-md'
} satisfies Record<'inline' | 'paper', string>;

export const ruledRuleClasses = {
	info: 'border-t-sky-700',
	success: 'border-t-emerald-700',
	warning: 'border-t-amber-500',
	error: 'border-t-red-700'
} satisfies Record<AlertIntent, string>;

export const ruledKickerClasses = {
	info: 'text-sky-700',
	success: 'text-emerald-700',
	warning: 'text-amber-600',
	error: 'text-red-700'
} satisfies Record<AlertIntent, string>;

export const ruledKickerLabels = {
	info: 'Info',
	success: 'Success',
	warning: 'Warning',
	error: 'Error'
} satisfies Record<AlertIntent, string>;

export const ruledIcons = {
	info: 'heroicons:information-circle',
	success: 'heroicons:check-circle',
	warning: 'heroicons:exclamation-triangle',
	error: 'heroicons:exclamation-circle'
} satisfies Record<AlertIntent, string>;
