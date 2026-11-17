export const countStatSizes = ['sm', 'md', 'lg'] as const;
export type CountStatSize = (typeof countStatSizes)[number];
