<script lang="ts">
	// Design #14 - Variant B "Class page, reading layout".
	// The asset class gets its own page: a narrow item ledger on the left and the selected holding
	// as an editorial article on the right - headline, price line, figures, value against paid, and
	// the purchases that produced them. Below lg the article takes the page and the ledger is one
	// step back.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Breadcrumb from '$lib/components/molecules/breadcrumb/Breadcrumb.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { chartColors } from '$lib/charts/theme';
	import {
		dayTick,
		goldCoins,
		holdings,
		holdingsTotals,
		longDate,
		qty,
		quoteResults
	} from '../fixtures';
	import type { DesignHolding, DesignPurchase, DesignQuote } from '../fixtures';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';

	type Props = {
		/** Which holding fills the reading column. */
		selectedId?: string;
		/** Open the two-step "Add item" dialog. */
		addOpen?: boolean;
		/** Which step of that dialog to show. */
		addStep?: 'quote' | 'purchase';
		/** Render the loading skeletons instead of the data. */
		loading?: boolean;
		/** Warn that today's price could not be fetched, so the figures are one day behind. */
		stale?: boolean;
	};

	let {
		selectedId = $bindable('hld-btc'),
		addOpen = $bindable(false),
		addStep = $bindable('quote'),
		loading = false,
		stale = false
	}: Props = $props();

	let quoteQuery = $state('');
	let pickedQuote = $state<DesignQuote>(quoteResults[0]);
	let purchaseDate = $state('2026-05-20');
	let quantity = $state('0.05');
	let unitPrice = $state('60000.00');

	const selected = $derived(holdings.find((h) => h.id === selectedId) ?? holdings[0]);
	const unit = $derived(selected.symbol.split('/')[0]);

	const classSummary = [
		{ label: 'Paid in', amount: holdingsTotals.paid + goldCoins.worth, colored: false },
		{ label: 'Worth', amount: holdingsTotals.value + goldCoins.worth, colored: false },
		{ label: 'Unrealized', amount: holdingsTotals.unrealized, colored: true }
	];

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

	const paidPreview = $derived(
		(Number.parseFloat(quantity) || 0) * (Number.parseFloat(unitPrice) || 0)
	);
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

