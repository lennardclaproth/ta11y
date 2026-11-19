/**
 * What the dialog hands back: the tag to put on every selected transaction, an existing one or a
 * new one, or `''` to take their tag off. A transaction carries one tag, so either replaces it.
 */
export type TagTransactionsValue = { tag: string };

/** How many selected rows the preview table shows before it says how many more there are. */
export const tagPreviewLimit = 50;
