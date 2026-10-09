import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';

/**
 * Variant B — "Ledger entry". A slip with a stub column (icon + intent word) separated from the
 * message by a hairline, the way the records table separates its columns.
 */
export const ledgerSurfaceClasses = {
	inline: 'bg-white',
	paper: 'bg-white shadow-md'
} satisfies Record<'inline' | 'paper', string>;

export const ledgerBaseClasses = [
	'flex items-stretch border-t border-r border-b',
	'border-t-slate-200 border-r-slate-200 border-b-slate-200',
	'border-l-[3px]'
].join(' ');

export const ledgerStubClasses = {
	info: 'border-l-sky-700 text-sky-700',
	success: 'border-l-emerald-700 text-emerald-700',
	warning: 'border-l-amber-500 text-amber-600',
	error: 'border-l-red-700 text-red-700'
} satisfies Record<AlertIntent, string>;

export const ledgerStubLabels = {
	info: 'Info',
	success: 'Success',
	warning: 'Warning',
	error: 'Error'
} satisfies Record<AlertIntent, string>;

export const ledgerIcons = {
	info: 'heroicons:information-circle',
	success: 'heroicons:check-circle',
	warning: 'heroicons:exclamation-triangle',
	error: 'heroicons:exclamation-circle'
} satisfies Record<AlertIntent, string>;
