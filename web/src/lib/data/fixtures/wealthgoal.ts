import type { WealthGoal } from '$lib/api/types';

/**
 * The monthly goal the fixtures run on [034]. The standing itself is not a fixture: it is
 * derived from `cashflowTransactions` in the service, so marking a row in mock mode moves
 * the numbers the same way it does against the real API.
 */
export const wealthGoal: WealthGoal = {
	share_percent: 30,
	effective_from: '2026-01-01'
};
