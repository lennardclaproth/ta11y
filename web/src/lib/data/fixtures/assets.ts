import type {
	AssetClass,
	AssetClassDetails,
	AssetHolding,
	AssetSnapshotPoint
} from '$lib/api/types';

/** Asset classes for the account. Money is a decimal string; `growth_pct` is a percentage number. */
export const assetClasses: AssetClass[] = [
	{
		id: 'cls-cash',
		name: 'Cash & savings',
		source: 'manual',
		archived: false,
		current_worth: '18450.000000',
		last_change_at: '2026-06-12T09:00:00Z',
		growth_pct: 2.1,
		updated_at: '2026-06-12T09:00:00Z'
	},
	{
		id: 'cls-stocks',
		name: 'Brokerage',
		source: 'degiro',
		archived: false,
		current_worth: '12230.400000',
		last_change_at: '2026-06-16T22:00:00Z',
		growth_pct: 12.31,
		updated_at: '2026-06-16T22:00:00Z'
	},
	{
		id: 'cls-pension',
		name: 'Pension (BND)',
		source: 'brandnewday',
		archived: false,
		current_worth: '34120.500000',
		last_change_at: '2026-06-01T06:00:00Z',
		growth_pct: 5.4,
		updated_at: '2026-06-01T06:00:00Z'
	},
	{
		id: 'cls-crypto',
		name: 'Crypto',
		source: 'manual',
		archived: false,
		current_worth: '2980.250000',
		last_change_at: '2026-06-15T20:00:00Z',
		growth_pct: -8.7,
		updated_at: '2026-06-15T20:00:00Z'
	},
	{
		id: 'cls-realestate',
		name: 'Real estate equity',
		source: 'manual',
		archived: false,
		current_worth: '85000.000000',
		last_change_at: '2026-04-01T06:00:00Z',
		growth_pct: null,
		updated_at: '2026-04-01T06:00:00Z'
	},
	{
		id: 'cls-old',
		name: 'Legacy savings (archived)',
		source: 'manual',
		archived: true,
		current_worth: '0.000000',
		last_change_at: '2025-11-30T09:00:00Z',
		growth_pct: null,
		updated_at: '2025-11-30T09:00:00Z'
	}
];

/** Six-point growth timeline for a class (decimal-string total worth, oldest → newest). */
function growth(values: number[]): { date: string; total_worth: string }[] {
	const months = [
		'2026-01-01',
		'2026-02-01',
		'2026-03-01',
		'2026-04-01',
		'2026-05-01',
		'2026-06-01'
	];
	return values.map((v, i) => ({ date: months[i], total_worth: v.toFixed(6) }));
}

/**
 * Daily-priced items in the crypto class, keyed by item id. The quantities, prices
 * and purchases are invented round numbers.
 */
export const assetHoldings: Record<string, AssetHolding> = {
	'ast-btc': {
		id: 'ast-btc',
		class_id: 'cls-crypto',
		class_name: 'Crypto',
		name: 'Bitcoin',
		symbol: 'BTC/EUR',
		instrument_name: 'Bitcoin',
		quantity: '0.2',
		paid: '9600.000000',
		avg_unit_price: '48000.000000',
		price: '64000.000000',
		price_date: '2026-06-17',
		price_carried_forward: false,
		value: '12800.000000',
		unrealized: '3200.000000',
		unrealized_pct: 33.33,
		series: [
			{ date: '2026-01-15', value: '4000.000000', paid: '4000.000000' },
			{ date: '2026-02-01', value: '4500.000000', paid: '4000.000000' },
			{ date: '2026-03-02', value: '7800.000000', paid: '6600.000000' },
			{ date: '2026-04-01', value: '7200.000000', paid: '6600.000000' },
			{ date: '2026-05-20', value: '12000.000000', paid: '9600.000000' },
			{ date: '2026-06-01', value: '11600.000000', paid: '9600.000000' },
			{ date: '2026-06-17', value: '12800.000000', paid: '9600.000000' }
		],
		purchases: [
			{
				id: 'buy-btc-3',
				date: '2026-05-20',
				quantity: '0.05',
				unit_price: '60000.000000',
				paid: '3000.000000'
			},
			{
				id: 'buy-btc-2',
				date: '2026-03-02',
				quantity: '0.05',
				unit_price: '52000.000000',
				paid: '2600.000000'
			},
			{
				id: 'buy-btc-1',
				date: '2026-01-15',
				quantity: '0.1',
				unit_price: '40000.000000',
				paid: '4000.000000'
			}
		]
	},
	'ast-eth': {
		id: 'ast-eth',
		class_id: 'cls-crypto',
		class_name: 'Crypto',
		name: 'Ethereum',
		symbol: 'ETH/EUR',
		instrument_name: 'Ethereum',
		quantity: '3',
		paid: '7000.000000',
		avg_unit_price: '2333.330000',
		price: '2200.000000',
		// Two days behind, so the UI has a carried-forward price to explain.
		price_date: '2026-06-15',
		price_carried_forward: true,
		value: '6600.000000',
		unrealized: '-400.000000',
		unrealized_pct: -5.71,
		series: [
			{ date: '2026-02-10', value: '5000.000000', paid: '5000.000000' },
			{ date: '2026-03-01', value: '5400.000000', paid: '5000.000000' },
			{ date: '2026-04-18', value: '6000.000000', paid: '7000.000000' },
			{ date: '2026-05-01', value: '6900.000000', paid: '7000.000000' },
			{ date: '2026-06-01', value: '6300.000000', paid: '7000.000000' },
			{ date: '2026-06-15', value: '6600.000000', paid: '7000.000000' }
		],
		purchases: [
			{
				id: 'buy-eth-2',
				date: '2026-04-18',
				quantity: '1',
				unit_price: '2000.000000',
				paid: '2000.000000'
			},
			{
				id: 'buy-eth-1',
				date: '2026-02-10',
				quantity: '2',
				unit_price: '2500.000000',
				paid: '5000.000000'
			}
		]
	}
};

