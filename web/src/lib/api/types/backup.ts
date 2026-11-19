/**
 * Export and import of one dataset at a time. Every export is a JSON file that an import
 * takes back; an import merges — what the account has is left alone — and a dry run
 * reports the same counts without writing anything.
 */

/** What a file holds; it is also the path segment of its endpoints. */
export type BackupDataset = 'cashflow' | 'portfolio' | 'assets' | 'eods';

/** How many records of one kind were new to the account, and how many it already had. */
export interface BackupCount {
	new: number;
	existing: number;
}

/** A listing named the way two ta11y instances agree on it. */
export interface BackupListingRef {
	symbol: string;
	source: string;
}

/** What a file names that this ta11y lacks; such a file is refused before anything is written. */
export interface BackupMissing {
	vendors: string[];
	listings: BackupListingRef[];
}

/** `POST /backup/cashflow/import` — mirrors `backup.CashflowImportResponse`. */
export interface CashflowBackupResult {
	dry_run: boolean;
	imports: BackupCount;
	transactions: BackupCount;
	ignore_rules: BackupCount;
	recurring_items: BackupCount;
	recurring_links: BackupCount;
	recurring_dismissals: BackupCount;
	missing: BackupMissing;
}

/** `POST /backup/portfolio/import` — mirrors `backup.PortfolioImportResponse`. */
export interface PortfolioBackupResult {
	dry_run: boolean;
	imports: BackupCount;
	transactions: BackupCount;
	missing: BackupMissing;
}

/** `POST /backup/assets/import` — mirrors `backup.AssetsImportResponse`. */
export interface AssetsBackupResult {
	dry_run: boolean;
	classes: BackupCount;
	items: BackupCount;
	mutations: BackupCount;
	purchases: BackupCount;
	missing: BackupMissing;
}

/** `POST /backup/eods/import` — mirrors `backup.EODImportResponse`. */
export interface EODBackupResult {
	dry_run: boolean;
	listings: BackupCount;
	eods: BackupCount;
	missing: BackupMissing;
}

/** What an import of one dataset did, or on a dry run would do. */
export type BackupResult =
	| CashflowBackupResult
	| PortfolioBackupResult
	| AssetsBackupResult
	| EODBackupResult;
