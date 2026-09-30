/**
 * Prototype-only fake data for the DeGiro import design (#10).
 * Round numbers and invented products — never real holdings, ISINs or amounts.
 */

export type ImportDestinationName = 'Cashflow' | 'Portfolio';

export type ImportDestinationStatus = 'completed' | 'partial' | 'failed';

export interface ImportDestinationResult {
	destination: ImportDestinationName;
	status: ImportDestinationStatus;
	/** Rows the parser read for this destination. */
	total: number;
	imported: number;
	duplicates: number;
	failed: number;
	/** Shown when the destination did not fully succeed. */
	message?: string;
}

export interface UnlinkedProduct {
	id: string;
	name: string;
	isin: string | null;
	symbol: string | null;
	transactions: number;
}

export const importFileName = 'degiro-account-2026-09.csv';
export const importFileSize = '48.0 KB';
export const importFinishedAt = '30 Sep 2026, 14:20';

export const destinationResults: ImportDestinationResult[] = [
	{ destination: 'Cashflow', status: 'completed', total: 240, imported: 180, duplicates: 60, failed: 0 },
	{
		destination: 'Portfolio',
		status: 'partial',
		total: 120,
		imported: 80,
		duplicates: 36,
		failed: 4,
		message: '4 rows had no product name and were skipped.'
	}
];

/** The likely "empty" outcome of a monthly export: every row overlapped with an earlier upload. */
export const noNewRowsResults: ImportDestinationResult[] = [
	{ destination: 'Cashflow', status: 'completed', total: 240, imported: 0, duplicates: 240, failed: 0 },
	{ destination: 'Portfolio', status: 'completed', total: 120, imported: 0, duplicates: 120, failed: 0 }
];

export const unlinkedProducts: UnlinkedProduct[] = [
	{ id: 'p-1', name: 'Example World Index Fund', isin: 'XX0000000001', symbol: null, transactions: 12 },
	{ id: 'p-2', name: 'Sample Technology Corp', isin: 'XX0000000002', symbol: 'SMPL', transactions: 6 },
	{ id: 'p-3', name: 'Demo Energy ETF', isin: null, symbol: 'DEMO', transactions: 3 }
];

/** Accounts the export can belong to. DEGIRO is the only vendor this bet covers. */
export const accountOptions = [
	{ value: 'vnd-degiro', label: 'DEGIRO — brokerage' },
	{ value: 'vnd-bnd', label: 'Brand New Day — brokerage' }
];

export const wrongFileMessage =
	'This is not the DEGIRO Account statement export. Export "Account statement" from DEGIRO and upload that CSV.';

export const countRows = [
	{ key: 'total', label: 'Rows read' },
	{ key: 'imported', label: 'New' },
	{ key: 'duplicates', label: 'Already imported' },
	{ key: 'failed', label: 'Failed' }
] as const;
