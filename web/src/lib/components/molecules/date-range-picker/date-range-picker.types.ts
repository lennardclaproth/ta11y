export const dateRangePickerSizes = ['sm', 'md', 'lg'] as const;
export type DateRangePickerSize = (typeof dateRangePickerSizes)[number];

/**
 * A named range offered inside the picker. Supplying these replaces the picker's own
 * list, so a caller that owns the vocabulary of its periods (1M, YTD, Max…) can show it
 * in the picker instead of parking a second control next to it.
 */
export type DateRangePreset = {
	value: string;
	label: string;
	from: string;
	to: string;
};

/** The value `preset` takes once days are picked by hand. */
export const customPresetValue = 'custom';
