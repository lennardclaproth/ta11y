import type { CashflowDirection, RecurringRhythm } from '$lib/api/types';

/** The two routes out of the dialog, in the order it offers them. */
export const markRecurringModes = ['existing', 'new'] as const;
export type MarkRecurringMode = (typeof markRecurringModes)[number];

/** What the dialog hands back: either an item that exists, or one to start. */
export type MarkRecurringValue =
	| { mode: 'existing'; itemId: string }
	| { mode: 'new'; name: string; direction: CashflowDirection; rhythm: RecurringRhythm };
