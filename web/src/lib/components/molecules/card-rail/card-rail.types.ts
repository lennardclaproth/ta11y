/** One named group of cards in a rail; `id` matches the children's `data-group`. */
export interface CardRailGroup {
	id: string;
	label: string;
}

/** How wide a card sits in the band; the classes live in `card-rail.variants.ts`. */
export const cardRailWidthNames = ['default', 'wide', 'solo'] as const;
export type CardRailWidth = (typeof cardRailWidthNames)[number];
