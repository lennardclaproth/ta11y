import { ApiError, apiUpload } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import { numberToScaled } from '$lib/api/money';
import type {
	CashflowImportInput,
	EOD,
	EODImportInput,
	ImportAcceptedResponse,
	PortfolioImportInput
} from '$lib/api/types';
import { eodByListing, listings } from '$lib/data/fixtures/marketdata';
import { delay, mockId } from './_mock';

function mockAccepted(): ImportAcceptedResponse {
	return { import_id: mockId(), status: 'pending' };
}

/** `POST /imports/cashflow` (multipart) */
export async function importCashflow(input: CashflowImportInput): Promise<ImportAcceptedResponse> {
	if (useMocks) {
		await delay();
		return mockAccepted();
	}
	const form = new FormData();
	form.append('file', input.file);
	form.append('vendor_id', input.vendor_id);
	return apiUpload<ImportAcceptedResponse>('/imports/cashflow', form);
}

/** `POST /imports/portfolio` (multipart) */
export async function importPortfolio(
	input: PortfolioImportInput
): Promise<ImportAcceptedResponse> {
	if (useMocks) {
		await delay();
		return mockAccepted();
	}
	const form = new FormData();
	form.append('file', input.file);
	form.append('vendor_id', input.vendor_id);
	return apiUpload<ImportAcceptedResponse>('/imports/portfolio', form);
}

/** `POST /imports/eod` (multipart) */
export async function importEOD(input: EODImportInput): Promise<ImportAcceptedResponse> {
	if (useMocks) {
		await delay();
		const listing = listings.find((item) => item.id === input.listing_id);
		if (!listing) {
			throw new ApiError(404, 'Listing not found', { listing_id: 'import listing not found' });
		}
		// The API refuses uploads for listings priced by an API provider; the fixture layer has
		// no provider table, so the listing's source stands in for the provider's ingestion mode.
		if (!manualSources.has(listing.source)) {
			throw new ApiError(422, 'Provider is not manual', {
				listing_id: 'import provider is not manual'
			});
		}
		await applyMockEODCsv(input);
		return mockAccepted();
	}
	const form = new FormData();
	form.append('file', input.file);
	form.append('listing_id', input.listing_id);
	return apiUpload<ImportAcceptedResponse>('/imports/eod', form);
}

/**
 * Sources whose provider ingests uploaded files. The real check is the provider's
 * `ingestion_mode`; mock mode has no provider table, so the source stands in for it.
 */
const manualSources = new Set(['brandnewday', 'brand_new_day']);

/**
 * Parses an uploaded CSV the way `internal/importer/eod/parsers/brandnewday.go` does and
 * appends the rows to the fixture, so fixture mode shows the prices that were actually
 * uploaded rather than accepting the file into nothing. Rows that do not parse are dropped,
 * matching an import whose per-row errors are counted but never reported to the client.
 */
async function applyMockEODCsv(input: EODImportInput): Promise<void> {
	const text = await input.file.text();
	const lines = text
		.split('\n')
		.map((line) => line.trim())
		.filter((line) => line !== '');
	if (lines.length < 2) return;

	const header = lines[0].split(',').map((cell) => cell.trim().toLowerCase());
	const dateAt = header.indexOf('date');
	const navAt = header.indexOf('nav');
	if (dateAt === -1 || navAt === -1) return;

	const symbol = listings.find((item) => item.id === input.listing_id)?.symbol ?? '';
	const existing = eodByListing[input.listing_id] ?? [];
	const seen = new Set(existing.map((row) => row.Date.slice(0, 10)));
	const added: EOD[] = [];

	for (const line of lines.slice(1)) {
		const cells = line.split(',');
		const [day, month, year] = (cells[dateAt] ?? '').trim().split('/');
		const nav = Number((cells[navAt] ?? '').trim());
		if (!day || !month || !year || !Number.isFinite(nav)) continue;

		const date = `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`;
		if (seen.has(date)) continue;
		seen.add(date);

		const scaled = numberToScaled(nav);
		const now = new Date().toISOString();
		added.push({
			ID: mockId(),
			ListingID: input.listing_id,
			Symbol: symbol,
			Date: `${date}T00:00:00Z`,
			// A fund publishes one NAV a day, so the candle is flat -- as the real parser stores it.
			Open: scaled,
			High: scaled,
			Low: scaled,
			Close: scaled,
			Volume: 0,
			CreatedAt: now,
			UpdatedAt: now
		});
	}

	eodByListing[input.listing_id] = [...existing, ...added].sort((a, b) =>
		a.Date.localeCompare(b.Date)
	);
}
