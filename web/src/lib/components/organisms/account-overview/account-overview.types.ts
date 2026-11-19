export const accountOverviewLayouts = ['inline', 'panel'] as const;
/**
 * `inline` is the region in the masthead on a wide screen; `panel` is the same content behind
 * the single entry on a narrow one. There is no separate mobile design, only a different
 * container.
 */
export type AccountOverviewLayout = (typeof accountOverviewLayouts)[number];
