<script lang="ts">
	// Design #14 - final, screen 1: Assets keeps its shape, and the class drawer it already opens
	// becomes the quick performance overview from variant A.
	//
	// Every daily-priced item opens its performance in place - the price it is valued at, paid
	// against value now and unrealized, and the value since the first purchase - so you can see how
	// a single item is doing without leaving Assets. Everything that needs room (the chart against
	// paid, the purchases) lives one step further, on the class page.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import HoldingPerformanceRow from './HoldingPerformanceRow.svelte';
	import AddItemDialog from './AddItemDialog.svelte';
	import { chartColors } from '$lib/charts/theme';
	import { dayTick, designClasses, goldCoins, holdings, longDate } from '../fixtures';
	import type { DesignClassRow } from '../fixtures';

	type Props = {
		/** Which item has its performance open; null collapses them all. */
		expandedId?: string | null;
		/** Open the "Add item" dialog over the drawer. */
		addOpen?: boolean;
	};

	let { expandedId = $bindable('hld-btc'), addOpen = $bindable(false) }: Props = $props();

	let drawerOpen = $state(true);
</script>

{#snippet worthCell(row: DesignClassRow)}
	<Money amount={row.worth} currency="EUR" size="sm" />
{/snippet}

{#snippet growthCell(row: DesignClassRow)}
	{#if row.growth_pct !== null}
		<Badge intent={row.growth_pct >= 0 ? 'success' : 'error'} variant="soft" size="sm">
			{row.growth_pct >= 0 ? '+' : ''}{row.growth_pct.toFixed(2)}%
		</Badge>
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Assets"
			showDateRange
			accountName="Design preview"
			accountEmail="preview@example.com"
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			<AnalyticsCard title="Total worth">
				<TimeSeriesChart
					height="h-40"
					labels={[
						'2026-01-15',
						'2026-02-10',
						'2026-03-02',
						'2026-04-18',
						'2026-05-20',
						'2026-06-17'
					]}
					xTickFormat={dayTick}
					datasets={[
						{
							label: 'Total worth',
							data: [48000, 52000, 54000, 55000, 57000, 58080],
							color: chartColors.positive,
							fill: true
						}
					]}
					ariaLabel="Total worth over time"
				/>
			</AnalyticsCard>
		{/snippet}

		<LedgerToolbar title="Asset classes" actionLabel="Add asset class" onAdd={() => {}} />
		<DataTable
			rows={designClasses}
			emptyText="No asset classes"
			columns={[
				{ key: 'name', header: 'Class', value: (r: DesignClassRow) => r.name },
				{ key: 'source', header: 'Source', value: (r: DesignClassRow) => r.source },
				{ key: 'worth', header: 'Current worth', align: 'right', cell: worthCell },
				{ key: 'growth', header: 'Growth', align: 'right', cell: growthCell }
			]}
		/>
	</PageContentTemplate>
</AppShellTemplate>

<Drawer bind:open={drawerOpen} title="Crypto & metals" width="max-w-lg">
	{#snippet header()}
		<Text as="p" size="sm" tone="muted">
			{holdings.length} of {holdings.length + 1} items follow a daily price
		</Text>
	{/snippet}

	<div class="space-y-6">
		<div class="flex items-end justify-between gap-3 border-t border-slate-400 pt-3">
			<div>
				<Text as="span" size="sm" tone="muted">Current worth</Text>
				<div>
					<Money amount={22400} currency="EUR" size="xl" weight="semibold" />
				</div>
			</div>
			<TrendIndicator value={14.2} />
		</div>

		<section>
			<h3 class="mb-2 text-sm text-slate-900">Items</h3>
			<ul class="rounded-xl border border-slate-200">
				{#each holdings as holding (holding.id)}
					<HoldingPerformanceRow
						{holding}
						mode="expand"
						open={expandedId === holding.id}
						onActivate={() => (expandedId = expandedId === holding.id ? null : holding.id)}
					/>
				{/each}
				<li class="flex items-center gap-3 px-3 py-2.5">
					<span class="w-4 shrink-0"></span>
					<span class="min-w-0 flex-1">
						<span class="flex flex-wrap items-center gap-2">
							<span class="truncate text-sm text-slate-900">{goldCoins.name}</span>
							<Badge intent="neutral" variant="soft" size="sm">Manual</Badge>
						</span>
						<span class="mt-0.5 block text-xs text-slate-500">
							Worth set on {longDate(goldCoins.last_set)}
						</span>
					</span>
					<span class="shrink-0 text-right">
						<Money amount={goldCoins.worth} currency="EUR" size="sm" weight="semibold" />
					</span>
				</li>
			</ul>
		</section>
	</div>

	{#snippet footer()}
		<Button variant="outline" intent="secondary" shape="default">
			Open class page<Icon icon="heroicons:arrow-right" size="sm" />
		</Button>
		<Button shape="default" onclick={() => (addOpen = true)}>
			<Icon icon="heroicons:plus" size="sm" />Add item
		</Button>
	{/snippet}
</Drawer>

<AddItemDialog bind:open={addOpen} />
