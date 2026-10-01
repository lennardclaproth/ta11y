import {
	addDaysISO,
	addMonthsUTC,
	formatDisplayDate,
	parseISODate,
	toISODate,
	todayISO
} from '$lib/components/molecules/calendar/calendar.utils';

/**
 * The one period of the app.
 *
 * Every page used to keep its own `from`/`to`, so changing page started over. The period now
 * lives here: the account overview owns the control, the charts and the table of whichever page
 * you are on read it, and it survives navigating between pages. Admin pages have no period and
 * simply never read it.
 */

export const periodPresets = ['1m', '3m', '6m', 'ytd', '1y', '3y', '5y', 'max', 'custom'] as const;
export type PeriodPreset = (typeof periodPresets)[number];

export type PeriodRange = { from: string; to: string; preset: PeriodPreset };

/** Short ranges, long ranges, then Max, in reading order. `custom` is never offered as a choice. */
export const periodPresetLabels = {
	'1m': '1M',
	'3m': '3M',
	'6m': '6M',
	ytd: 'YTD',
	'1y': '1Y',
	'3y': '3Y',
	'5y': '5Y',
	max: 'Max',
	custom: 'Custom'
} satisfies Record<PeriodPreset, string>;

/** The presets offered in the picker, in reading order. */
export const periodPresetOrder: PeriodPreset[] = ['1m', '3m', '6m', 'ytd', '1y', '3y', '5y', 'max'];

function monthsBack(today: string, months: number): string {
	const base = parseISODate(today);
	if (!base) return today;
	return toISODate(addMonthsUTC(base, -months));
}

let from = $state(`${todayISO().slice(0, 4)}-01-01`);
let to = $state(todayISO());
let preset = $state<PeriodPreset>('ytd');
/**
 * The oldest date the account has data for, used by `Max`. It is filled by whoever knows it
 * (the net-worth snapshots); without it `Max` falls back to the start of the current year
 * rather than inventing a start date.
 */
let earliest = $state<string | null>(null);
/** A range supplied by a URL is only honoured before anything else has set the period. */
let seeded = $state(false);

function rangeFor(value: PeriodPreset): { from: string; to: string } {
	const today = todayISO();
	switch (value) {
		case '1m':
			return { from: addDaysISO(monthsBack(today, 1), 1), to: today };
		case '3m':
			return { from: addDaysISO(monthsBack(today, 3), 1), to: today };
		case '6m':
			return { from: addDaysISO(monthsBack(today, 6), 1), to: today };
		case '1y':
			return { from: addDaysISO(monthsBack(today, 12), 1), to: today };
		case '3y':
			return { from: addDaysISO(monthsBack(today, 36), 1), to: today };
		case '5y':
			return { from: addDaysISO(monthsBack(today, 60), 1), to: today };
		case 'max':
			return { from: earliest ?? `${today.slice(0, 4)}-01-01`, to: today };
		case 'ytd':
		case 'custom':
		default:
			return { from: `${today.slice(0, 4)}-01-01`, to: today };
	}
}

export const periodStore = {
	get from(): string {
		return from;
	},
	get to(): string {
		return to;
	},
	get preset(): PeriodPreset {
		return preset;
	},
	/** The preset's own name, e.g. `YTD`. */
	get presetLabel(): string {
		return periodPresetLabels[preset];
	},
	/** The range the preset resolves to, e.g. `1 Jan 2026 – 30 Jun 2026`. */
	get rangeLabel(): string {
		return `${formatDisplayDate(from)} – ${formatDisplayDate(to)}`;
	},
	/** The presets offered in the picker, already resolved to dates. */
	get options(): { value: PeriodPreset; label: string; from: string; to: string }[] {
		return periodPresetOrder.map((value) => ({
			value,
			label: periodPresetLabels[value],
			...rangeFor(value)
		}));
	},
	/** Commit a range. A range that is not one of the presets is `custom`. */
	set(next: { from: string; to: string; preset?: PeriodPreset }): void {
		seeded = true;
		from = next.from;
		to = next.to;
		preset = next.preset ?? 'custom';
	},
	applyPreset(value: PeriodPreset): void {
		periodStore.set({ ...rangeFor(value), preset: value });
	},
	/**
	 * Adopt a range a URL carried in, but only while nothing has set the period yet: coming back
	 * to a page whose URL still holds an older range must not undo the period chosen since.
	 */
	seed(nextFrom: string | null | undefined, nextTo: string | null | undefined): void {
		if (seeded) return;
		seeded = true;
		if (!nextFrom || !nextTo) return;
		from = nextFrom;
		to = nextTo;
		preset = 'custom';
	},
	/** Tell the store how far back the account's data goes, so `Max` has a real start. */
	setEarliest(date: string | null): void {
		earliest = date;
		if (preset === 'max') {
			const range = rangeFor('max');
			from = range.from;
			to = range.to;
		}
	}
};
