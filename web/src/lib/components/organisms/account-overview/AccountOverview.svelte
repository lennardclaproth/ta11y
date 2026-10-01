<script lang="ts">
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import AccountMenu from '$lib/components/molecules/account-menu/AccountMenu.svelte';
	import DateRangePicker from '$lib/components/molecules/date-range-picker/DateRangePicker.svelte';
	import {
		customPresetValue,
		type DateRangePreset
	} from '$lib/components/molecules/date-range-picker/date-range-picker.types';
	import type { PopoverApi } from '$lib/components/molecules/popover/popover.types';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { AccountOverviewLayout } from './account-overview.types';

	/**
	 * "Your overview": who is signed in, what the account is worth, how that moved over the
	 * selected period, and the one control that selects that period. It is the personal region
	 * of the masthead — nothing here acts on the page you are looking at, which is why the
	 * navigation bar no longer carries anything that does.
	 *
	 * Admin pages pass `showPeriod={false}`: they have no charts and no period, so the worth is
	 * labelled as the latest snapshot rather than compared against a range you cannot choose.
	 */
	type Props = {
		layout?: AccountOverviewLayout;
		accountName: string;
		accountEmail?: string;
		/** Latest total worth in major units, or null when the account has no snapshot at all. */
		netWorth?: number | null;
		/** Percentage change across the period, or null when the period holds no comparison. */
		changePct?: number | null;
		loading?: boolean;
		/** The snapshots could not be read. Not the same as having none. */
		failed?: boolean;
		showPeriod?: boolean;
		periodFrom?: string;
		periodTo?: string;
		/** Name of the chosen preset, or `custom` for a hand-picked range. */
		periodPreset?: string;
		periodPresets?: DateRangePreset[];
		onPeriodChange?: (range: { from: string; to: string; preset: string }) => void;
		onSignOut?: () => void;
		class?: string;
	};

	let {
		layout = 'inline',
		accountName,
		accountEmail,
		netWorth = null,
		changePct = null,
		loading = false,
		failed = false,
		showPeriod = true,
		periodFrom = '',
		periodTo = '',
		periodPreset = customPresetValue,
		periodPresets = [],
		onPeriodChange,
		onSignOut,
		class: className = ''
	}: Props = $props();

	let pickerOpen = $state(false);

	const presetLabel = $derived(
		periodPresets.find((item) => item.value === periodPreset)?.label ?? 'Custom'
	);
	const rangeLabel = $derived(
		periodFrom && periodTo
			? `${formatDisplayDate(periodFrom)} – ${formatDisplayDate(periodTo)}`
			: 'Select a period'
	);

	/**
	 * The situations stay visibly different, so a displayed zero never stands in for "we do not
	 * know": still loading, unreachable, no snapshots at all, a period without a comparison, and
	 * a real change. A failed read in particular must not read as an empty account.
	 */
	const worthCaption = $derived(
		failed
			? "Couldn't be loaded. Reload the page to try again."
			: netWorth === null
				? 'No snapshots yet'
				: !showPeriod
					? 'Latest snapshot'
					: changePct === null
						? 'No snapshot in this period, so no change is shown'
						: `Change over ${rangeLabel}`
	);

	function handlePeriodChange(range: { from: string | null; to: string | null; preset: string }) {
		if (!range.from || !range.to) return;
		onPeriodChange?.({ from: range.from, to: range.to, preset: range.preset });
	}
</script>

