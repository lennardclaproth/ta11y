/**
 * The two steps of adding an item to an asset class.
 *
 * - `quote` — how the item is valued, and for a daily-priced one which instrument it follows.
 * - `purchase` — when, how much and at what price. Reached on its own when a purchase is added
 *   to an item that already exists.
 */
export const addItemSteps = ['quote', 'purchase'] as const;
export type AddItemStep = (typeof addItemSteps)[number];
