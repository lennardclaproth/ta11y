import { ApiError, apiGet, apiSend } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	CatalogueSearchRequest,
	CatalogueSearchResponse,
	CatalogueStatus,
	CatalogueSync,
	CatalogueSyncRequest,
	CreateListingRequest,
	EODQuery,
	EODResponse,
	Listing,
	ListingSearchRow,
	ListingsResponse,
	ListingsSearchQuery,
	ListingsSearchResponse,
	ProviderCredential,
	RevealProviderAPIKeyResponse,
	UpdateListingFieldsRequest,
	UpdateProviderCredentialsRequest
} from '$lib/api/types';
import {
	catalogueEntries,
	catalogueSync,
	upstreamOnlyEntries,
	upstreamTotals,
	type CatalogueEntry
} from '$lib/data/fixtures/catalogue';
import { maskKey, mockKeys, providerCredentials } from '$lib/data/fixtures/credentials';
import { eodByListing, listings } from '$lib/data/fixtures/marketdata';
import { clone, delay, mockId } from './_mock';

/** Mirrors the backend's adoptability rules so mock rows behave like live ones. */
function adoptability(
	entry: CatalogueEntry
): Pick<ListingSearchRow, 'adoptable' | 'adoptable_reason'> {
	if (!entry.name || !entry.name.trim()) {
		return { adoptable: false, adoptable_reason: 'the provider reports no name for this symbol' };
	}
	if (!entry.has_eod) {
		return {
			adoptable: false,
			adoptable_reason: 'the provider has no end-of-day history for this symbol'
		};
	}
	return { adoptable: true, adoptable_reason: null };
}

function trackedRow(listing: Listing): ListingSearchRow {
	return {
		...listing,
		exchange_mic: null,
		tracked: true,
		has_eod: true,
		adoptable: false,
		adoptable_reason: 'this instrument is already in your listings'
	};
}

function catalogueRow(entry: CatalogueEntry): ListingSearchRow {
	return { ...entry, tracked: false, ...adoptability(entry) };
}

/** Case-insensitive substring match, the same shape the backend's SQL LIKE uses. */
function matches(row: { symbol: string; name: string | null; isin?: string | null }, q: string) {
	const needle = q.trim().toLowerCase();
	if (!needle) return true;
	return (
		row.symbol.toLowerCase().includes(needle) ||
		(row.name ?? '').toLowerCase().includes(needle) ||
		(row.isin ?? '').toLowerCase().includes(needle)
	);
}

/**
 * Builds the mock combined result set: tracked listings first, then cached
 * catalogue entries that are not already tracked.
 */
function combinedRows(q: string, scope: ListingsSearchQuery['scope']): ListingSearchRow[] {
	const tracked = listings.filter((listing) => matches(listing, q)).map(trackedRow);
	const trackedSymbols = new Set(listings.map((listing) => listing.symbol));
	const catalogue = catalogueEntries
		.filter((entry) => !trackedSymbols.has(entry.symbol) && matches(entry, q))
		.map(catalogueRow);

	if (scope === 'catalogue') return catalogue;
	if (scope === 'all') return [...tracked, ...catalogue];
	return tracked;
}

/** `GET /marketdata/listings` */
export async function listListings(): Promise<ListingsResponse> {
	if (useMocks) {
		await delay();
		return clone(listings.slice().sort((a, b) => a.symbol.localeCompare(b.symbol)));
	}
	return apiGet<ListingsResponse>('/marketdata/listings');
}

/** `GET /marketdata/listings/search` */
export async function searchListings(query: ListingsSearchQuery): Promise<ListingsSearchResponse> {
	if (useMocks) {
		await delay();
		const matched = combinedRows(query.q, query.scope);
		const limit = query.limit ?? 25;
		const offset = query.offset ?? 0;
		const data = clone(matched.slice(offset, offset + limit));
		return { pagination: { limit, offset, count: data.length, total: matched.length }, data };
	}
	return apiGet<ListingsSearchResponse>('/marketdata/listings/search', { ...query });
}

/** `GET /marketdata/eods` */
export async function getEOD(query: EODQuery): Promise<EODResponse> {
	if (useMocks) {
		await delay();
		const listing = query.listing_id
			? listings.find((l) => l.id === query.listing_id)
			: listings.find((l) => l.symbol.toLowerCase() === (query.symbol ?? '').toLowerCase());
		const all = (listing && eodByListing[listing.id]) ?? [];
		const filtered = all.filter(
			(e) =>
				(!query.from || e.Date.slice(0, 10) >= query.from) &&
				(!query.to || e.Date.slice(0, 10) <= query.to)
		);
		const dir = query.sort_order === 'desc' ? -1 : 1;
		const sorted = filtered.slice().sort((a, b) => a.Date.localeCompare(b.Date) * dir);
		const limit = query.limit ?? 100;
		const offset = query.offset ?? 0;
		const data = clone(sorted.slice(offset, offset + limit));
		return {
			Data: data,
			Metadata: { Message: '', ResultCount: data.length, TotalCount: sorted.length }
		};
	}
	return apiGet<EODResponse>('/marketdata/eods', { ...query });
}

