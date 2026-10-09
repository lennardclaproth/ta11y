<script lang="ts">
	// Design #14 - final, screen 2: the asset class as its own page (variant B) with the drawer's
	// performance line from variant A reused in its item ledger.
	//
	// Desktop: a narrow item ledger on the left, the selected item as an editorial article on the
	// right - headline, the price it is valued at, three figures, value against paid, and the
	// purchases that produced them.
	// Below lg there is no room for two columns, so the page lands on the ledger and every item
	// opens its performance in place, exactly like the class drawer; "Open <item>" follows through
	// to the article, and "All items" steps back.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Breadcrumb from '$lib/components/molecules/breadcrumb/Breadcrumb.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import HoldingPerformanceRow from './HoldingPerformanceRow.svelte';
	import AddItemDialog from './AddItemDialog.svelte';
	import { chartColors } from '$lib/charts/theme';
	import { dayTick, goldCoins, holdings, holdingsTotals, longDate, qty } from '../fixtures';
	import type { DesignPurchase } from '../fixtures';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';

	type Props = {
		/** Which holding fills the reading column. */
		selectedId?: string;
		/** `ledger` is the narrow-screen landing; `item` is the article. Desktop shows both. */
		view?: 'ledger' | 'item';
		/** Which holding has its performance open in the narrow ledger. */
		expandedId?: string | null;
		/** Open the two-step "Add item" dialog. */
		addOpen?: boolean;
		/** Which step of that dialog to show. */
		addStep?: 'quote' | 'purchase';
		/** Pre-filled filter in the instrument picker. */
		quoteQuery?: string;
		/** The purchase came back as a failed save. */
		saveError?: boolean;
		/** Render the loading skeletons instead of the data. */
		loading?: boolean;
		/** The class has no items yet. */
		empty?: boolean;
		/** The class could not be loaded at all. */
		error?: boolean;
		/** Today's price could not be fetched, so the figures are one day behind. */
		stale?: boolean;
	};

	let {
		selectedId = $bindable('hld-btc'),
		view = $bindable('item'),
		expandedId = $bindable('hld-btc'),
		addOpen = $bindable(false),
		addStep = $bindable('quote'),
		quoteQuery = '',
		saveError = false,
		loading = false,
		empty = false,
		error = false,
		stale = false
	}: Props = $props();

	const items = $derived(empty ? [] : holdings);
	const selected = $derived(holdings.find((h) => h.id === selectedId) ?? holdings[0]);
	const unit = $derived(selected.symbol.split('/')[0]);

	// An empty class is genuinely worth zero, so the two totals stay. "Unrealized" drops out:
	// there is nothing to be unrealized on, and a zero there would read as "no gain" instead.
	const classSummary = $derived(
		empty
			? [
					{ label: 'Paid in', amount: 0, colored: false },
					{ label: 'Worth', amount: 0, colored: false }
				]
			: [
					{ label: 'Paid in', amount: holdingsTotals.paid + goldCoins.worth, colored: false },
					{ label: 'Worth', amount: holdingsTotals.value + goldCoins.worth, colored: false },
					{ label: 'Unrealized', amount: holdingsTotals.unrealized, colored: true }
				]
	);

	const holdingKpis = $derived<KpiItem[]>([
		{ label: 'Paid', amount: selected.paid, currency: 'EUR' },
		{ label: 'Value now', amount: selected.value, currency: 'EUR' },
		{
			label: 'Unrealized',
			amount: selected.unrealized,
			currency: 'EUR',
			change: selected.unrealized_pct
		}
	]);

	function openItem(id: string) {
		selectedId = id;
		view = 'item';
	}
</script>