/** Class detail (drawer) fixtures keyed by class id. */
export const assetClassDetails: Record<string, AssetClassDetails> = {
	'cls-cash': {
		class: assetClasses[0],
		assets: [
			{
				id: 'ast-ing-savings',
				name: 'ING savings',
				current_worth: '12450.000000',
				archived: false,
				updated_at: '2026-06-12T09:00:00Z'
			},
			{
				id: 'ast-n26-space',
				name: 'N26 space',
				current_worth: '6000.000000',
				archived: false,
				updated_at: '2026-06-10T09:00:00Z'
			}
		],
		holdings: [],
		growth: growth([16800, 17100, 17500, 17900, 18200, 18450]),
		mutations: [
			{
				id: 'mut-c1',
				item_id: 'ast-ing-savings',
				change_type: 'deposit',
				direction: 'in',
				amount: '500.000000',
				previous_worth: '11950.000000',
				new_worth: '12450.000000',
				class_total_worth: '18450.000000',
				effective_date: '2026-06-12',
				note: 'Monthly transfer',
				created_at: '2026-06-12T09:00:00Z'
			},
			{
				id: 'mut-c2',
				item_id: 'ast-n26-space',
				change_type: 'deposit',
				direction: 'in',
				amount: '250.000000',
				previous_worth: '5750.000000',
				new_worth: '6000.000000',
				class_total_worth: '17950.000000',
				effective_date: '2026-05-12',
				note: null,
				created_at: '2026-05-12T09:00:00Z'
			}
		]
	},
	'cls-stocks': {
		class: assetClasses[1],
		assets: [
			{
				id: 'ast-vwrl',
				name: 'VWRL',
				current_worth: '4973.400000',
				archived: false,
				updated_at: '2026-06-16T22:00:00Z'
			},
			{
				id: 'ast-iwda',
				name: 'IWDA',
				current_worth: '3105.000000',
				archived: false,
				updated_at: '2026-06-16T22:00:00Z'
			},
			{
				id: 'ast-aapl',
				name: 'AAPL',
				current_worth: '2256.000000',
				archived: false,
				updated_at: '2026-06-16T22:00:00Z'
			},
			{
				id: 'ast-asml',
				name: 'ASML',
				current_worth: '1896.000000',
				archived: false,
				updated_at: '2026-06-16T22:00:00Z'
			}
		],
		holdings: [],
		growth: growth([9000, 9610, 9740, 10120, 10790, 12230.4]),
		mutations: [
			{
				id: 'mut-s1',
				item_id: 'ast-vwrl',
				change_type: 'valuation',
				direction: null,
				amount: '180.400000',
				previous_worth: '4793.000000',
				new_worth: '4973.400000',
				class_total_worth: '12230.400000',
				effective_date: '2026-06-16',
				note: 'EOD revaluation',
				created_at: '2026-06-16T22:00:00Z'
			}
		]
	},
	// A class that mixes both kinds of item: two that follow a daily price and one
	// whose worth is still set by hand, which is what the drawer and class page are
	// built to show side by side.
	'cls-crypto': {
		class: assetClasses[3],
		assets: [
			{
				id: 'ast-gold-coins',
				name: 'Gold coins',
				current_worth: '3000.000000',
				archived: false,
				updated_at: '2026-02-28T09:00:00Z'
			}
		],
		holdings: [assetHoldings['ast-btc'], assetHoldings['ast-eth']],
		growth: growth([9000, 14500, 17800, 18000, 21600, 22400]),
		mutations: [
			{
				id: 'mut-x1',
				item_id: 'ast-gold-coins',
				change_type: 'set',
				direction: null,
				amount: '3000.000000',
				previous_worth: '2800.000000',
				new_worth: '3000.000000',
				class_total_worth: '22400.000000',
				effective_date: '2026-02-28',
				note: 'Revalued after the last coin fair',
				created_at: '2026-02-28T09:00:00Z'
			}
		]
	}
};

/** Account-level total worth timeline (decimal-string), oldest → newest. */
export const assetSnapshots: AssetSnapshotPoint[] = [
	{ date: '2026-01-01', total_worth: '142300.000000' },
	{ date: '2026-02-01', total_worth: '144980.000000' },
	{ date: '2026-03-01', total_worth: '146120.000000' },
	{ date: '2026-04-01', total_worth: '149870.000000' },
	{ date: '2026-05-01', total_worth: '151540.000000' },
	{ date: '2026-06-01', total_worth: '152781.150000' }
];