{#snippet itemRow(holding: DesignHolding)}
	{@const current = holding.id === selected.id}
	<li>
		<button
			type="button"
			aria-current={current ? 'true' : undefined}
			class={[
				'flex w-full items-center gap-3 border-b border-slate-200 px-4 py-3 text-left transition-colors',
				current ? 'bg-amber-50' : 'hover:bg-slate-50',
				'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none'
			].join(' ')}
			onclick={() => (selectedId = holding.id)}
		>
			<span class="min-w-0 flex-1">
				<span class="block truncate text-sm text-slate-900">{holding.name}</span>
				<span class="mt-0.5 block text-xs text-slate-500 tabular-nums">
					{qty(Number.parseFloat(holding.quantity))}
					{holding.symbol.split('/')[0]} · daily price
				</span>
			</span>
			<span class="shrink-0 text-right">
				<Money amount={holding.value} currency="EUR" size="sm" weight="semibold" />
				<span class="mt-0.5 block">
					<TrendIndicator value={holding.unrealized_pct} size="sm" />
				</span>
			</span>
		</button>
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
				<TrendIndicator value={holdingsTotals.unrealized_pct} size="sm" />
			</dl>
		</div>

		{#if stale}
			<!-- The figures are still on screen, so the warning says which price they are based on
			     rather than claiming nothing could be read. -->
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

		<div class="flex min-h-0 flex-1 flex-col gap-5 lg:flex-row">
			<Panel
				variant="default"
				shape="square"
				padding="none"
				class="hidden w-72 shrink-0 flex-col overflow-hidden lg:flex xl:w-80"
			>
				<LedgerToolbar title="Items" actionLabel="Add item" onAdd={() => (addOpen = true)} />
				<ul class="min-h-0 flex-1 overflow-y-auto">
					{#each holdings as holding (holding.id)}
						{@render itemRow(holding)}
					{/each}
					<li class="flex items-center gap-3 border-b border-slate-200 px-4 py-3">
						<span class="min-w-0 flex-1">
							<span class="block truncate text-sm text-slate-900">{goldCoins.name}</span>
							<span class="mt-0.5 block text-xs text-slate-500">
								Manual · set {longDate(goldCoins.last_set)}
							</span>
						</span>
						<span class="shrink-0 text-right">
							<Money amount={goldCoins.worth} currency="EUR" size="sm" weight="semibold" />
						</span>
					</li>
				</ul>
			</Panel>

			<div class="flex min-w-0 flex-1 flex-col gap-4 lg:min-h-0 lg:overflow-y-auto">
				<div class="lg:hidden">
					<Button size="sm" variant="ghost" intent="secondary" shape="default">
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
					<div class="text-right">
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
					<TimeSeriesChart
						height="h-44"
						{loading}
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
					<LedgerToolbar title="Purchases" actionLabel="Add purchase" onAdd={() => {}} />
					<DataTable
						rows={selected.purchases}
						{loading}
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
	</div>
</AppShellTemplate>

<Dialog bind:open={addOpen} title="Add item to Crypto & metals" size="md">
	<div class="space-y-4">
		{#if addStep === 'quote'}
			<div>
				<p class="mb-2 text-xs font-semibold tracking-wide text-slate-500 uppercase">
					Step 1 of 2 · What do you own?
				</p>
				<SearchInput
					bind:value={quoteQuery}
					placeholder="Filter bitcoin, ethereum, gold…"
					ariaLabel="Filter the daily-priced instruments"
				/>
			</div>

			<Text as="p" size="sm" tone="muted">
				Instruments with a daily price. Only euro prices can be tracked; the rest stay manual.
			</Text>

			<ul class="divide-y divide-slate-200 rounded-md border border-slate-300">
				{#each quoteResults as quote (quote.id)}
					<li>
						<button
							type="button"
							disabled={!quote.selectable}
							aria-current={pickedQuote.id === quote.id ? 'true' : undefined}
							class={[
								'flex w-full items-start justify-between gap-3 px-3 py-2.5 text-left transition-colors',
								quote.selectable ? 'hover:bg-slate-50' : 'cursor-not-allowed bg-taupe-50',
								pickedQuote.id === quote.id ? 'bg-amber-50' : '',
								'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none'
							]
								.filter(Boolean)
								.join(' ')}
							onclick={() => (pickedQuote = quote)}
						>
							<span class="min-w-0">
								<span class="flex flex-wrap items-center gap-2">
									<span class="text-sm font-medium text-slate-900">{quote.symbol}</span>
									<span class="truncate text-sm text-slate-600">{quote.name}</span>
									<Badge
										intent={quote.selectable ? 'info' : 'neutral'}
										variant="soft"
										size="sm"
									>
										{quote.currency}
									</Badge>
								</span>
								{#if quote.reason}
									<span class="mt-0.5 block text-xs text-slate-500">{quote.reason}</span>
								{/if}
							</span>
							<span class="shrink-0 text-right">
								<Money amount={quote.price} currency={quote.currency} size="sm" />
								<span class="mt-0.5 block text-xs text-slate-500">
									{longDate(quote.price_date)}
								</span>
							</span>
						</button>
					</li>
				{/each}
			</ul>
		{:else}
			<div
				class="-mx-5 -mt-4 mb-1 flex flex-wrap items-end justify-between gap-x-4 gap-y-2 border-b border-slate-300 bg-taupe-50 px-5 py-3"
			>
				<div class="min-w-0">
					<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">
						Step 2 of 2 · First purchase
					</p>
					<p class="font-heading text-2xl leading-tight text-slate-900">{pickedQuote.name}</p>
					<Text as="span" size="sm" tone="muted">
						{pickedQuote.symbol} ·
						<Money amount={pickedQuote.price} currency="EUR" size="sm" />
						on {longDate(pickedQuote.price_date)}
					</Text>
				</div>
				<Button size="sm" variant="outline" intent="secondary" shape="default">Change</Button>
			</div>

			<div class="grid gap-3 sm:grid-cols-3">
				<FormField label="Date" id="vb-date">
					<DatePicker
						bind:value={purchaseDate}
						max="2026-06-17"
						portal={false}
						ariaLabel="Purchase date"
					/>
				</FormField>
				<FormField label="Quantity" id="vb-qty" hint={`In ${pickedQuote.symbol.split('/')[0]}`}>
					{#snippet children(ctx)}
						<Input
							id={ctx.id}
							bind:value={quantity}
							ariaDescribedby={ctx.describedby}
							class="text-right tabular-nums"
						/>
					{/snippet}
				</FormField>
				<FormField label="Price paid per unit" id="vb-price">
					{#snippet children(ctx)}
						<CurrencyInput id={ctx.id} bind:value={unitPrice} />
					{/snippet}
				</FormField>
			</div>

			<div class="flex items-baseline justify-between gap-3 border-t border-slate-300 pt-3">
				<Text as="span" size="sm" tone="muted">Paid</Text>
				<Money amount={paidPreview} currency="EUR" size="lg" weight="semibold" />
			</div>
			<Text as="p" size="sm" tone="muted">
				From this day on the value follows the daily price. You can add more purchases later.
			</Text>
		{/if}
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (addOpen = false)}>Cancel</Button>
		{#if addStep === 'quote'}
			<Button onclick={() => (addStep = 'purchase')}>Next</Button>
		{:else}
			<Button intent="success">Save</Button>
		{/if}
	{/snippet}
</Dialog>
