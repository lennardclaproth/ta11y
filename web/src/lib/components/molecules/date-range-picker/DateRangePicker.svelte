<script lang="ts">
	import type { Snippet } from 'svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Popover from '$lib/components/molecules/popover/Popover.svelte';
	import Calendar from '$lib/components/molecules/calendar/Calendar.svelte';
	import type { PopoverApi } from '$lib/components/molecules/popover/popover.types';
	import {
		addDaysISO,
		addMonthsUTC,
		formatDisplayDate,
		monthLabel,
		parseISODate,
		startOfMonthUTC,
		toISODate,
		todayISO
	} from '$lib/components/molecules/calendar/calendar.utils';
	import {
		customPresetValue,
		type DateRangePickerSize,
		type DateRangePreset
	} from './date-range-picker.types';

	type Props = {
		/** Bindable range endpoints ("YYYY-MM-DD") or null. */
		from?: string | null;
		to?: string | null;
		placeholder?: string;
		min?: string | null;
		max?: string | null;
		size?: DateRangePickerSize;
		disabled?: boolean;
		/** Show the quick-range preset column. */
		showPresets?: boolean;
		/**
		 * Offer "Clear". Turn it off where an empty range is not a state the caller can hold —
		 * an app-wide period has no "no period", and clearing it would only leave the picker and
		 * its owner disagreeing.
		 */
		showClear?: boolean;
		/**
		 * Named ranges to offer instead of the built-in list. Supplied presets read as a row
		 * above the calendars rather than a column beside them, so one control holds both the
		 * named period and the exact days it resolves to.
		 */
		presets?: DateRangePreset[];
		/** Bindable name of the chosen preset; becomes `custom` once days are picked by hand. */
		preset?: string;
		/** Bindable open state, so a caller can open the picker from its own control. */
		open?: boolean;
		ariaLabel?: string;
		onChange?: (range: { from: string | null; to: string | null; preset: string }) => void;
		/** Replaces the default trigger button. Wire `aria-expanded={api.open}` on your control. */
		trigger?: Snippet<[PopoverApi]>;
		class?: string;
	};

	let {
		from = $bindable(null),
		to = $bindable(null),
		placeholder = 'Select range',
		min = null,
		max = null,
		size = 'md',
		disabled = false,
		showPresets = true,
		showClear = true,
		presets,
		preset = $bindable(customPresetValue),
		open = $bindable(false),
		ariaLabel = 'Select date range',
		onChange,
		trigger,
		class: className = ''
	}: Props = $props();

	let leftMonth = $state(startOfMonthUTC(parseISODate(from) ?? new Date()));
	let draftStart = $state<string | null>(from);
	let draftEnd = $state<string | null>(to);
	let draftPreset = $state(preset);
	let hover = $state<string | null>(null);

	const rightMonth = $derived(addMonthsUTC(leftMonth, 1));

	// Reset the draft to the committed range each time the popover opens.
	$effect(() => {
		if (open) {
			draftStart = from;
			draftEnd = to;
			draftPreset = preset;
			hover = null;
			leftMonth = startOfMonthUTC(parseISODate(from) ?? new Date());
		}
	});

	function handleSelect(iso: string) {
		// A hand-picked day makes the range custom again, whatever it was named before.
		draftPreset = customPresetValue;
		if (!draftStart || (draftStart && draftEnd)) {
			draftStart = iso;
			draftEnd = null;
			return;
		}
		// Second pick: order the endpoints.
		if (iso < draftStart) {
			draftEnd = draftStart;
			draftStart = iso;
		} else {
			draftEnd = iso;
		}
	}

	function shift(delta: number) {
		leftMonth = addMonthsUTC(leftMonth, delta);
	}

	function apply() {
		const start = draftStart;
		const end = draftEnd ?? draftStart;
		from = start;
		to = end;
		preset = draftPreset;
		onChange?.({ from, to, preset });
		open = false;
	}

	function clear() {
		draftStart = null;
		draftEnd = null;
		draftPreset = customPresetValue;
		from = null;
		to = null;
		preset = customPresetValue;
		onChange?.({ from: null, to: null, preset });
		open = false;
	}

	const sizeClasses = {
		sm: 'h-8 px-2.5 text-sm',
		md: 'h-10 px-3 text-sm',
		lg: 'h-12 px-3 text-base'
	} satisfies Record<DateRangePickerSize, string>;

	const label = $derived(
		from && to
			? `${formatDisplayDate(from)} – ${formatDisplayDate(to)}`
			: from
				? formatDisplayDate(from)
				: placeholder
	);

	const builtInPresets = $derived.by<DateRangePreset[]>(() => {
		const today = todayISO();
		const monthStart = toISODate(startOfMonthUTC(parseISODate(today) ?? new Date()));
		const yearStart = `${today.slice(0, 4)}-01-01`;
		return [
			{ value: '7d', label: 'Last 7 days', from: addDaysISO(today, -6), to: today },
			{ value: '30d', label: 'Last 30 days', from: addDaysISO(today, -29), to: today },
			{ value: 'month', label: 'This month', from: monthStart, to: today },
			{
				value: '3m',
				label: 'Last 3 months',
				from: toISODate(addMonthsUTC(parseISODate(today) ?? new Date(), -3)),
				to: today
			},
			{ value: 'ytd', label: 'Year to date', from: yearStart, to: today }
		];
	});

	const presetList = $derived(presets ?? builtInPresets);
	/** Supplied presets read across the top; the built-in list keeps its column on the left. */
	const presetsInline = $derived(presets !== undefined);

	function applyPreset(item: DateRangePreset) {
		draftPreset = item.value;
		draftStart = item.from;
		draftEnd = item.to;
		leftMonth = startOfMonthUTC(parseISODate(item.from) ?? new Date());
	}
