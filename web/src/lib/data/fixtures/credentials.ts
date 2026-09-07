import type { ProviderCredential } from '$lib/api/types';

/**
 * Masks a key the way the backend's `marketdata.MaskAPIKey` does, so mock hints look
 * like live ones. Short keys are fully masked rather than mostly revealed.
 */
export function maskKey(key: string): string {
	const trimmed = key.trim();
	if (trimmed === '') return '';
	if (trimmed.length <= 8) return '*'.repeat(trimmed.length);
	return `****${trimmed.slice(-4)}`;
}

/**
 * The keys the mock backend "stores". Kept apart from the credential records so the
 * mock mirrors the real contract: listing never exposes a key, only revealing does.
 */
export const mockKeys: Record<string, string> = {
	'prv-marketstack': 'ms_demo_key_2f9c41ab',
	'prv-alphavantage': 'av_demo_key_77d1'
};

/** Provider connection records, mirroring `GET /marketdata/providers`. */
export const providerCredentials: ProviderCredential[] = [
	{
		id: 'prv-marketstack',
		name: 'marketstack',
		ingestion_mode: 'API',
		base_uri: 'https://api.marketstack.com/v2',
		has_api_key: true,
		api_key_hint: maskKey(mockKeys['prv-marketstack']),
		remaining: 4821,
		used: 179,
		total: 5000,
		resets_at: '2026-10-01T00:00:00Z'
	},
	{
		id: 'prv-alphavantage',
		name: 'alphavantage',
		ingestion_mode: 'API',
		base_uri: 'https://www.alphavantage.co',
		has_api_key: true,
		api_key_hint: maskKey(mockKeys['prv-alphavantage']),
		remaining: 0,
		used: 0,
		total: 0
	},
	{
		// A manual provider ingests uploaded files, so there is no endpoint or key to set.
		id: 'prv-brandnewday',
		name: 'brandnewday',
		ingestion_mode: 'MANUAL',
		has_api_key: false,
		api_key_hint: '',
		remaining: 0,
		used: 0,
		total: 0
	}
];
