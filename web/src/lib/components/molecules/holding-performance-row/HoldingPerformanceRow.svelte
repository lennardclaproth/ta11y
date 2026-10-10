<script lang="ts">
	// One daily-priced item as a ruled ledger line: name, how much is held, what it is worth now
	// and which way it moved. In `expand` mode the line opens its performance in place — the price
	// it is valued at, paid against value now and unrealized, and the value since the first
	// purchase. The class page reuses the same line in `select` mode, where the reading column
	// carries the detail instead.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import { decimalStringToNumber, formatQuantity } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { AssetHolding } from '$lib/api/types';
	import type { HoldingRowMode } from './holding-performance-row.types';

	type Props = {
		holding: AssetHolding;
		mode?: HoldingRowMode;
		/** Only meaningful in `expand` mode. */
		open?: boolean;
		/** Only meaningful in `select` mode. */
		current?: boolean;
		/** Show the "Daily price" label; off where the whole list is daily-priced already. */
		showKind?: boolean;
		onActivate?: () => void;
		/** Follow through to the item's own reading view. */
		onOpenItem?: () => void;
	};

	let {
		holding,
		mode = 'expand',
		open = false,
		current = false,
		showKind = true,
		onActivate,
		onOpenItem
	}: Props = $props();

	const unit = $derived(holding.symbol.split('/')[0] || holding.name);
	const expanded = $derived(mode === 'expand' && open);
	const value = $derived(decimalStringToNumber(holding.value));
	const series = $derived(holding.series ?? []);
</script>

{#snippet figure(label: string, amount: number, colored = false)}
	<div class="min-w-0 border-t border-slate-400 pt-2">
		<Text as="span" size="xs" tone="muted">{label}</Text>
		<div class="mt-0.5">
			<Money
				{amount}
				currency="EUR"
				size="lg"
				weight="semibold"
				{colored}
				signDisplay={colored ? 'exceptZero' : 'auto'}
			/>
		</div>
	</div>
{/snippet}

<li class="border-b border-slate-200 last:border-b-0">
	<button
		type="button"
		aria-expanded={mode === 'expand' ? expanded : undefined}
		aria-current={mode === 'select' && current ? 'true' : undefined}
		class={[
			'flex w-full items-center gap-3 px-3 py-2.5 text-left transition-colors',
			current ? 'bg-amber-50' : 'hover:bg-slate-50',
			'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none'
		].join(' ')}
		onclick={() => onActivate?.()}
	>
		{#if mode === 'expand'}
			<Icon
				icon={expanded ? 'heroicons:chevron-down' : 'heroicons:chevron-right'}
				size="sm"
				class="shrink-0 text-slate-500"
			/>
		{/if}
		<span class="min-w-0 flex-1">
			<span class="flex flex-wrap items-center gap-2">
				<span class="truncate text-sm text-slate-900">{holding.name}</span>
				{#if showKind}
					<Badge intent="info" variant="soft" size="sm">Daily price</Badge>
				{/if}
			</span>
			<span class="mt-0.5 block text-xs text-slate-500 tabular-nums">
				{formatQuantity(holding.quantity)}
				{unit}{holding.symbol ? ` · ${holding.symbol}` : ''}
			</span>
		</span>
		<span class="shrink-0 text-right">
			<Money amount={value} currency="EUR" size="sm" weight="semibold" />
			{#if holding.unrealized_pct !== null && holding.unrealized_pct !== undefined}
				<span class="mt-0.5 block">
					<TrendIndicator value={holding.unrealized_pct} size="sm" />
				</span>
			{/if}
		</span>
	</button>

	{#if expanded}
		<div class="border-t border-slate-200 bg-taupe-50 px-3 py-4">
			<div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
				<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">
					Price per {unit}
				</p>
				<p class="text-sm text-slate-700">
					<Money amount={decimalStringToNumber(holding.price)} currency="EUR" size="sm" />
					{#if holding.price_date}
						· {formatDisplayDate(holding.price_date)}{holding.price_carried_forward
							? ' · last known price, carried forward'
							: ''}
					{:else}
						· no price yet
					{/if}
				</p>
			</div>

			<div class="mt-3 grid grid-cols-3 gap-3">
				{@render figure('Paid', decimalStringToNumber(holding.paid))}
				{@render figure('Value now', value)}
				{@render figure('Unrealized', decimalStringToNumber(holding.unrealized), true)}
			</div>

			{#if series.length > 0}
				<div class="mt-3 flex items-center justify-between gap-3 border-t border-slate-200 pt-3">
					<Text as="span" size="xs" tone="muted">
						Value since {formatDisplayDate(series[0].date)}
					</Text>
					<Sparkline
						data={series.map((p) => decimalStringToNumber(p.value))}
						width={160}
						height={36}
						fill
						ariaLabel={`Value of ${holding.name} since the first purchase`}
					/>
				</div>
			{/if}

			{#if onOpenItem}
				<div class="mt-3">
					<Button size="sm" variant="outline" intent="secondary" shape="default" onclick={onOpenItem}>
						Open {holding.name}<Icon icon="heroicons:arrow-right" size="sm" />
					</Button>
				</div>
			{/if}
		</div>
	{/if}
</li>