</script>

{#snippet defaultTrigger(api: PopoverApi)}
	<button
		type="button"
		{disabled}
		aria-label={ariaLabel}
		aria-expanded={api.open}
		class={[
			'inline-flex items-center gap-2 rounded-xl border bg-white whitespace-nowrap',
			'transition-colors hover:bg-slate-50',
			'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
			'disabled:pointer-events-none disabled:opacity-50',
			api.open ? 'border-slate-400' : 'border-slate-300',
			from ? 'text-slate-800' : 'text-slate-500',
			sizeClasses[size],
			className
		].join(' ')}
		onclick={api.toggle}
	>
		<Icon icon="heroicons:calendar-days" size="sm" class="text-slate-500" />
		<span>{label}</span>
	</button>
{/snippet}

{#snippet presetButtons(layout: 'row' | 'column')}
	{#each presetList as item (item.value)}
		<button
			type="button"
			aria-pressed={draftPreset === item.value}
			class={[
				'rounded-lg text-sm text-slate-600 transition-colors',
				'hover:bg-amber-100 hover:text-slate-900',
				'focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none',
				'aria-pressed:bg-amber-100 aria-pressed:font-semibold aria-pressed:text-slate-950',
				layout === 'row' ? 'h-8 px-2.5' : 'px-2 py-1.5 text-left'
			].join(' ')}
			onclick={() => applyPreset(item)}
		>
			{item.label}
		</button>
	{/each}
{/snippet}

<Popover bind:open placement="bottom-start" class="p-3" trigger={trigger ?? defaultTrigger}>
	<div class="flex gap-3">
		{#if showPresets && !presetsInline}
			<div class="flex w-36 flex-col gap-1 border-r border-slate-200 pr-3">
				{@render presetButtons('column')}
			</div>
		{/if}

		<div class="flex flex-col gap-3">
			{#if showPresets && presetsInline}
				<div class="flex flex-wrap items-center gap-1 border-b border-slate-200 pb-3">
					{@render presetButtons('row')}
					{#if draftPreset === customPresetValue}
						<span class="ml-auto text-xs text-slate-500">Custom range</span>
					{/if}
				</div>
			{/if}

			<div class="flex items-center justify-between">
				<button
					type="button"
					aria-label="Previous month"
					class="inline-flex size-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
					onclick={() => shift(-1)}
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
					class="inline-flex size-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
					onclick={() => shift(1)}
				>
					<Icon icon="heroicons:chevron-right" size="sm" />
				</button>
			</div>

			<div class="flex gap-4">
				<Calendar
					month={leftMonth}
					mode="range"
					rangeStart={draftStart}
					rangeEnd={draftEnd}
					hoverDate={hover}
					{min}
					{max}
					showNav={false}
					onSelect={handleSelect}
					onHover={(iso) => (hover = iso)}
				/>
				<Calendar
					month={rightMonth}
					mode="range"
					rangeStart={draftStart}
					rangeEnd={draftEnd}
					hoverDate={hover}
					{min}
					{max}
					showNav={false}
					onSelect={handleSelect}
					onHover={(iso) => (hover = iso)}
					class="hidden sm:block"
				/>
			</div>

			<div class="flex items-center justify-between border-t border-slate-200 pt-3">
				<span class="text-xs text-slate-500 tabular-nums">
					{draftStart ? formatDisplayDate(draftStart) : '—'}
					{draftEnd ? `→ ${formatDisplayDate(draftEnd)}` : ''}
				</span>
				<div class="flex gap-2">
					{#if showClear}
						<Button size="sm" variant="ghost" intent="secondary" onclick={clear}>Clear</Button>
					{/if}
					<Button size="sm" onclick={apply} disabled={!draftStart}>Apply</Button>
				</div>
			</div>
		</div>
	</div>
</Popover>
