import type { UpdateListingFieldsRequest } from '$lib/api/types';

/**
 * Listing metadata fields, shared by the create and edit forms so the two never drift.
 *
 * These are exactly the fields `PATCH /marketdata/listing` accepts. Symbol, name and source
 * are deliberately absent: they identify the instrument at the provider, so changing one
 * would silently repoint every price and position that already resolved through it.
 */
export const listingMetadataFields = [
	{ key: 'exchange', label: 'Exchange', placeholder: 'e.g. XAMS' },
	{ key: 'isin', label: 'ISIN', placeholder: 'e.g. IE00B3RBWM25' },
	{ key: 'ticker', label: 'Ticker', placeholder: 'e.g. VWRL' },
	{ key: 'region', label: 'Region', placeholder: 'e.g. Netherlands' },
	{ key: 'type', label: 'Type', placeholder: 'e.g. ETF' },
	{ key: 'description', label: 'Description', placeholder: 'Additional information' }
] as const;

export type ListingMetadataKey = (typeof listingMetadataFields)[number]['key'];

/** Currencies offered for a listing, plus the "leave unset" choice. */
export const listingCurrencyOptions = [
	{ value: '', label: 'Not specified' },
	...['EUR', 'USD', 'GBP', 'JPY'].map((value) => ({ value, label: value }))
];

/** Data sources a listing can be created against. */
export const listingSources = [
	{ value: 'market_stack', label: 'Marketstack' },
	{ value: 'alpha_vantage', label: 'Alpha Vantage' },
	{ value: 'brandnewday', label: 'Brand New Day (manual uploads)' }
];

/** Every field the edit form can send, keyed as the patch request expects them. */
export type ListingEditableKey = ListingMetadataKey | 'currency';

/** The patch payload, minus the id the caller supplies. */
export type ListingFieldPatch = Omit<UpdateListingFieldsRequest, 'id'>;
