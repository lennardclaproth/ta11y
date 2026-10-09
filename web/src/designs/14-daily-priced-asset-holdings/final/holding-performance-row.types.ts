/**
 * Option vocabulary for the proposed `holding-performance-row` molecule (design #14).
 *
 * - `expand` — the row opens its own performance block in place. Used by the class drawer on
 *   Assets and by the item ledger on a narrow class page.
 * - `select` — the row only marks itself as the current item. Used by the class page's left
 *   ledger, where the reading column already carries the detail.
 */
export const holdingRowModes = ['expand', 'select'] as const;
export type HoldingRowMode = (typeof holdingRowModes)[number];