{#snippet quantityCell(row: DesignPurchase)}
	<span class="tabular-nums">{qty(Number.parseFloat(row.quantity))} {unit}</span>
{/snippet}

{#snippet unitPriceCell(row: DesignPurchase)}
	<Money amount={Number.parseFloat(row.unit_price)} currency="EUR" size="sm" />
{/snippet}

{#snippet paidCell(row: DesignPurchase)}
	<Money amount={Number.parseFloat(row.paid)} currency="EUR" size="sm" />
{/snippet}

{#snippet valueCell(row: DesignPurchase)}
	<Money amount={Number.parseFloat(row.quantity) * selected.price} currency="EUR" size="sm" />
{/snippet}

{#snippet manualRow()}
	<li class="flex items-center gap-3 border-b border-slate-200 px-3 py-2.5 last:border-b-0">
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
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Crypto & metals"
			showDateRange
			accountName="Design preview"
			accountEmail="preview@example.com"
		/>
	{/snippet}

	<div class="relative flex min-h-full flex-col gap-5 px-4 pb-6 lg:h-full lg:min-h-0 lg:px-8">
		<!-- The class line: where you are, and what the whole class is worth. The figures that matter
		     per item live in the reading column, so this stays one rule high. -->
		<div
			class="flex shrink-0 flex-wrap items-baseline justify-between gap-x-6 gap-y-2 border-b border-slate-300 pb-2"
		>
			<Breadcrumb items={[{ label: 'Assets', href: '/assets' }, { label: 'Crypto & metals' }]} />
			{#if loading}
				<!-- The class totals wait with the rest: a real figure next to skeletons reads as a
				     result that is already in. -->
				<Skeleton variant="rect" class="h-5 w-72" />
			{:else if !error}
				<dl class="flex flex-wrap items-baseline gap-x-6 gap-y-1">
					{#each classSummary as entry (entry.label)}
						<div class="flex items-baseline gap-2">
							<dt class="text-xs font-semibold tracking-wide text-slate-500 uppercase">
								{entry.label}
							</dt>
							<dd>
								<Money
									amount={entry.amount}
									currency="EUR"
									size="sm"
									weight="semibold"
									colored={entry.colored}
									signDisplay={entry.colored ? 'exceptZero' : 'auto'}
								/>
							</dd>
						</div>
					{/each}
					{#if !empty}
						<TrendIndicator value={holdingsTotals.unrealized_pct} size="sm" />
					{/if}
				</dl>
			{/if}
		</div>

		{#if stale}
			<!-- The figures stay on screen, so the warning says which price they are based on rather
			     than claiming nothing could be read. -->
			<Alert intent="warning" title="Prices are not up to date">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<span>
						Today's daily prices could not be fetched. The values below still use the price of
						{longDate(selected.price_date)}.
					</span>
					<Button size="sm" variant="outline" intent="secondary" shape="default">Try again</Button>
				</div>
			</Alert>
		{/if}

		{#if error}
			<!-- Nothing could be read, so the page does not pretend to show figures. -->
			<Alert intent="error" title="This asset class could not be loaded">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<span>
						The items and their daily prices are unavailable right now. Nothing was changed; your
						purchases are still recorded.
					</span>
					<Button size="sm" variant="outline" intent="secondary" shape="default">Try again</Button>
				</div>
			</Alert>
		{:else}
			<div class="flex min-h-0 flex-1 flex-col gap-5 lg:flex-row">
				<Panel
					variant="default"
					shape="square"
					padding="none"
					class={[
						'w-full shrink-0 flex-col overflow-hidden lg:flex lg:w-72 xl:w-80',
						view === 'ledger' ? 'flex' : 'hidden'
					].join(' ')}
				>
					<LedgerToolbar title="Items" actionLabel="Add item" onAdd={() => (addOpen = true)} />
					{#if loading}
						<div class="space-y-2 p-3" aria-busy="true">
							{#each [0, 1, 2] as i (i)}
								<Skeleton variant="rect" class="h-12 w-full" />
							{/each}
						</div>
					{:else if empty}
						<!-- The toolbar above already carries "Add item"; a second button here would make
						     one empty class look like it has two different ways in. -->
						<div class="px-4 py-10 text-center">
							<Text as="p" size="sm" tone="muted">
								No items in this class yet. Add one with a worth you set yourself, or link it to a
								daily price.
							</Text>
						</div>
					{:else}
						<ul class="min-h-0 flex-1 overflow-y-auto">
							{#each items as holding (holding.id)}
								<HoldingPerformanceRow
									{holding}
									mode={view === 'ledger' ? 'expand' : 'select'}
									open={expandedId === holding.id}
									current={selectedId === holding.id}
									showKind={view === 'ledger'}
									onActivate={() =>
										view === 'ledger'
											? (expandedId = expandedId === holding.id ? null : holding.id)
											: (selectedId = holding.id)}
									onOpenItem={() => openItem(holding.id)}
								/>
							{/each}
							{@render manualRow()}
						</ul>
					{/if}
				</Panel>

				<div
					class={[
						'min-w-0 flex-1 flex-col gap-4 lg:flex lg:min-h-0 lg:overflow-y-auto',
						view === 'ledger' ? 'hidden' : 'flex'
					].join(' ')}
				>
					<div class="lg:hidden">
						<Button
							size="sm"
							variant="ghost"
							intent="secondary"
							shape="default"
							onclick={() => (view = 'ledger')}
						>
							<Icon icon="heroicons:chevron-left" size="sm" />All items
						</Button>
					</div>

					{#if loading}
						<!-- The whole item loads as one: a figure that is already on screen while its chart is
						     still empty reads as a result, not as a wait. -->
						<div class="space-y-4" aria-busy="true">
							<Skeleton variant="rect" class="h-9 w-56" />
							<div class="grid gap-3 sm:grid-cols-3">
								<Skeleton variant="rect" class="h-16 w-full" />
								<Skeleton variant="rect" class="h-16 w-full" />
								<Skeleton variant="rect" class="h-16 w-full" />
							</div>
							<Skeleton variant="rect" class="h-44 w-full rounded-xl" />
							<Skeleton variant="rect" class="h-56 w-full rounded-none" />
						</div>
					{:else if empty}
						<div
							class="flex flex-1 flex-col items-center justify-center gap-3 border-t border-slate-400 px-6 py-16 text-center"
						>
							<Heading level="h2" size="lg" class="text-slate-900">Nothing to follow yet</Heading>
							<Text as="p" size="sm" tone="muted" class="max-w-sm">
								Link an item to a daily price and record what you bought. From that day on its value
								follows the price, back in time as well, and counts towards your net worth.
							</Text>
							<Button shape="default" onclick={() => (addOpen = true)}>
								<Icon icon="heroicons:plus" size="sm" />Add item
							</Button>
						</div>
					{:else}
						<header
							class="flex flex-wrap items-end justify-between gap-x-6 gap-y-2 border-t border-slate-400 pt-3"
						>
							<div class="min-w-0">
								<div class="flex flex-wrap items-center gap-2">
									<Heading level="h2" size="xl" class="text-slate-900">{selected.name}</Heading>
									<Badge intent="info" variant="soft">Daily price</Badge>
								</div>
								<Text as="p" size="sm" tone="muted">
									{selected.symbol} · {qty(Number.parseFloat(selected.quantity))}
									{unit} · average cost
									<Money amount={selected.avg_unit_price} currency="EUR" size="sm" /> per {unit}
								</Text>
							</div>
							<!-- The price block only hangs on the right once there is a column to hang it on;
							     stacked, a right-aligned label reads as a stray fragment. -->
							<div class="text-left sm:text-right">
								<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">
									Price per {unit}
								</p>
								<p class="leading-tight">
									<Money amount={selected.price} currency="EUR" size="xl" weight="semibold" />
								</p>
								<Text as="span" size="sm" tone="muted">
									{longDate(selected.price_date)}{selected.price_carried_forward
										? ' · last known price, carried forward'
										: ''}
								</Text>
							</div>
						</header>

						<KpiRow columns={3} items={holdingKpis} />

						<AnalyticsCard title="Value against paid">
							<!-- Two series need naming: the chart's own tooltip is hover-only, so the key and
							     the sentence under it carry the same reading without a pointer. -->
							<div
								class="mb-3 flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1 text-xs text-slate-600"
							>
								<span class="flex flex-wrap items-center gap-x-5 gap-y-1">
									<span class="flex items-center gap-2">
										<span class="inline-block h-0.5 w-6 bg-emerald-600"></span>Value
									</span>
									<span class="flex items-center gap-2">
										<span class="inline-block w-6 border-t-2 border-dashed border-slate-700"></span>
										Paid
									</span>
								</span>
								<span class="text-slate-500">
									Since {longDate(selected.series[0].date)} you paid
									<Money amount={selected.paid} currency="EUR" size="sm" /> and it is worth
									<Money amount={selected.value} currency="EUR" size="sm" /> today.
								</span>
							</div>
							<TimeSeriesChart
								height="h-44"
								labels={selected.series.map((p) => p.date)}
								xTickFormat={dayTick}
								datasets={[
									{
										label: 'Value',
										data: selected.series.map((p) => p.value),
										color: chartColors.positive,
										fill: true
									},
									{
										label: 'Paid',
										data: selected.series.map((p) => p.paid),
										color: chartColors.net,
										dashed: true
									}
								]}
								ariaLabel={`Value of ${selected.name} against what was paid for it`}
							/>
						</AnalyticsCard>

						<Panel
							variant="muted"
							shape="square"
							padding="none"
							class="flex min-h-56 flex-col overflow-hidden"
						>
							<LedgerToolbar
								title="Purchases"
								actionLabel="Add purchase"
								onAdd={() => {
									addStep = 'purchase';
									addOpen = true;
								}}
							/>
							<DataTable
								rows={selected.purchases}
								emptyText="No purchases yet"
								columns={[
									{
										key: 'date',
										header: 'Date',
										value: (r: DesignPurchase) => longDate(r.date)
									},
									{ key: 'quantity', header: 'Quantity', align: 'right', cell: quantityCell },
									{ key: 'unit', header: 'Price paid', align: 'right', cell: unitPriceCell },
									{ key: 'paid', header: 'Paid', align: 'right', cell: paidCell },
									{ key: 'value', header: 'Value now', align: 'right', cell: valueCell }
								]}
							/>
						</Panel>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</AppShellTemplate>

<AddItemDialog bind:open={addOpen} bind:step={addStep} {quoteQuery} {saveError} />
