/**
 * Origin conventions on the wire (mirrors the Go backend).
 *
 * Cashflow carries the origin in `source`: `"manual"` or `"manual:<vendor>"` for a transaction
 * entered by hand, the vendor name alone (`"ing"`, `"n26"`) for an imported one. Portfolio has a
 * dedicated `origin` field instead. Only manual transactions can be moved to another date — an
 * imported row's dedup checksum carries its date, so moving it would make the next import of the
 * same statement insert it again.
 */
import type { CashflowTransaction, PortfolioTransaction } from './types';

const MANUAL_SOURCE = 'manual';

/** Whether a cashflow transaction was entered by hand rather than imported. */
export function isManualCashflowTransaction(tx: CashflowTransaction): boolean {
	return tx.source === MANUAL_SOURCE || tx.source.startsWith(`${MANUAL_SOURCE}:`);
}

/** Short label for where a cashflow transaction came from, e.g. "Manual" or "ING". */
export function cashflowOriginLabel(tx: CashflowTransaction): string {
	if (tx.source === MANUAL_SOURCE) return 'Manual';
	if (tx.source.startsWith(`${MANUAL_SOURCE}:`)) {
		return `Manual · ${tx.source.slice(MANUAL_SOURCE.length + 1)}`;
	}
	return tx.source.toUpperCase() || 'Unknown';
}

/**
 * The tag filter's "Untagged" choice. The tag totals name the empty bucket `untagged`, but a
 * transaction without a tag stores an empty one, so this is sent as `untagged=true`, never as
 * a tag to match.
 */
export const UNTAGGED_FILTER = 'untagged';

/** Splits the ledger's tag filter into the tags to match and the "Untagged" choice. */
export function cashflowTagQuery(selected: string[]): { tags?: string; untagged?: boolean } {
	const tags = selected.filter((tag) => tag !== UNTAGGED_FILTER);
	return {
		tags: tags.join(',') || undefined,
		untagged: selected.includes(UNTAGGED_FILTER) || undefined
	};
}

/** Whether a portfolio transaction was entered by hand rather than imported. */
export function isManualPortfolioTransaction(tx: PortfolioTransaction): boolean {
	return tx.origin === 'MANUAL';
}

/** Short label for where a portfolio transaction came from. */
export function portfolioOriginLabel(tx: PortfolioTransaction): string {
	return tx.origin === 'MANUAL' ? 'Manual' : 'Import';
}
