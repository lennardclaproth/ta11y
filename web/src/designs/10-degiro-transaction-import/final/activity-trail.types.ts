/**
 * Proposed vocabulary for the import activity trail (#10).
 * Would move to `molecules/activity-trail/activity-trail.types.ts` when promoted.
 */

export const trailDirections = ['horizontal', 'vertical'] as const;
export type TrailDirection = (typeof trailDirections)[number];

export const trailStates = ['done', 'warning', 'failed', 'running', 'pending'] as const;
export type TrailState = (typeof trailStates)[number];

export interface TrailStep {
	key: string;
	title: string;
	/** Short line under the title: a count, a status word, or what it is waiting for. */
	detail?: string;
	state: TrailState;
}
