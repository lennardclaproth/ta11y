import type { PaginatedResponse } from './common';

/** One market listing. Mirrors `marketdata.ListingResponse`. */
export interface Listing {
	id: string;
	symbol: string;
	name: string;
	source: string;
	description?: string | null;
	exchange?: string | null;
	region?: string | null;
	currency?: string | null;
	isin?: string | null;
	ticker?: string | null;
	type?: string | null;
	created_at: string;
	updated_at: string;
}

/** `GET /marketdata/listings` returns a bare array of listings. */
export type ListingsResponse = Listing[];

/**
 * One row of `GET /marketdata/listings/search` — mirrors `marketdata.ListingSearchRow`.
 *
 * A row is either a listing the user tracks (`tracked: true`, every listing field
 * populated) or a cached provider-catalogue entry that could become one. Catalogue
 * entries omit what provider ticker search cannot supply, so `name` may be null and
 * the timestamps absent.
 */
export interface ListingSearchRow {
	id: string;
	symbol: string;
	name: string | null;
	source: string;
	description?: string | null;
	exchange?: string | null;
	/** Market Identifier Code, e.g. `XAMS`. Only catalogue rows carry it. */
	exchange_mic?: string | null;
	region?: string | null;
	currency?: string | null;
	isin?: string | null;
	ticker?: string | null;
	type?: string | null;
	/** True when this row is already one of the user's listings. */
	tracked: boolean;
	/** Whether the provider holds end-of-day history for the symbol. */
	has_eod: boolean;
	/** False when the row cannot become a listing; see `adoptable_reason`. */
	adoptable: boolean;
	adoptable_reason?: string | null;
	created_at?: string;
	updated_at?: string;
}

/** `GET /marketdata/listings/search` — mirrors `marketdata.ListingsSearchResponse`. */
export type ListingsSearchResponse = PaginatedResponse<ListingSearchRow>;

/** Which sides of the catalogue a listing search covers. */
export type CatalogueScope = 'tracked' | 'catalogue' | 'all';

/** `POST /marketdata/catalogue/search` request. Costs one provider request. */
export interface CatalogueSearchRequest {
	source: string;
	q: string;
	limit?: number;
}

/** `POST /marketdata/catalogue/search` response. */
export interface CatalogueSearchResponse extends PaginatedResponse<ListingSearchRow> {
	/** How many entries the provider matched in total, often far more than one page. */
	upstream_total: number;
	/** True when the provider matched more than the single page that was fetched. */
	truncated: boolean;
	/** How many catalogue entries this search cached locally. */
	cached: number;
}

/** One bounded catalogue seed run — mirrors `marketdata.CatalogueSync`. */
export interface CatalogueSync {
	id: string;
	source: string;
	status: 'running' | 'completed' | 'failed';
	pages_fetched: number;
	rows_upserted: number;
	upstream_total?: number | null;
	last_error?: string | null;
	started_at: string;
	finished_at?: string | null;
}

/** `POST /marketdata/catalogue/sync` request. */
export interface CatalogueSyncRequest {
	source: string;
	/** Page budget; omitted uses the backend default. One provider request per page. */
	pages?: number;
}

/** `GET /marketdata/catalogue/status` response. */
export interface CatalogueStatus {
	source: string;
	/** Entries cached locally, searchable without spending a provider request. */
	entries: number;
	latest_sync?: CatalogueSync | null;
}

/**
 * One end-of-day OHLCV row. Mirrors the Go `marketdata.EOD` struct, which is serialized **without**
 * json tags — so the keys are PascalCase and prices are raw 1e6-scaled integers.
 */
export interface EOD {
	ID: string;
	ListingID: string;
	Symbol: string;
	/** RFC3339 timestamp. */
	Date: string;
	Open: number;
	Close: number;
	High: number;
	Low: number;
	Volume: number;
	CreatedAt: string;
	UpdatedAt: string;
}

/** EOD retrieval metadata. Mirrors `marketdata.GetEODMetadataResponse` (also PascalCase). */
export interface EODMetadata {
	Message: string;
	ResultCount: number;
	TotalCount: number;
}

/** `GET /marketdata/eods` — mirrors `marketdata.GetEODResponse` (PascalCase envelope). */
export interface EODResponse {
	Data: EOD[];
	Metadata: EODMetadata;
}

/** `POST /marketdata/listing` request — mirrors `marketdata.CreateListingRequest`. */
export interface CreateListingRequest {
	name: string;
	symbol: string;
	source: string;
	description?: string;
	exchange?: string;
	region?: string;
	/** ISO currency code; validated by the backend if present. */
	currency?: string;
	isin?: string;
	ticker?: string;
	/**
	 * Whether creating the listing immediately backfills price history. Defaults to
	 * true on the backend. The catalogue drawer sends false when adopting a batch,
	 * because the backfill is synchronous and one per listing would stall the request.
	 */
	sync_prices?: boolean;
	type?: string;
}

/** `PATCH /marketdata/listing` request — mirrors `marketdata.UpdateListingFieldsRequest`. */
export interface UpdateListingFieldsRequest {
	id: string;
	description?: string;
	exchange?: string;
	region?: string;
	currency?: string;
	isin?: string;
	ticker?: string;
	type?: string;
}

/** Query filters for `GET /marketdata/listings/search`. */
export interface ListingsSearchQuery {
	q: string;
	limit?: number;
	offset?: number;
	/** Defaults to `tracked` on the backend, which is the pre-catalogue behaviour. */
	scope?: CatalogueScope;
}

/** Query filters for `GET /marketdata/eods`. */
export interface EODQuery {
	listing_id?: string;
	symbol?: string;
	from?: string;
	to?: string;
	sort_order?: 'asc' | 'desc';
	limit?: number;
	offset?: number;
}

/**
 * One external market-data provider's connection record. The stored API key is never
 * part of this payload: `has_api_key` says whether one is configured and
 * `api_key_hint` identifies it (e.g. `****9876`) without disclosing it.
 */
export interface ProviderCredential {
	id: string;
	name: string;
	ingestion_mode: 'API' | 'MANUAL';
	base_uri?: string | null;
	has_api_key: boolean;
	api_key_hint: string;
	remaining: number;
	used: number;
	total: number;
	resets_at?: string | null;
}

/** `PATCH /marketdata/providers/{id}/credentials` — omitted fields are left unchanged. */
export interface UpdateProviderCredentialsRequest {
	api_key?: string;
	base_uri?: string;
}

/** `POST /marketdata/providers/{id}/credentials/reveal` — returns the key in full. */
export interface RevealProviderAPIKeyResponse {
	id: string;
	api_key: string;
}
