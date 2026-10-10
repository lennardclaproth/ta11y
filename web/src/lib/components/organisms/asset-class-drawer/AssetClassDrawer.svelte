<script lang="ts">
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import HoldingPerformanceRow from '$lib/components/molecules/holding-performance-row/HoldingPerformanceRow.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { decimalStringToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { AssetClassDetails } from '$lib/api/types';

	type Props = {
		open?: boolean;
		details?: AssetClassDetails | null;
		loading?: boolean;
		error?: string | null;
		onClose?: () => void;
		/** Follow through to the class's own page. */
		onOpenClass?: () => void;
		/** Open the "Add item" dialog over the drawer. */
		onAddItem?: () => void;
		/** Follow through to one daily-priced item on the class page. */
		onOpenItem?: (assetId: string) => void;
	};

	let {
		open = $bindable(false),
		details = null,
		loading = false,
		error = null,
		onClose,
		onOpenClass,
		onAddItem,
		onOpenItem
	}: Props = $props();

	const growth = $derived(details?.class.growth_pct ?? null);
	const holdings = $derived(details?.holdings ?? []);
	const manual = $derived(details?.assets ?? []);
	const itemCount = $derived(holdings.length + manual.length);

	// Which daily-priced item has its performance open; only one at a time, so the drawer stays
	// a list you scan rather than a wall of figures.
	let expandedId = $state<string | null>(null);
	let shownClassId = $state<string | null>(null);

	// A different class is a different set of items, so a row left open on the previous one
	// would be a stale reading. Reloading the same class is not a different class: an item
	// expanded before adding a purchase to it must stay expanded afterwards.
	$effect(() => {
		const classId = details?.class.id ?? null;
		if (classId !== shownClassId) {
			shownClassId = classId;
			expandedId = null;
		}
	});
</script>

<Drawer bind:open title={details?.class.name ?? 'Asset class'} width="max-w-lg" {onClose}>
	{#snippet header()}
		{#if details && holdings.length > 0}
			<Text as="p" size="sm" tone="muted">
				{holdings.length} of {itemCount} items follow a daily price
			</Text>
		{/if}
	{/snippet}

	{#if loading}
		<div class="space-y-3">
			{#each [0, 1, 2, 3] as i (i)}
				<Skeleton variant="rect" class="h-10 w-full rounded-lg" />
			{/each}
		</div>
	{:else if error}
		<div class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
			{error}
		</div>
	{:else if details}
		<div class="space-y-6">
			<div class="flex items-end justify-between gap-3">
				<div>
					<p class="text-xs text-slate-500">Current worth</p>
					<Money
						amount={decimalStringToNumber(details.class.current_worth)}
						currency="EUR"
						size="xl"
						weight="semibold"
					/>
				</div>
				{#if growth !== null}
					<Badge intent={growth >= 0 ? 'success' : 'error'} variant="soft">
						{growth >= 0 ? '+' : ''}{growth.toFixed(2)}%
					</Badge>
				{/if}
			</div>

			<section>
				<h3 class="mb-2 text-sm text-slate-900">Items</h3>
				{#if itemCount === 0}
					<p class="text-sm text-slate-500">No items in this class</p>
				{:else}
					<ul class="rounded-xl border border-slate-200">
						{#each holdings as holding (holding.id)}
							<HoldingPerformanceRow
								{holding}
								mode="expand"
								open={expandedId === holding.id}
								onActivate={() => (expandedId = expandedId === holding.id ? null : holding.id)}
								onOpenItem={onOpenItem ? () => onOpenItem(holding.id) : undefined}
							/>
						{/each}
						{#each manual as asset (asset.id)}
							<li
								class="flex items-center gap-3 border-b border-slate-200 px-3 py-2.5 last:border-b-0"
							>
								{#if holdings.length > 0}
									<!-- Line the manual rows up with the holdings' disclosure arrow. -->
									<span class="w-4 shrink-0"></span>
								{/if}
								<span class="min-w-0 flex-1">
									<span class="flex flex-wrap items-center gap-2">
										<span class="truncate text-sm text-slate-900">{asset.name}</span>
										{#if holdings.length > 0}
											<Badge intent="neutral" variant="soft" size="sm">Manual</Badge>
										{/if}
									</span>
									<span class="mt-0.5 block text-xs text-slate-500">
										Worth set on {formatDisplayDate(asset.updated_at.slice(0, 10))}
									</span>
								</span>
								<span class="shrink-0 text-right">
									<Money
										amount={decimalStringToNumber(asset.current_worth)}
										currency="EUR"
										size="sm"
										weight="semibold"
									/>
								</span>
							</li>
						{/each}
					</ul>
				{/if}
			</section>

			<section>
				<h3 class="mb-2 text-sm text-slate-900">Recent changes</h3>
				{#if details.mutations.length === 0}
					<p class="text-sm text-slate-500">No recorded changes</p>
				{:else}
					<ul class="space-y-2">
						{#each details.mutations as mutation (mutation.id)}
							<li class="flex items-center justify-between gap-3 text-sm">
								<span class="min-w-0">
									<span class="text-slate-700 capitalize">{mutation.change_type}</span>
									<span class="ml-2 text-xs text-slate-500"
										>{formatDisplayDate(mutation.effective_date)}</span
									>
								</span>
								<Money
									amount={decimalStringToNumber(mutation.new_worth)}
									currency="EUR"
									size="sm"
								/>
							</li>
						{/each}
					</ul>
				{/if}
			</section>
		</div>
	{:else}
		<p class="text-sm text-slate-500">Select an asset class to see details.</p>
	{/if}

	{#snippet footer()}
		{#if details && !loading && !error}
			{#if onOpenClass}
				<Button variant="outline" intent="secondary" shape="default" onclick={onOpenClass}>
					Open class page<Icon icon="heroicons:arrow-right" size="sm" />
				</Button>
			{/if}
			{#if onAddItem}
				<Button shape="default" onclick={onAddItem}>
					<Icon icon="heroicons:plus" size="sm" />Add item
				</Button>
			{/if}
		{/if}
	{/snippet}
</Drawer>
