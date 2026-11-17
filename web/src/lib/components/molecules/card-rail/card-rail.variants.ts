import type { CardRailWidth } from './card-rail.types';

/**
 * Card widths for a rail. One vocabulary across the pages that use it: a full-width card on
 * a phone, a fixed column on desktop, so a group of cards adds up to a full band and
 * jumping to it lands flush against the left edge instead of leaving a sliver behind.
 * Every utility is a whole literal, as the Tailwind 4 scanner needs.
 */
export const cardRailWidths = {
	/** A quarter-band card: a donut, a streak. */
	default: 'w-[84vw] shrink-0 snap-start sm:w-[20rem]',
	/** A half-band card: a trend chart, the running month, the monthly standing. */
	wide: 'w-[84vw] shrink-0 snap-start sm:w-[24rem] lg:w-[32rem]',
	/** A group of one card has to fill the band itself, or it can never reach the left edge. */
	solo: 'w-[84vw] shrink-0 snap-start sm:w-full'
} satisfies Record<CardRailWidth, string>;
