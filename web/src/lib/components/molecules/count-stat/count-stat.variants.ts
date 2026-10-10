import type { CountStatSize } from './count-stat.types';

/** The figure itself. Serif and tabular, never coloured by sign: this is a count, not money. */
export const countStatValueClasses = {
	sm: 'font-heading text-xl leading-tight text-slate-900 tabular-nums',
	md: 'font-heading text-2xl leading-tight text-slate-900 tabular-nums',
	lg: 'font-heading text-3xl leading-tight text-slate-900 tabular-nums'
} satisfies Record<CountStatSize, string>;

/** Height of the placeholder that stands in for the figure while it is unknown. */
export const countStatSkeletonClasses = {
	sm: 'h-6 w-12',
	md: 'h-7 w-14',
	lg: 'h-8 w-16'
} satisfies Record<CountStatSize, string>;