/** `POST /marketdata/listing` */
export async function createListing(body: CreateListingRequest): Promise<Listing> {
	if (useMocks) {
		await delay();
		if (
			listings.some((listing) => listing.symbol === body.symbol && listing.source === body.source)
		) {
			throw new ApiError(409, 'Listing already exists', { listing: 'listing already exists' });
		}
		const now = new Date().toISOString();
		const listing: Listing = {
			id: mockId(),
			symbol: body.symbol,
			name: body.name,
			source: body.source,
			description: body.description ?? null,
			exchange: body.exchange ?? null,
			region: body.region ?? null,
			currency: body.currency ?? null,
			isin: body.isin ?? null,
			ticker: body.ticker ?? null,
			type: body.type ?? null,
			created_at: now,
			updated_at: now
		};
		// Share session changes with the list and portfolio listing search in demo mode.
		listings.push(listing);
		return clone(listing);
	}
	return apiSend<Listing>('POST', '/marketdata/listing', body);
}

/** `PATCH /marketdata/listing` */
export async function updateListing(body: UpdateListingFieldsRequest): Promise<Listing> {
	if (useMocks) {
		await delay();
		const existing = listings.find((l) => l.id === body.id) ?? listings[0];
		return clone({ ...existing, ...body, updated_at: new Date().toISOString() });
	}
	return apiSend<Listing>('PATCH', '/marketdata/listing', body);
}

/**
 * `POST /marketdata/catalogue/search` — asks the provider directly and caches the
 * results. Costs one provider request, so call it only on an explicit user action.
 *
 * The local catalogue can never be known to be complete for a query it has not
 * seen: the provider matches substrings across its whole universe, so "ASM" has far
 * more matches than "ASML". That is why widening a search is a deliberate step
 * rather than an automatic fallback.
 */
export async function searchProviderCatalogue(
	body: CatalogueSearchRequest
): Promise<CatalogueSearchResponse> {
	if (useMocks) {
		await delay();
		// Entries the mock cache has never seen only surface here, so the metered
		// action is genuinely the thing that finds them — as it is against the API.
		const discovered = upstreamOnlyEntries.filter((entry) => matches(entry, body.q));
		for (const entry of discovered) {
			if (!catalogueEntries.some((cached) => cached.symbol === entry.symbol)) {
				catalogueEntries.push(entry);
			}
		}
		const limit = body.limit ?? 25;
		const matched = combinedRows(body.q, 'all');
		const data = clone(matched.slice(0, limit));
		const upstream = upstreamTotals[body.q.trim().toLowerCase()] ?? matched.length;
		return {
			pagination: { limit, offset: 0, count: data.length, total: matched.length },
			upstream_total: upstream,
			// The provider returns 100 rows per request; anything beyond that is not
			// in this answer and the UI has to say so.
			truncated: upstream > 100,
			cached: discovered.length,
			data
		};
	}
	return apiSend<CatalogueSearchResponse>('POST', '/marketdata/catalogue/search', body);
}

/**
 * `POST /marketdata/catalogue/sync` — starts a bounded background seed run that
 * caches the provider's most-traded entries. One provider request per page.
 */
export async function startCatalogueSync(body: CatalogueSyncRequest): Promise<CatalogueSync> {
	if (useMocks) {
		await delay();
		return clone({
			...catalogueSync,
			id: mockId(),
			source: body.source,
			status: 'running',
			pages_fetched: 0,
			rows_upserted: 0,
			started_at: new Date().toISOString(),
			finished_at: null
		});
	}
	return apiSend<CatalogueSync>('POST', '/marketdata/catalogue/sync', body);
}

/** `GET /marketdata/catalogue/status` */
export async function getCatalogueStatus(source: string): Promise<CatalogueStatus> {
	if (useMocks) {
		await delay();
		return clone({ source, entries: catalogueEntries.length, latest_sync: catalogueSync });
	}
	return apiGet<CatalogueStatus>('/marketdata/catalogue/status', { source });
}

/**
 * `GET /marketdata/providers` — external provider connection records.
 *
 * Keys are never returned here, only a masked hint; use `revealProviderApiKey` when a
 * key genuinely needs to be read back.
 */
export async function listProviderCredentials(): Promise<ProviderCredential[]> {
	if (useMocks) {
		await delay();
		return clone(providerCredentials);
	}
	return apiGet<ProviderCredential[]>('/marketdata/providers');
}

/** `PATCH /marketdata/providers/{id}/credentials` */
export async function updateProviderCredentials(
	id: string,
	body: UpdateProviderCredentialsRequest
): Promise<ProviderCredential> {
	if (useMocks) {
		await delay();
		const existing = providerCredentials.find((provider) => provider.id === id);
		if (!existing)
			throw new ApiError(404, 'Provider not found', { provider: 'provider not found' });
		if (existing.ingestion_mode === 'MANUAL') {
			throw new ApiError(409, 'Manual providers have no credentials to configure', {
				provider: 'manual providers have no credentials to configure'
			});
		}
		if (body.api_key !== undefined) {
			mockKeys[id] = body.api_key.trim();
			existing.has_api_key = mockKeys[id].length > 0;
			existing.api_key_hint = maskKey(mockKeys[id]);
		}
		if (body.base_uri !== undefined) existing.base_uri = body.base_uri.trim();
		return clone(existing);
	}
	return apiSend<ProviderCredential>('PATCH', `/marketdata/providers/${id}/credentials`, body);
}

/**
 * `POST /marketdata/providers/{id}/credentials/reveal` — returns the stored key in full.
 * Every call is logged server-side, so treat it as a deliberate action.
 */
export async function revealProviderApiKey(id: string): Promise<RevealProviderAPIKeyResponse> {
	if (useMocks) {
		await delay();
		const key = mockKeys[id];
		if (!key) {
			throw new ApiError(400, 'No API key configured', {
				credentials: 'provider API key cannot be empty'
			});
		}
		return { id, api_key: key };
	}
	return apiSend<RevealProviderAPIKeyResponse>(
		'POST',
		`/marketdata/providers/${id}/credentials/reveal`
	);
}
