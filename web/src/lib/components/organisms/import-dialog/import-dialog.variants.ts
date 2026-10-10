import type { BadgeIntent } from '$lib/components/atoms/badge/badge.types';
import type { DestinationStatus } from './import-dialog.types';

/** Badge intent per destination status; the Badge always carries its status word too. */
export const destinationBadgeIntents = {
	completed: 'success',
	partial: 'warning',
	failed: 'error'
} satisfies Record<DestinationStatus, BadgeIntent>;

/**
 * One colour class per count cell: two competing `text-*` utilities would be resolved
 * by stylesheet order rather than by intent. A zero is always muted — emphasis belongs
 * on counts that actually happened.
 */
export const countToneClasses = {
	zero: 'text-slate-500',
	failed: 'font-medium text-red-700',
	imported: 'font-semibold text-slate-950',
	plain: 'text-slate-800'
} as const;
