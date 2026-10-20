import type { Vendor } from '$lib/api/types';

/** Active vendors available for imports / manual transactions (matches the parser vendors). */
export const vendors: Vendor[] = [
	{
		id: 'vnd-degiro',
		name: 'DEGIRO',
		type: 'brokerage',
		active: true,
		import_disabled: false,
		created_at: '2025-01-04T09:00:00Z',
		updated_at: '2026-01-10T09:00:00Z'
	},
	{
		id: 'vnd-ing',
		name: 'ING',
		type: 'bank',
		active: true,
		import_disabled: false,
		created_at: '2025-01-04T09:00:00Z',
		updated_at: '2026-01-10T09:00:00Z'
	},
	{
		id: 'vnd-n26',
		name: 'N26',
		type: 'bank',
		active: true,
		import_disabled: false,
		created_at: '2025-01-04T09:00:00Z',
		updated_at: '2026-01-10T09:00:00Z'
	},
	{
		id: 'vnd-bnd',
		name: 'Brand New Day',
		type: 'brokerage',
		active: true,
		import_disabled: false,
		created_at: '2025-01-04T09:00:00Z',
		updated_at: '2026-01-10T09:00:00Z'
	}
];