{#snippet netWorthBlock(align: 'start' | 'end')}
	<div class={['flex flex-col', align === 'end' ? 'items-end' : 'items-start'].join(' ')}>
		<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">Net worth</Text>

		{#if loading}
			<Skeleton class="mt-1 h-6 w-32" />
			<Text as="span" size="xs" tone="subtle">Loading…</Text>
		{:else if failed}
			<span class="text-lg font-semibold text-slate-500 tabular-nums">—</span>
			<Text as="span" size="xs" tone="danger">{worthCaption}</Text>
		{:else if netWorth === null}
			<span class="text-lg font-semibold text-slate-500 tabular-nums">—</span>
			<Text as="span" size="xs" tone="subtle">{worthCaption}</Text>
		{:else}
			<div class="flex items-baseline gap-2">
				<Money amount={netWorth} currency="EUR" size="lg" weight="semibold" />
				{#if showPeriod && changePct !== null}
					<TrendIndicator value={changePct} size="sm" />
				{/if}
			</div>
			<Text as="span" size="xs" tone="subtle">{worthCaption}</Text>
		{/if}
	</div>
{/snippet}

{#snippet periodCard(api: PopoverApi, full: boolean)}
	<!-- One white card holds the name of the period and the days it resolves to; clicking it
	     opens the picker those presets now live in, so there is one place to change the period. -->
	<button
		type="button"
		aria-expanded={api.open}
		aria-label="Period: {presetLabel}, {rangeLabel}. Opens the date picker."
		onclick={api.toggle}
		class={[
			'inline-flex h-10 items-center gap-2.5 rounded-md border border-slate-300 bg-white px-3',
			'text-sm whitespace-nowrap transition-all duration-150 ease-out',
			'hover:border-slate-400 hover:bg-slate-50',
			'focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none',
			'active:ring-2 active:ring-amber-300',
			full ? 'w-full justify-center' : ''
		]
			.filter(Boolean)
			.join(' ')}
	>
		<Icon icon="heroicons:calendar-days" size="sm" class="shrink-0 text-slate-500" />
		<span class="font-medium text-slate-900">{presetLabel}</span>
		<span class="h-4 w-px shrink-0 bg-slate-300" aria-hidden="true"></span>
		<span class="text-slate-700 tabular-nums">{rangeLabel}</span>
	</button>
{/snippet}

{#snippet inlineTrigger(api: PopoverApi)}
	{@render periodCard(api, false)}
{/snippet}

{#snippet blockTrigger(api: PopoverApi)}
	{@render periodCard(api, true)}
{/snippet}

{#snippet periodPicker(full: boolean)}
	<!-- The committed range comes back through `onChange`; the picker's own copies of it are
	     overwritten from above on the next render, so they are passed one way. -->
	<DateRangePicker
		bind:open={pickerOpen}
		from={periodFrom || null}
		to={periodTo || null}
		preset={periodPreset}
		presets={periodPresets}
		trigger={full ? blockTrigger : inlineTrigger}
		onChange={handlePeriodChange}
	/>
{/snippet}

{#if layout === 'inline'}
	<div
		aria-label="Your overview"
		class={['hidden items-center gap-6 border-l border-slate-300 pl-6 lg:flex', className]
			.filter(Boolean)
			.join(' ')}
	>
		{@render netWorthBlock('end')}
		{#if showPeriod}
			{@render periodPicker(false)}
		{/if}
		<AccountMenu
			name={accountName}
			email={accountEmail}
			items={[
				{
					label: 'Sign out',
					icon: 'heroicons:arrow-right-start-on-rectangle',
					intent: 'danger',
					onSelect: onSignOut
				}
			]}
		/>
	</div>
{:else}
	<Panel
		variant="floating"
		shape="sm"
		shadow="md"
		padding="none"
		class={['w-80 sm:w-96', className].filter(Boolean).join(' ')}
	>
		<div class="flex items-center gap-3 p-4">
			<Avatar name={accountName} size="md" />
			<div class="min-w-0">
				<p class="truncate text-sm font-medium text-slate-900">{accountName}</p>
				{#if accountEmail}
					<p class="truncate text-xs text-slate-500">{accountEmail}</p>
				{/if}
			</div>
		</div>

		<div class="border-t border-slate-300 p-4">
			{@render netWorthBlock('start')}
		</div>

		{#if showPeriod}
			<div class="border-t border-slate-300 p-4">
				<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">Period</Text>
				<div class="mt-2">
					{@render periodPicker(true)}
				</div>
				<Text as="p" size="xs" tone="subtle" class="mt-2">
					Applies to the charts and the table on every page.
				</Text>
			</div>
		{/if}

		<div class="border-t border-slate-300 p-4">
			<Button variant="ghost" intent="error" size="lg" shape="default" class="w-full" onclick={onSignOut}>
				<Icon icon="heroicons:arrow-right-start-on-rectangle" />
				Sign out
			</Button>
		</div>
	</Panel>
{/if}
