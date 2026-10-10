import type { MonthResult } from '$lib/api/types';
import type { ProgressBarIntent } from '$lib/components/atoms/progress-bar/progress-bar.types';
import type { BadgeIntent } from '$lib/components/atoms/badge/badge.types';

/**
 * The smallest track the bar uses: half of your income rather than 100%, so a goal of 30%
 * sits comfortably inside it and a month that put aside more than its goal still reads as
 * such. A larger goal grows the track instead — see `goalProgressTrack`.
 */
export const goalProgressScale = 50;

/**
 * The track for a goal, half as long again as the goal itself so the marker keeps room to
 * its right and an overshoot stays visible. Never past 100%: a share cannot exceed it.
 */
export function goalProgressTrack(goalPercent: number): number {
	return Math.min(100, Math.max(goalProgressScale, Math.ceil(goalPercent * 1.5)));
}

/**
 * A missed month needs attention, but it is not a validation error, so it is amber rather
 * than red. Shared by the bar and the badge so the two never drift apart.
 */
export const goalProgressIntents = {
	met: 'success',
	missed: 'warning',
	in_progress: 'primary',
	not_scored: 'secondary'
} satisfies Record<MonthResult, ProgressBarIntent>;

export const goalResultBadgeIntents = {
	met: 'success',
	missed: 'warning',
	in_progress: 'neutral',
	not_scored: 'neutral'
} satisfies Record<MonthResult, BadgeIntent>;
