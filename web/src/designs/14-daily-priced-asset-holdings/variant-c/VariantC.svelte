<script lang="ts">
	// Design #14 - Variant C "Holdings in the ledger".
	// The Assets page keeps its shape and gains a second view next to Classes: one account-wide
	// ledger of every daily-priced item, with aligned money columns. Picking a row retitles the
	// analytics region above it and opens that item's purchases in the ledger footer - master and
	// detail on one page, no overlay.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { chartColors } from '$lib/charts/theme';
	import {
		combinedSeries,
		dayTick,
		goldCoins,
		holdings,
		holdingsTotals,
		longDate,
		qty,
		quoteResults
	} from '../fixtures';
	import type { DesignHolding, DesignPurchase, DesignQuote } from '../fixtures';

	type Props = {
		/** Holding whose detail fills the analytics region and the ledger footer. */
		selectedId?: string | null;
		/** Open the "Add holding" drawer. */
		addOpen?: boolean;
		/** Render an account with no daily-priced items yet. */
		empty?: boolean;
		/** Render the loading skeletons. */
		loading?: boolean;
	};

	let {
		selectedId = $bindable('hld-btc'),
		addOpen = $bindable(false),
		empty = false,
		loading = false
	}: Props = $props();

	let tab = $state('holdings');
	let quoteQuery = $state('');
	let pickedQuote = $state<DesignQuote>(quoteResults[0]);
	let purchaseDate = $state('2026-05-20');
	let quantity = $state('0.05');
	let unitPrice = $state('60000.00');

	const rows = $derived(empty ? [] : holdings);
	const selected = $derived(empty ? null : (holdings.find((h) => h.id === selectedId) ?? null));
	const series = $derived(empty ? [] : selected ? selected.series : combinedSeries);

	// One ruled line rather than a KPI grid: the ledger row below already carries these figures per
	// item, so the summary only has to say which scope they describe.
	const summary = $derived({
		scope: empty
			? 'No daily-priced items yet'
			: selected
				? selected.name
				: 'All daily-priced items',
		paid: empty ? 0 : selected ? selected.paid : holdingsTotals.paid,
		value: empty ? 0 : selected ? selected.value : holdingsTotals.value,
		unrealized: empty ? 0 : selected ? selected.unrealized : holdingsTotals.unrealized,
		pct: empty ? 0 : selected ? selected.unrealized_pct : holdingsTotals.unrealized_pct
	});

	const paidPreview = $derived(
		(Number.parseFloat(quantity) || 0) * (Number.parseFloat(unitPrice) || 0)
	);

	function pick(row: DesignHolding) {
		selectedId = selectedId === row.id ? null : row.id;
	}
</script>

