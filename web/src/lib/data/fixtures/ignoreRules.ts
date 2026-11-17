import type { IgnoreRule } from '$lib/api/types';

/**
 * Deterministic ignore rules for fixture mode. Two kinds of rule the feature exists for —
 * an own transfer and a credit-card settlement — plus one that is switched off, so the
 * "off" state and the empty-harvest state are both reachable without a backend.
 */
export const ignoreRules: IgnoreRule[] = [
	{
		id: 'ir-0001',
		name: 'Transfer to savings',
		match_field: 'description',
		contains: 'Transfer to savings',
		direction: 'out',
		source: '',
		enabled: true,
		ignored_total: 48,
		last_applied_at: '2026-06-25T08:05:00Z',
		created_at: '2026-02-11T09:00:00Z',
		updated_at: '2026-06-25T08:05:00Z'
	},
	{
		id: 'ir-0002',
		name: 'Credit card payment',
		match_field: 'description',
		contains: 'Credit card',
		direction: 'out',
		source: 'ing',
		enabled: true,
		ignored_total: 19,
		last_applied_at: '2026-06-25T08:05:00Z',
		created_at: '2026-03-04T19:30:00Z',
		updated_at: '2026-06-25T08:05:00Z'
	},
	{
		id: 'ir-0003',
		name: 'Incoming own transfer',
		match_field: 'note',
		contains: 'Own account',
		direction: 'in',
		source: '',
		enabled: true,
		ignored_total: 27,
		last_applied_at: '2026-06-23T09:40:00Z',
		created_at: '2026-03-04T19:34:00Z',
		updated_at: '2026-06-23T09:40:00Z'
	},
	{
		id: 'ir-0004',
		name: 'Round-up savings',
		match_field: 'description',
		contains: 'Round-up',
		direction: 'out',
		source: 'n26',
		enabled: false,
		ignored_total: 12,
		last_applied_at: '2026-05-20T07:00:00Z',
		created_at: '2026-04-18T12:00:00Z',
		updated_at: '2026-05-21T10:12:00Z'
	}
];
