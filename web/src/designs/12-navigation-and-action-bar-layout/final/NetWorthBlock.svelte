<script lang="ts">
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import { netWorth, netWorthChangePct, period } from '../shared/mock-data';

	/**
	 * Net worth with its change over the selected period, from the existing snapshots.
	 * The four situations stay visibly different, so a displayed zero never stands in for
	 * "we do not know": loading, no snapshots at all, no snapshot inside the period, and a
	 * real change. Admin pages have no period, so they show the latest value without a change.
	 */
	type Change = 'value' | 'no-snapshot' | 'no-period' | 'none';

	type Props = {
		align?: 'end' | 'start';
		loading?: boolean;
		change?: Change;
	};

	let { align = 'end', loading = false, change = 'value' }: Props = $props();

	const caption = {
		value: `Change over ${period.shortLabel}`,
		'no-snapshot': 'No snapshot in this period, so no change is shown',
		'no-period': 'Latest snapshot',
		none: 'No snapshots yet'
	} satisfies Record<Change, string>;
</script>

<div class={['flex flex-col', align === 'end' ? 'items-end' : 'items-start'].join(' ')}>
	<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">Net worth</Text>

	{#if loading}
		<Skeleton class="mt-1 h-6 w-32" />
		<Text as="span" size="xs" tone="subtle">Loading…</Text>
	{:else if change === 'none'}
		<span class="text-lg font-semibold tabular-nums text-slate-500">—</span>
		<Text as="span" size="xs" tone="subtle">{caption.none}</Text>
	{:else}
		<div class="flex items-baseline gap-2">
			<Money amount={netWorth} currency="EUR" size="lg" weight="semibold" />
			{#if change === 'value'}
				<TrendIndicator value={netWorthChangePct} size="sm" />
			{/if}
		</div>
		<Text as="span" size="xs" tone="subtle">{caption[change]}</Text>
	{/if}
</div>