<!-- The marker column has no visible header; DataTable has no slot for a hidden one, so the
     header's filter slot carries the name for screen readers. -->
{#snippet markerHeader()}
	<span class="sr-only">Show purchases</span>
{/snippet}

{#snippet markerCell(row: DesignHolding)}
	<Icon
		icon={row.id === selectedId ? 'heroicons:chevron-down' : 'heroicons:chevron-right'}
		size="sm"
		class={row.id === selectedId ? 'text-slate-700' : 'text-slate-300'}
	/>
{/snippet}

{#snippet itemCell(row: DesignHolding)}
	<span class="block text-slate-900">{row.name}</span>
	<span class="mt-0.5 block text-xs text-slate-500">{row.symbol} · {row.class_name}</span>
{/snippet}

{#snippet quantityCell(row: DesignHolding)}
	<span class="tabular-nums">
		{qty(Number.parseFloat(row.quantity))} {row.symbol.split('/')[0]}
	</span>
{/snippet}

{#snippet priceCell(row: DesignHolding)}
	<Money amount={row.price} currency="EUR" size="sm" class="block" />
	<span class="mt-0.5 block text-xs text-slate-500">
		{longDate(row.price_date)}{row.price_carried_forward ? ' · carried forward' : ''}
	</span>
{/snippet}

{#snippet paidCell(row: DesignHolding)}
	<Money amount={row.paid} currency="EUR" size="sm" />
{/snippet}

{#snippet valueCell(row: DesignHolding)}
	<Money amount={row.value} currency="EUR" size="sm" weight="semibold" />
{/snippet}

{#snippet unrealizedCell(row: DesignHolding)}
	<Money
		amount={row.unrealized}
		currency="EUR"
		size="sm"
		colored
		signDisplay="exceptZero"
		class="block"
	/>
	<span class="mt-0.5 block">
		<TrendIndicator value={row.unrealized_pct} size="sm" showArrow={false} />
	</span>
{/snippet}

{#snippet trendCell(row: DesignHolding)}
	<Sparkline
		data={row.series.map((p) => p.value)}
		width={84}
		height={28}
		fill
		ariaLabel={`Value of ${row.name} since the first purchase`}
	/>
{/snippet}

{#snippet purchasesFooter()}
	{#if selected}
		<div class="max-h-48 overflow-y-auto bg-white px-4 py-3">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<h3 class="text-sm text-slate-900">
					Purchases of {selected.name}
					<span class="text-slate-500">
						· {selected.purchases.length} since {longDate(selected.purchases.at(-1)!.date)}
					</span>
				</h3>
				<Button size="sm" variant="outline" intent="secondary" shape="default">
					<Icon icon="heroicons:plus" size="sm" />Add purchase
				</Button>
			</div>
			<table class="mt-2 w-full text-sm tabular-nums">
				<thead>
					<tr class="border-b border-slate-300 text-xs text-slate-500">
						<th scope="col" class="py-1 text-left font-medium">Date</th>
						<th scope="col" class="py-1 text-right font-medium">Quantity</th>
						<th scope="col" class="py-1 text-right font-medium">Price paid</th>
						<th scope="col" class="py-1 text-right font-medium">Paid</th>
						<th scope="col" class="py-1 text-right font-medium">Value now</th>
					</tr>
				</thead>
				<tbody>
					{#each selected.purchases as purchase (purchase.id)}
						{@render purchaseRow(purchase)}
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
{/snippet}

{#snippet purchaseRow(purchase: DesignPurchase)}
	<tr class="border-b border-slate-200 last:border-b-0">
		<td class="py-1.5 text-slate-700">{longDate(purchase.date)}</td>
		<td class="py-1.5 text-right text-slate-700">
			{qty(Number.parseFloat(purchase.quantity))}
		</td>
		<td class="py-1.5 text-right">
			<Money amount={Number.parseFloat(purchase.unit_price)} currency="EUR" size="sm" />
		</td>
		<td class="py-1.5 text-right">
			<Money amount={Number.parseFloat(purchase.paid)} currency="EUR" size="sm" />
		</td>
		<td class="py-1.5 text-right">
			<Money
				amount={Number.parseFloat(purchase.quantity) * (selected?.price ?? 0)}
				currency="EUR"
				size="sm"
				weight="semibold"
			/>
		</td>
	</tr>
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
			<div class="flex flex-col gap-3">
				<div
					class="flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1 border-t border-slate-400 pt-2"
				>
					<p class="text-sm text-slate-700">{summary.scope}</p>
					<dl class="flex flex-wrap items-baseline gap-x-6 gap-y-1">
						<div class="flex items-baseline gap-2">
							<dt class="text-xs font-semibold tracking-wide text-slate-500 uppercase">Paid</dt>
							<dd><Money amount={summary.paid} currency="EUR" size="sm" weight="semibold" /></dd>
						</div>
						<div class="flex items-baseline gap-2">
							<dt class="text-xs font-semibold tracking-wide text-slate-500 uppercase">
								Value now
							</dt>
							<dd><Money amount={summary.value} currency="EUR" size="sm" weight="semibold" /></dd>
						</div>
						<div class="flex items-baseline gap-2">
							<dt class="text-xs font-semibold tracking-wide text-slate-500 uppercase">
								Unrealized
							</dt>
							<dd class="flex items-baseline gap-2">
								<Money
									amount={summary.unrealized}
									currency="EUR"
									size="sm"
									weight="semibold"
									colored
									signDisplay="exceptZero"
								/>
								<TrendIndicator value={summary.pct} size="sm" showArrow={false} />
							</dd>
						</div>
					</dl>
				</div>
				<AnalyticsCard
					title={selected
						? `${selected.name}: value against paid`
						: 'Daily-priced items: value against paid'}
				>
					<TimeSeriesChart
						height="h-40"
						{loading}
						labels={series.map((p) => p.date)}
						xTickFormat={dayTick}
						datasets={[
							{
								label: 'Value',
								data: series.map((p) => p.value),
								color: chartColors.positive,
								fill: true
							},
							{
								label: 'Paid',
								data: series.map((p) => p.paid),
								color: chartColors.net,
								dashed: true
							}
						]}
						ariaLabel="Value against what was paid"
					/>
				</AnalyticsCard>
			</div>
		{/snippet}

		<LedgerToolbar actionLabel="Add holding" onAdd={() => (addOpen = true)}>
			<Tabs
				tabs={[
					{ value: 'classes', label: 'Classes' },
					{ value: 'holdings', label: 'Holdings' }
				]}
				bind:value={tab}
				ariaLabel="Assets view"
			/>
			<Text as="span" size="sm" tone="muted">
				{rows.length} items follow a daily price · {goldCoins.name} stays manual
			</Text>
		</LedgerToolbar>

		<DataTable
			{rows}
			{loading}
			emptyText="No daily-priced items yet. Add a holding to follow its daily price."
			onRowClick={pick}
			footer={selected ? purchasesFooter : undefined}
			columns={[
				{
					key: 'marker',
					header: '',
					width: 'w-8',
					cell: markerCell,
					filter: markerHeader
				},
				{ key: 'item', header: 'Item', cell: itemCell },
				{ key: 'quantity', header: 'Quantity', align: 'right', cell: quantityCell },
				{ key: 'price', header: 'Daily price', align: 'right', cell: priceCell },
				{ key: 'paid', header: 'Paid', align: 'right', cell: paidCell },
				{ key: 'value', header: 'Value now', align: 'right', cell: valueCell },
				{ key: 'unrealized', header: 'Unrealized', align: 'right', cell: unrealizedCell },
				{ key: 'trend', header: 'Since first purchase', align: 'right', cell: trendCell }
			]}
		/>
	</PageContentTemplate>
</AppShellTemplate>

<Drawer bind:open={addOpen} title="Add holding" width="max-w-md">
	{#snippet header()}
		<Text as="p" size="sm" tone="muted">Its value follows the daily price in euro</Text>
	{/snippet}

	<div class="space-y-4">
		<div class="space-y-1.5">
			<p class="text-sm font-medium text-slate-700">What do you own?</p>
			<SearchInput
				bind:value={quoteQuery}
				placeholder="Filter bitcoin, ethereum, gold…"
				ariaLabel="Filter the daily-priced instruments"
			/>
			<Text as="p" size="sm" tone="muted">Only euro prices can be tracked.</Text>
		</div>

		<ul class="divide-y divide-slate-200 rounded-md border border-slate-300">
			{#each quoteResults.slice(0, 3) as quote (quote.id)}
				<li>
					<button
						type="button"
						aria-current={pickedQuote.id === quote.id ? 'true' : undefined}
						class={[
							'flex w-full items-center justify-between gap-3 px-3 py-2.5 text-left transition-colors',
							pickedQuote.id === quote.id ? 'bg-amber-50' : 'hover:bg-slate-50',
							'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none'
						].join(' ')}
						onclick={() => (pickedQuote = quote)}
					>
						<span class="flex min-w-0 items-center gap-2">
							{#if pickedQuote.id === quote.id}
								<Icon icon="heroicons:check" size="sm" class="shrink-0 text-slate-600" />
							{/if}
							<span class="text-sm font-medium text-slate-900">{quote.symbol}</span>
							<span class="truncate text-sm text-slate-500">{quote.name}</span>
						</span>
						<span class="shrink-0">
							<Money amount={quote.price} currency="EUR" size="sm" />
						</span>
					</button>
				</li>
			{/each}
			<li class="flex items-start gap-2 bg-taupe-50 px-3 py-2.5">
				<Icon icon="heroicons:information-circle" size="sm" class="mt-0.5 shrink-0 text-sky-700" />
				<span class="min-w-0">
					<span class="block text-sm text-slate-700">XAU · Gold (troy ounce)</span>
					<span class="block text-xs text-slate-500">
						Priced in US dollars. Track gold as a manual item for now.
					</span>
				</span>
			</li>
		</ul>

		<div class="border-t border-slate-300 pt-4">
			<p class="mb-2 text-xs font-semibold tracking-wide text-slate-500 uppercase">
				First purchase
			</p>
			<div class="space-y-3">
				<FormField label="Date" id="vc-date">
					<DatePicker
						bind:value={purchaseDate}
						max="2026-06-17"
						layer="filterPopover"
						ariaLabel="Purchase date"
					/>
				</FormField>
				<div class="grid grid-cols-2 gap-3">
					<FormField label="Quantity" id="vc-qty" hint={`In ${pickedQuote.symbol.split('/')[0]}`}>
						{#snippet children(ctx)}
							<Input
								id={ctx.id}
								bind:value={quantity}
								ariaDescribedby={ctx.describedby}
								class="text-right tabular-nums"
							/>
						{/snippet}
					</FormField>
					<FormField label="Price paid per unit" id="vc-price">
						{#snippet children(ctx)}
							<CurrencyInput id={ctx.id} bind:value={unitPrice} />
						{/snippet}
					</FormField>
				</div>
			</div>
			<div class="mt-3 flex items-baseline justify-between gap-3 border-t border-slate-200 pt-3">
				<Text as="span" size="sm" tone="muted">Paid</Text>
				<Money amount={paidPreview} currency="EUR" size="lg" weight="semibold" />
			</div>
		</div>

		<div class="flex items-start gap-2">
			<Badge intent="neutral" variant="soft" size="sm">Note</Badge>
			<Text as="p" size="sm" tone="muted">
				Prices are fetched back to this date. On days without a price the last known one is
				carried forward.
			</Text>
		</div>
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (addOpen = false)}>Cancel</Button>
		<Button intent="success">Save holding</Button>
	{/snippet}
</Drawer>
