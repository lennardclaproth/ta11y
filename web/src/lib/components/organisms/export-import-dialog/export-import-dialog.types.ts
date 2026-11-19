import type { BackupCount, BackupDataset, BackupResult } from '$lib/api/types';

/** One dataset that can be exported to its own file and imported from one. */
export interface DatasetOption {
	value: BackupDataset;
	title: string;
	/** What the file holds, in the reader's words. */
	holds: string;
	/** Only administrators move it: it is shared by every account. */
	adminOnly?: boolean;
	/** What ta11y rebuilds after an import, when it rebuilds anything. */
	rebuilds?: string;
}

export const datasetOptions: DatasetOption[] = [
	{
		value: 'cashflow',
		title: 'Cashflow',
		holds: 'Transactions with their tags and ignored state, ignore rules and recurring items.'
	},
	{
		value: 'portfolio',
		title: 'Portfolio',
		holds: 'Every trade, dividend, fee and cash movement.',
		rebuilds: 'Positions and snapshots are rebuilt afterwards.'
	},
	{
		value: 'assets',
		title: 'Asset classes',
		holds: 'The classes you keep by hand, with their items, worth history and purchases.',
		rebuilds: 'Asset snapshots are rebuilt afterwards.'
	},
	{
		value: 'eods',
		title: 'End-of-day prices',
		holds:
			'Every listing with its instrument details, and its daily prices with their splits. Shared by every account — on a new instance, import it before the other datasets.',
		adminOnly: true
	}
];

/**
 * Where an import stands. `refused` is a file the server would not read (not a ta11y
 * export, or another dataset's); `missing` is one naming what this ta11y lacks.
 */
export const transferPhases = [
	'idle',
	'checking',
	'preview',
	'missing',
	'refused',
	'importing',
	'done'
] as const;

export type TransferPhase = (typeof transferPhases)[number];

/** One ruled line of a dry run: a kind of record and how many of it are new. */
export interface CountLine {
	key: string;
	label: string;
	count: BackupCount;
}

/** The lines a dataset's import reports, in the order the dialog reads them. */
export function countLines(dataset: BackupDataset, result: BackupResult): CountLine[] {
	const r = result as unknown as Record<string, BackupCount>;
	const labels: Record<BackupDataset, [string, string][]> = {
		cashflow: [
			['transactions', 'Transactions'],
			['imports', 'Imports'],
			['ignore_rules', 'Ignore rules'],
			['recurring_items', 'Recurring items'],
			['recurring_links', 'Linked to a recurring item'],
			['recurring_dismissals', 'Dismissed suggestions']
		],
		portfolio: [
			['transactions', 'Transactions'],
			['imports', 'Imports']
		],
		assets: [
			['classes', 'Classes'],
			['items', 'Items'],
			['mutations', 'Worth changes'],
			['purchases', 'Purchases']
		],
		eods: [
			['listings', 'Listings'],
			['eods', 'Days of prices']
		]
	};
	return labels[dataset]
		.filter(([key]) => r[key])
		.map(([key, label]) => ({ key, label, count: r[key] }));
}
