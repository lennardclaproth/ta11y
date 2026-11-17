import type {
	AssetsBackupResult,
	BackupDataset,
	BackupResult,
	CashflowBackupResult,
	EODBackupResult,
	PortfolioBackupResult
} from '$lib/api/types';

const none = { vendors: [], listings: [] };

/**
 * What importing each dataset's file reports in fixture mode: most records new, a few
 * already present, nothing missing — the shape a restore into a fresh account takes.
 */
export const backupPreviews: {
	cashflow: CashflowBackupResult;
	portfolio: PortfolioBackupResult;
	assets: AssetsBackupResult;
	eods: EODBackupResult;
} = {
	cashflow: {
		dry_run: true,
		imports: { new: 14, existing: 2 },
		transactions: { new: 1284, existing: 36 },
		ignore_rules: { new: 4, existing: 1 },
		recurring_items: { new: 9, existing: 1 },
		recurring_links: { new: 118, existing: 12 },
		recurring_dismissals: { new: 3, existing: 0 },
		missing: none
	},
	portfolio: {
		dry_run: true,
		imports: { new: 6, existing: 0 },
		transactions: { new: 212, existing: 0 },
		missing: none
	},
	assets: {
		dry_run: true,
		classes: { new: 3, existing: 0 },
		items: { new: 7, existing: 0 },
		mutations: { new: 41, existing: 0 },
		purchases: { new: 6, existing: 0 },
		missing: none
	},
	eods: {
		dry_run: true,
		listings: { new: 14, existing: 3 },
		eods: { new: 18240, existing: 2110 },
		missing: none
	}
} satisfies Record<BackupDataset, BackupResult>;

/**
 * A file fixture mode downloads for an export. It carries the header an import checks, so
 * exporting and importing again round-trips without a backend.
 */
export function backupFixture(dataset: BackupDataset): Record<string, unknown> {
	return {
		format: 'ta11y.backup',
		version: 1,
		kind: dataset,
		exported_at: new Date().toISOString()
	};
}
