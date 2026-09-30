<script lang="ts">
	import Popover from '$lib/components/molecules/popover/Popover.svelte';
	import Calendar from '$lib/components/molecules/calendar/Calendar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import {
		addMonthsUTC,
		formatDisplayDate,
		monthLabel,
		parseISODate,
		startOfMonthUTC
	} from '$lib/components/molecules/calendar/calendar.utils';
	import { icons } from '../shared/icons';
	import { firstRecordDate, period } from '../shared/mock-data';
	import { pressClasses } from '../shared/press';

	/**
	 * The one period of the app, as one white card: a calendar glyph, the name of the preset, a
	 * rule, and the range it resolves to. The card is the only control — clicking it opens the
	 * picker, and the presets live inside that picker instead of next to it.
	 *
	 * Build note: this is `DateRangePicker` with two additions — a `trigger` snippet so a caller
	 * can render its own control, and a `presets` prop so this list (1M … YTD) replaces the
	 * picker's built-in one. The calendar, the range logic and Apply/Clear stay as they are.
	 */
	type Props = {
		/** `block` makes the card full width for the narrow overview panel. */
		layout?: 'inline' | 'block';
		showCaption?: boolean;
		/** Opens the picker so the prototype can show it. */
		open?: boolean;
	};

	let { layout = 'inline', showCaption = false, open = $bindable(false) }: Props = $props();

	/**
	 * Short ranges, long ranges and Max, in that reading order. `Max` runs from the first record on
	 * the account, so it is the only preset whose start is data and not arithmetic on today.
	 */
	const presets = [
		{ value: '1m', label: '1M', from: '2026-06-01', to: '2026-06-30' },
		{ value: '3m', label: '3M', from: '2026-04-01', to: '2026-06-30' },
		{ value: '6m', label: '6M', from: '2026-01-01', to: '2026-06-30' },
		{ value: 'ytd', label: 'YTD', from: period.from, to: period.to },
		{ value: '1y', label: '1Y', from: '2025-07-01', to: '2026-06-30' },
		{ value: '3y', label: '3Y', from: '2023-07-01', to: '2026-06-30' },
		{ value: '5y', label: '5Y', from: '2021-07-01', to: '2026-06-30' },
		{ value: 'max', label: 'Max', from: firstRecordDate, to: '2026-06-30' }
	];

	let preset = $state('ytd');
	let from = $state<string>(period.from);
	let to = $state<string>(period.to);

	let draftFrom = $state<string | null>(period.from);
	let draftTo = $state<string | null>(period.to);
	let draftPreset = $state('ytd');
	let hover = $state<string | null>(null);
	let leftMonth = $state(startOfMonthUTC(parseISODate(period.from) ?? new Date()));

	const rightMonth = $derived(addMonthsUTC(leftMonth, 1));
	const presetLabel = $derived(presets.find((p) => p.value === preset)?.label ?? 'Custom');
	const rangeLabel = $derived(`${formatDisplayDate(from)} – ${formatDisplayDate(to)}`);

	function choosePreset(value: string, nextFrom: string, nextTo: string) {
		draftPreset = value;
		draftFrom = nextFrom;
		draftTo = nextTo;
		leftMonth = startOfMonthUTC(parseISODate(nextFrom) ?? new Date());
	}

	function selectDay(iso: string) {
		// Any hand-picked day makes the range custom again.
		draftPreset = 'custom';
		if (!draftFrom || (draftFrom && draftTo)) {
			draftFrom = iso;
			draftTo = null;
			return;
		}
		if (iso < draftFrom) {
			draftTo = draftFrom;
			draftFrom = iso;
		} else {
			draftTo = iso;
		}
	}

	function apply() {
		if (!draftFrom) return;
		from = draftFrom;
		to = draftTo ?? draftFrom;
		preset = draftPreset;
		open = false;
	}
</script>

