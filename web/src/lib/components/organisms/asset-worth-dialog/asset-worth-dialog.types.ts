/**
 * The two ways the worth of a hand-valued item changes.
 *
 * - `adjust` — money added or withdrawn, e.g. €300 transferred to savings. Paid in moves with it.
 * - `set` — the item is now worth a different amount on its own, e.g. after interest or a
 *   revaluation. Paid in stays where it was, so the difference shows up as growth.
 */
export const worthChangeModes = ['adjust', 'set'] as const;
export type WorthChangeMode = (typeof worthChangeModes)[number];

/** The item whose worth is changed, as the class payload already carries it. */
export interface WorthDialogAsset {
	id: string;
	name: string;
	current_worth: string;
}
