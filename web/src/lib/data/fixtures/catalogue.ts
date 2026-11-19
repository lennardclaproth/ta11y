import type { CatalogueSync, ListingSearchRow } from '$lib/api/types';

/**
 * A cached provider-catalogue entry, before it is joined against the user's own
 * listings. Ids are stable so selection survives a re-render.
 */
export type CatalogueEntry = Pick<
	ListingSearchRow,
	'id' | 'symbol' | 'name' | 'source' | 'exchange' | 'exchange_mic' | 'has_eod'
>;

/**
 * Entries already cached locally, so searching them costs nothing.
 *
 * Shaped after real Marketstack responses, awkward rows included: the same company
 * listed on several exchanges under different symbols, a symbol the provider has no
 * name for, and one with no price history. Those last two are what the drawer has
 * to refuse to adopt.
 */
export const catalogueEntries: CatalogueEntry[] = [
	{
		id: 'cat-asml-xnas',
		symbol: 'ASML',
		name: 'ASML Holding NV',
		source: 'market_stack',
		exchange: 'NASDAQ - ALL MARKETS',
		exchange_mic: 'XNAS',
		has_eod: true
	},
	{
		id: 'cat-asml-xams',
		symbol: 'ASML.XAMS',
		name: 'ASML HOLDING',
		source: 'market_stack',
		exchange: 'EURONEXT - EURONEXT AMSTERDAM',
		exchange_mic: 'XAMS',
		has_eod: true
	},
	{
		id: 'cat-asmlf',
		symbol: 'ASMLF',
		name: 'ASML Holding NV',
		source: 'market_stack',
		exchange: 'OTC LINK ATS - OTC MARKETS',
		exchange_mic: 'OTCM',
		has_eod: true
	},
	{
		id: 'cat-asmcx',
		symbol: 'ASMCX',
		name: null,
		source: 'market_stack',
		exchange: 'US MUTUAL FUNDS',
		exchange_mic: 'NMFQS',
		has_eod: true
	},
	{
		id: 'cat-asmb',
		symbol: 'ASMB',
		name: 'Assembly Biosciences Inc',
		source: 'market_stack',
		exchange: 'NASDAQ - ALL MARKETS',
		exchange_mic: 'XNAS',
		has_eod: false
	},
	{
		id: 'cat-vwrl-as',
		symbol: 'VWRL.AS',
		name: 'Vanguard FTSE All-World UCITS ETF',
		source: 'market_stack',
		exchange: 'EURONEXT - EURONEXT AMSTERDAM',
		exchange_mic: 'XAMS',
		has_eod: true
	},
	{
		id: 'cat-msft',
		symbol: 'MSFT',
		name: 'Microsoft Corporation',
		source: 'market_stack',
		exchange: 'NASDAQ - ALL MARKETS',
		exchange_mic: 'XNAS',
		has_eod: true
	}
];

/**
 * Entries the local cache has never seen. They only appear after an explicit
 * provider search, which is what makes the metered "search the provider" action
 * demonstrable on mocks: the cache genuinely cannot answer for them.
 */
export const upstreamOnlyEntries: CatalogueEntry[] = [
	{
		id: 'cat-vwrl-l',
		symbol: 'VWRL.L',
		name: 'Vanguard FTSE All-World UCITS ETF',
		source: 'market_stack',
		exchange: 'LONDON STOCK EXCHANGE',
		exchange_mic: 'XLON',
		has_eod: true
	},
	{
		id: 'cat-vwrl-sw',
		symbol: 'VWRL.SW',
		name: 'Vanguard FTSE All-World UCITS ETF',
		source: 'market_stack',
		exchange: 'SIX SWISS EXCHANGE',
		exchange_mic: 'XSWX',
		has_eod: true
	},
	{
		id: 'cat-vwrl-aq',
		symbol: 'VWRL.AQ',
		name: null,
		source: 'market_stack',
		exchange: 'AQUIS EXCHANGE',
		exchange_mic: 'AQXE',
		has_eod: true
	},
	{
		id: 'cat-tdt-as',
		symbol: 'TDT.AS',
		name: 'Triodos Fair Share Fund',
		source: 'market_stack',
		exchange: 'EURONEXT - EURONEXT AMSTERDAM',
		exchange_mic: 'XAMS',
		has_eod: true
	}
];

/** How many entries the provider claims to match, per query, for the mock. */
export const upstreamTotals: Record<string, number> = {
	asml: 31,
	asm: 125,
	vwrl: 9,
	vanguard: 1701,
	micro: 220
};

/** The most recent seed run reported by the mock catalogue status. */
export const catalogueSync: CatalogueSync = {
	id: 'cat-sync-1',
	source: 'market_stack',
	status: 'completed',
	pages_fetched: 20,
	rows_upserted: 2000,
	upstream_total: 683174,
	started_at: '2026-06-01T09:00:00Z',
	finished_at: '2026-06-01T09:00:31Z'
};
