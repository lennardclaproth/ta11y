/**
 * Fixture-mode stand-in for the import pipeline. It reads the uploaded CSV the way
 * `internal/importer/{cashflow,portfolio}` does — headers decide whether the file is
 * the export at all, content decides which rows are new — so running the portal on
 * mocks exercises the behaviour this feature is about instead of returning a
 * hard-coded success. Nothing leaves the browser and nothing is persisted.
 */

import type { ImportResult, ImportType, UnlinkedProduct } from '$lib/api/types';
import { listings } from '$lib/data/fixtures/marketdata';

/** The headers both DEGIRO parsers require before they will read a single row. */
const requiredHeaders = ['Date', 'Value date', 'Product', 'ISIN', 'Description', 'Change'];

const notRecognised =
	'This is not the DEGIRO Account statement export. Export "Account statement" from DEGIRO and upload that CSV.';

/**
 * Rows already imported in this session, per destination. The real deduplication lives
 * in the database; a module-level set is the fixture equivalent, so re-uploading a
 * file in mock mode reports it as already imported rather than doubling the counts.
 */
const importedRows: Record<ImportType, Set<string>> = {
	cashflow: new Set(),
	portfolio: new Set(),
	eod: new Set()
};

const results = new Map<string, ImportResult>();

/** Number of `getImport` calls a mock import spends in `processing` before finishing. */
const processingPolls = 1;
const pollsLeft = new Map<string, number>();

function splitCsvLine(line: string): string[] {
	return line.split(',').map((cell) => cell.trim().replace(/^"|"$/g, ''));
}

function emptyResult(importId: string, type: ImportType): ImportResult {
	const now = new Date().toISOString();
	return {
		import_id: importId,
		type,
		status: 'pending',
		status_msg: '',
		reason: '',
		total_rows: 0,
		imported: 0,
		duplicates: 0,
		failed: 0,
		created_at: now,
		updated_at: now,
		unlinked_products: []
	};
}

/** Exact identity, matching `marketdata.Queries.ListingByIdentity`: ISIN, else symbol. */
function hasListing(isin: string, symbol: string): boolean {
	if (isin) return listings.some((listing) => listing.isin === isin);
	if (symbol) return listings.some((listing) => listing.symbol === symbol);
	return false;
}

function collectUnlinkedProducts(rows: Record<string, string>[]): UnlinkedProduct[] {
	const products = new Map<string, UnlinkedProduct>();
	for (const row of rows) {
		const isin = row['ISIN'] ?? '';
		const name = row['Product'] ?? '';
		if (!isin && !name) continue;
		if (hasListing(isin, '')) continue;
		const key = isin || name;
		const existing = products.get(key);
		if (existing) {
			existing.transactions += 1;
			continue;
		}
		products.set(key, { name, isin: isin || null, symbol: null, transactions: 1 });
	}
	return [...products.values()];
}

/**
 * Reads the upload and stores the outcome the next `getImport` will report. The counts
 * are the file's own; only the lifecycle is simulated.
 */
export async function registerMockImport(
	importId: string,
	type: ImportType,
	file: File
): Promise<void> {
	const result = emptyResult(importId, type);
	results.set(importId, result);
	pollsLeft.set(importId, processingPolls);

	const lines = (await file.text())
		.split('\n')
		.map((line) => line.trim())
		.filter((line) => line !== '');

	const header = lines.length > 0 ? splitCsvLine(lines[0]) : [];
	const missing = requiredHeaders.filter((column) => !header.includes(column));
	if (missing.length > 0) {
		result.status = 'failed';
		result.reason = 'file_not_recognised';
		result.status_msg = notRecognised;
		return;
	}

	const seen = importedRows[type];
	const rows: Record<string, string>[] = [];
	// Two identical rows inside one file are two transactions, so the occurrence count
	// is part of the key — exactly what DedupSequencer does on the server.
	const occurrences = new Map<string, number>();

	for (const line of lines.slice(1)) {
		const cells = splitCsvLine(line);
		const row: Record<string, string> = {};
		header.forEach((column, index) => {
			row[column] = cells[index] ?? '';
		});
		rows.push(row);

		const content = header.map((column) => row[column]).join('');
		const occurrence = (occurrences.get(content) ?? 0) + 1;
		occurrences.set(content, occurrence);
		const key = `${content}${occurrence}`;

		result.total_rows += 1;
		if (seen.has(key)) {
			result.duplicates += 1;
			continue;
		}
		seen.add(key);
		result.imported += 1;
	}

	result.status = 'completed';
	if (type === 'portfolio') {
		result.unlinked_products = collectUnlinkedProducts(rows);
	}
}

/**
 * Returns the stored outcome, reporting `processing` on the first poll so the caller's
 * activity trail has something to show before the result lands.
 */
export function resolveMockImport(importId: string): ImportResult | null {
	const result = results.get(importId);
	if (!result) return null;

	const remaining = pollsLeft.get(importId) ?? 0;
	if (remaining > 0) {
		pollsLeft.set(importId, remaining - 1);
		return { ...result, status: 'processing', unlinked_products: [] };
	}
	return result;
}
