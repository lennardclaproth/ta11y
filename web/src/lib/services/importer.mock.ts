/**
 * Fixture-mode stand-in for the import pipeline. It reads the uploaded CSV the way
 * `internal/importer/{cashflow,portfolio}` does — the vendor's headers decide whether the
 * file is its export at all, content decides which rows are new — so running the portal on
 * mocks exercises the behaviour this feature is about instead of returning a
 * hard-coded success. Nothing leaves the browser and nothing is persisted.
 */

import type { ImportResult, ImportType, UnlinkedProduct } from '$lib/api/types';
import { listings } from '$lib/data/fixtures/marketdata';
import { vendors } from '$lib/data/fixtures/vendors';

/** How one vendor's export is recognised and when one of its rows counts as failed. */
interface ExportFormat {
	delimiter: string;
	/** The headers the vendor's parser requires before it will read a single row. */
	required: string[];
	/** True for a row the parser would refuse; such a row is counted, not imported. */
	rowFails?: (row: Record<string, string>) => boolean;
}

const isAmount = (raw: string | undefined) => /^-?\d+([.,]\d+)?$/.test((raw ?? '').trim());

/**
 * The exports per vendor name, after `internal/importer/cashflow/parsers`. N26 is matched
 * on its English headers only; the server also accepts the German ones. DEGIRO rows are
 * never counted as failed, matching its parsers, which skip a row they cannot read.
 */
const exportFormats: Record<string, ExportFormat> = {
	DEGIRO: {
		delimiter: ',',
		required: ['Date', 'Value date', 'Product', 'ISIN', 'Description', 'Change']
	},
	ING: {
		delimiter: ';',
		required: ['Date', 'Name / Description', 'Debit/credit', 'Amount (EUR)'],
		rowFails: (row) =>
			!/^\d{8}$/.test(row['Date'] ?? '') ||
			!['debit', 'credit'].includes((row['Debit/credit'] ?? '').toLowerCase()) ||
			!isAmount(row['Amount (EUR)'])
	},
	N26: {
		delimiter: ',',
		required: ['Booking Date', 'Amount (EUR)'],
		rowFails: (row) => !row['Booking Date'] || !isAmount(row['Amount (EUR)'])
	}
};

function formatFor(vendorId: string): ExportFormat {
	const name = vendors.find((vendor) => vendor.id === vendorId)?.name ?? 'DEGIRO';
	return exportFormats[name] ?? exportFormats.DEGIRO;
}

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

/** Splits one CSV line, keeping a delimiter inside a quoted cell ("Smith, J") in its cell. */
function splitCsvLine(line: string, delimiter: string): string[] {
	const cells: string[] = [];
	let cell = '';
	let quoted = false;
	for (const char of line) {
		if (char === '"') quoted = !quoted;
		else if (char === delimiter && !quoted) {
			cells.push(cell.trim());
			cell = '';
		} else cell += char;
	}
	cells.push(cell.trim());
	return cells;
}

function emptyResult(importId: string, type: ImportType): ImportResult {
	const now = new Date().toISOString();
	return {
		import_id: importId,
		type,
		status: 'pending',
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
	file: File,
	vendorId: string
): Promise<void> {
	const result = emptyResult(importId, type);
	results.set(importId, result);
	pollsLeft.set(importId, processingPolls);

	const format = formatFor(vendorId);
	const lines = (await file.text())
		.replace(/^\uFEFF/, '')
		.split('\n')
		.map((line) => line.trim())
		.filter((line) => line !== '');

	// N26 exports with either delimiter, so the header line decides.
	const delimiter =
		format === exportFormats.N26 && lines[0]?.includes(';') ? ';' : format.delimiter;
	const header = lines.length > 0 ? splitCsvLine(lines[0], delimiter) : [];
	const missing = format.required.filter((column) => !header.includes(column));
	if (missing.length > 0) {
		result.status = 'failed';
		result.reason = 'file_not_recognised';
		return;
	}

	const seen = importedRows[type];
	const rows: Record<string, string>[] = [];
	// Two identical rows inside one file are two transactions, so the occurrence count
	// is part of the key — exactly what DedupSequencer does on the server.
	const occurrences = new Map<string, number>();

	for (const line of lines.slice(1)) {
		const cells = splitCsvLine(line, delimiter);
		const row: Record<string, string> = {};
		header.forEach((column, index) => {
			row[column] = cells[index] ?? '';
		});

		if (format.rowFails?.(row)) {
			result.total_rows += 1;
			result.failed += 1;
			continue;
		}
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
