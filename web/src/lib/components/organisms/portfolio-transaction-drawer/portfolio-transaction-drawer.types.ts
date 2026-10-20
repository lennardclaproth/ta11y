import type { PortfolioTransactionType } from '$lib/api/types';

/** The transaction types a manual portfolio transaction can take, in form order. */
export const portfolioTransactionTypeOptions = [
	{ value: 'BUY', label: 'Buy' },
	{ value: 'SELL', label: 'Sell' },
	{ value: 'DIVIDEND', label: 'Dividend' },
	{ value: 'TAX', label: 'Tax' },
	{ value: 'FEE', label: 'Fee' },
	{ value: 'CASH', label: 'Cash' }
] as const satisfies readonly { value: PortfolioTransactionType; label: string }[];

/** Backend rule: every type except CASH names a listing. */
export function typeNeedsListing(type: PortfolioTransactionType): boolean {
	return type !== 'CASH';
}

/** Backend rule: BUY and SELL carry a quantity; no other type may. */
export function typeNeedsQuantity(type: PortfolioTransactionType): boolean {
	return type === 'BUY' || type === 'SELL';
}