<div class={layout === 'block' ? 'flex flex-col items-stretch gap-2' : 'flex flex-col items-end'}>
	<Popover bind:open placement="bottom-end" class="p-3">
		{#snippet trigger(api)}
			<button
				type="button"
				aria-expanded={api.open}
				aria-label="Period: {presetLabel}, {rangeLabel}. Opens the date picker."
				onclick={api.toggle}
				class="inline-flex h-10 items-center gap-2.5 rounded-md border border-slate-300 bg-white px-3
				       text-sm whitespace-nowrap transition-all duration-150 ease-out
				       hover:border-slate-400 hover:bg-slate-50
				       focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none
				       {layout === 'block' ? 'w-full justify-center' : ''} {pressClasses}"
			>
				<Icon icon={icons.calendar} size="sm" class="shrink-0 text-slate-500" />
				<span class="font-medium text-slate-900">{presetLabel}</span>
				<span class="h-4 w-px shrink-0 bg-slate-300" aria-hidden="true"></span>
				<span class="text-slate-700 tabular-nums">{rangeLabel}</span>
			</button>
		{/snippet}

		<div class="w-[19rem] sm:w-auto">
			<!-- The presets moved inside the picker: one control, one place to change the period. -->
			<div class="flex flex-wrap items-center gap-1 border-b border-slate-200 pb-3">
				{#each presets as item (item.value)}
					<button
						type="button"
						aria-pressed={draftPreset === item.value}
						onclick={() => choosePreset(item.value, item.from, item.to)}
						class="h-8 rounded-md px-2.5 text-sm transition-colors
						       hover:bg-amber-100 focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none
						       aria-pressed:bg-amber-100 aria-pressed:font-semibold aria-pressed:text-slate-950
						       text-slate-700 {pressClasses}"
					>
						{item.label}
					</button>
				{/each}
				{#if draftPreset === 'custom'}
					<span class="ml-auto text-xs text-slate-500">Custom range</span>
				{/if}
			</div>

			<div class="flex items-center justify-between pt-3">
				<button
					type="button"
					aria-label="Previous month"
					onclick={() => (leftMonth = addMonthsUTC(leftMonth, -1))}
					class="inline-flex size-8 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none"
				>
					<Icon icon="heroicons:chevron-left" size="sm" />
				</button>
				<div class="flex flex-1 justify-around text-sm font-medium text-slate-800">
					<span>{monthLabel(leftMonth)}</span>
					<span class="hidden sm:inline">{monthLabel(rightMonth)}</span>
				</div>
				<button
					type="button"
					aria-label="Next month"
					onclick={() => (leftMonth = addMonthsUTC(leftMonth, 1))}
					class="inline-flex size-8 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none"
				>
					<Icon icon="heroicons:chevron-right" size="sm" />
				</button>
			</div>

			<div class="flex gap-4 pt-2">
				<Calendar
					month={leftMonth}
					mode="range"
					rangeStart={draftFrom}
					rangeEnd={draftTo}
					hoverDate={hover}
					showNav={false}
					onSelect={selectDay}
					onHover={(iso) => (hover = iso)}
				/>
				<Calendar
					month={rightMonth}
					mode="range"
					rangeStart={draftFrom}
					rangeEnd={draftTo}
					hoverDate={hover}
					showNav={false}
					onSelect={selectDay}
					onHover={(iso) => (hover = iso)}
					class="hidden sm:block"
				/>
			</div>

			<div class="mt-3 flex items-center justify-between border-t border-slate-200 pt-3">
				<span class="text-xs text-slate-500 tabular-nums">
					{draftFrom ? formatDisplayDate(draftFrom) : '—'}
					{draftTo ? `→ ${formatDisplayDate(draftTo)}` : ''}
				</span>
				<div class="flex gap-2">
					<Button
						size="sm"
						variant="ghost"
						onclick={() => (open = false)}
						class={pressClasses}
					>
						Cancel
					</Button>
					<Button size="sm" onclick={apply} disabled={!draftFrom} class={pressClasses}>
						Apply
					</Button>
				</div>
			</div>
		</div>
	</Popover>

	{#if showCaption}
		<Text as="p" size="xs" tone="subtle">
			Applies to the charts and the table on every page.
		</Text>
	{/if}
</div>
