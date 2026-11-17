import type { MonthResult } from '$lib/api/types';

/**
 * Every utility is a whole literal so the Tailwind 4 scanner can see it. A missed month is
 * amber, not red: it needs attention, but a missed goal is not a validation error.
 */
export const streakMarkClasses = {
	met: 'border-emerald-700 bg-emerald-100 text-emerald-800',
	missed: 'border-amber-500 bg-amber-100 text-amber-800',
	in_progress: 'border-slate-300 bg-slate-100 text-slate-500',
	not_scored: 'border-slate-200 bg-slate-50 text-slate-400'
} satisfies Record<MonthResult, string>;

export const streakMarkIcons = {
	met: 'heroicons:check',
	missed: 'heroicons:x-mark',
	in_progress: 'heroicons:ellipsis-horizontal',
	not_scored: 'heroicons:minus-small'
} satisfies Record<MonthResult, string>;
