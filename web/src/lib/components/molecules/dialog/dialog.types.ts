export const dialogSizes = ['sm', 'md', 'lg', 'xl', '2xl'] as const;

export type DialogSize = (typeof dialogSizes)[number];

export const dialogSizeClasses = {
	sm: 'max-w-sm',
	md: 'max-w-md',
	lg: 'max-w-lg',
	xl: 'max-w-2xl',
	/** Two columns side by side: a list to pick from next to what the pick changes. */
	'2xl': 'max-w-4xl'
} satisfies Record<DialogSize, string>;
