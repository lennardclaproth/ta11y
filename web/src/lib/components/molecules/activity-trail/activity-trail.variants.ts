import type { TrailState } from './activity-trail.types';

/** Marker fill per state. Colour never stands alone — every step also carries its status word. */
export const trailMarkerClasses = {
	done: 'border-emerald-700 bg-emerald-700 text-white',
	warning: 'border-amber-500 bg-amber-500 text-slate-900',
	failed: 'border-red-700 bg-red-700 text-white',
	running: 'border-slate-600 bg-white text-slate-700',
	pending: 'border-slate-300 bg-white text-slate-400'
} satisfies Record<TrailState, string>;

export const trailIcons = {
	done: 'heroicons:check',
	warning: 'heroicons:exclamation-triangle',
	failed: 'heroicons:x-mark',
	running: 'heroicons:arrow-path',
	pending: 'heroicons:minus'
} satisfies Record<TrailState, string>;

/** Titles stay slate; only the detail line picks up the state tone. */
export const trailDetailClasses = {
	done: 'text-slate-600',
	// Warning stays slate: amber is legible as a marker fill, not as small text on white.
	warning: 'text-slate-700',
	failed: 'text-red-700',
	running: 'text-slate-700',
	pending: 'text-slate-400'
} satisfies Record<TrailState, string>;
