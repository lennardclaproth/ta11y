<script lang="ts">
	// Design #14 - Variant A "Expandable drawer".
	// Everything happens in the asset-class drawer that Assets already opens: items are one ruled
	// ledger, and a daily-priced item expands in place to show paid against value now, its value
	// trend, and its purchases. No new page, no second overlay.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { dayTick, designClasses, goldCoins, holdings, longDate, qty } from '../fixtures';
	import type { DesignClassRow, DesignHolding } from '../fixtures';

	type Props = {
		/** Holding expanded in the drawer; null collapses them all. */
		expandedId?: string | null;
		/** Open the "Add item" dialog over the drawer. */
		addOpen?: boolean;
		/** Drop the items so the empty class reads. */
		empty?: boolean;
	};

	let {
		expandedId = $bindable('hld-btc'),
		addOpen = $bindable(false),
		empty = false
	}: Props = $props();

	let drawerOpen = $state(true);
	let itemKind = $state('priced');
	let quantity = $state('0.05');
	let unitPrice = $state('60000.00');
	let purchaseDate = $state('2026-05-20');

	const items = $derived(empty ? [] : holdings);
	const classWorth = $derived(empty ? 0 : 22400);

	function toggle(id: string) {
		expandedId = expandedId === id ? null : id;
	}

	const paidPreview = $derived(
		(Number.parseFloat(quantity) || 0) * (Number.parseFloat(unitPrice) || 0)
	);
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

{#snippet holdingRow(holding: DesignHolding)}
	{@const open = expandedId === holding.id}
	<li class="border-b border-slate-200 last:border-b-0">
		<button
			type="button"
			aria-expanded={open}
			class="flex w-full items-center gap-3 px-3 py-2.5 text-left transition-colors hover:bg-slate-50 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
			onclick={() => toggle(holding.id)}
		>
			<Icon
				icon={open ? 'heroicons:chevron-down' : 'heroicons:chevron-right'}
				size="sm"
				class="shrink-0 text-slate-500"
			/>
			<span class="min-w-0 flex-1">
				<span class="flex items-center gap-2">
					<span class="truncate text-sm text-slate-900">{holding.name}</span>
					<Badge intent="info" variant="soft" size="sm">Daily price</Badge>
				</span>
				<span class="mt-0.5 block text-xs text-slate-500 tabular-nums">
					{qty(Number.parseFloat(holding.quantity))}
					{holding.symbol.split('/')[0]} · {holding.symbol}
				</span>
			</span>
			<span class="shrink-0 text-right">
				<Money amount={holding.value} currency="EUR" size="sm" weight="semibold" />
				<span class="mt-0.5 block">
					<TrendIndicator value={holding.unrealized_pct} size="sm" />
				</span>
			</span>
		</button>

		{#if open}
			<div class="border-t border-slate-200 bg-taupe-50 px-3 py-4">
				<div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
					<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">Daily price</p>
					<p class="text-sm text-slate-700">
						<Money amount={holding.price} currency="EUR" size="sm" />
						per {holding.symbol.split('/')[0]} · {longDate(holding.price_date)}{holding.price_carried_forward
							? ' (last known)'
							: ''}
					</p>
				</div>

				<div class="mt-3 grid grid-cols-3 gap-3">
					{@render figure('Paid', holding.paid)}
					{@render figure('Value now', holding.value)}
					{@render figure('Unrealized', holding.unrealized, true)}
				</div>

				<div class="mt-3 flex items-center justify-between gap-3 border-t border-slate-200 pt-3">
					<Text as="span" size="xs" tone="muted">
						Value since {longDate(holding.series[0].date)}
					</Text>
					<Sparkline
						data={holding.series.map((p) => p.value)}
						width={160}
						height={36}
						fill
						ariaLabel={`Value of ${holding.name} since the first purchase`}
					/>
				</div>

				<p class="mt-4 text-xs font-semibold tracking-wide text-slate-500 uppercase">Purchases</p>
				<table class="mt-1 w-full text-sm tabular-nums">
					<thead>
						<tr class="border-b border-slate-300 text-xs text-slate-500">
							<th scope="col" class="py-1 text-left font-medium">Date</th>
							<th scope="col" class="py-1 text-right font-medium">Quantity</th>
							<th scope="col" class="py-1 text-right font-medium">Unit price</th>
							<th scope="col" class="py-1 text-right font-medium">Paid</th>
						</tr>
					</thead>
					<tbody>
						{#each holding.purchases as purchase (purchase.id)}
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
							</tr>
						{/each}
					</tbody>
				</table>

				<div class="mt-3">
					<Button size="sm" variant="outline" intent="secondary" shape="default">
						<Icon icon="heroicons:plus" size="sm" />Add purchase
					</Button>
				</div>
			</div>
		{/if}
	</li>
{/snippet}

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
							color: '#059669',
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
			{empty ? 'No items yet' : `${items.length} of ${items.length + 1} items follow a daily price`}
		</Text>
	{/snippet}

	<div class="space-y-6">
		<div class="flex items-end justify-between gap-3 border-t border-slate-400 pt-3">
			<div>
				<Text as="span" size="sm" tone="muted">Current worth</Text>
				<div>
					<Money amount={classWorth} currency="EUR" size="xl" weight="semibold" />
				</div>
			</div>
			{#if !empty}
				<TrendIndicator value={14.2} />
			{/if}
		</div>

		<section>
			<h3 class="mb-2 text-sm text-slate-900">Items</h3>
			{#if empty}
				<div class="rounded-xl border border-slate-200 px-4 py-8 text-center">
					<Text as="p" size="sm" tone="muted">
						No items in this class yet. Add one with a value you set yourself, or link it to a daily
						price.
					</Text>
				</div>
			{:else}
				<ul class="divide-slate-200 rounded-xl border border-slate-200">
					{#each items as holding (holding.id)}
						{@render holdingRow(holding)}
					{/each}
					<li class="flex items-center gap-3 px-3 py-2.5">
						<span class="w-4 shrink-0"></span>
						<span class="min-w-0 flex-1">
							<span class="flex items-center gap-2">
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
			{/if}
		</section>
	</div>

	{#snippet footer()}
		<Button shape="default" onclick={() => (addOpen = true)}>
			<Icon icon="heroicons:plus" size="sm" />Add item
		</Button>
	{/snippet}
</Drawer>

<Dialog bind:open={addOpen} title="Add item to Crypto & metals" size="md">
	<div class="space-y-4">
		<div>
			<p class="mb-2 text-xs font-semibold tracking-wide text-slate-500 uppercase">
				How is it valued?
			</p>
			<Tabs
				tabs={[
					{ value: 'manual', label: 'Worth I set' },
					{ value: 'priced', label: 'Daily price' }
				]}
				bind:value={itemKind}
				ariaLabel="How the item is valued"
			/>
		</div>

		{#if itemKind === 'priced'}
			<FormField
				label="What do you own?"
				id="va-quote"
				hint="Crypto priced in euro. Instruments quoted in another currency stay manual."
			>
				{#snippet children(ctx)}
					<div
						id={ctx.id}
						class="flex items-center justify-between gap-3 rounded-md border border-slate-300 bg-white py-1.5 pr-1.5 pl-3"
					>
						<span class="flex min-w-0 items-center gap-2">
							<span class="text-sm font-medium text-slate-900">BTC/EUR</span>
							<span class="truncate text-sm text-slate-500">Bitcoin</span>
							<Badge intent="info" variant="soft" size="sm">EUR</Badge>
							<span class="shrink-0 text-xs text-slate-500">
								<Money amount={64000} currency="EUR" size="sm" /> · 17 Jun 2026
							</span>
						</span>
						<Button size="sm" variant="ghost" intent="secondary" shape="default">Change</Button>
					</div>
				{/snippet}
			</FormField>

			<div class="border-t border-slate-300 pt-4">
				<p class="mb-2 text-xs font-semibold tracking-wide text-slate-500 uppercase">
					First purchase
				</p>
				<div class="grid gap-3 sm:grid-cols-3">
					<FormField label="Date" id="va-date">
						<DatePicker
							bind:value={purchaseDate}
							max="2026-06-17"
							portal={false}
							placement="bottom-start"
							ariaLabel="Purchase date"
						/>
					</FormField>
					<FormField label="Quantity" id="va-qty" hint="BTC">
						{#snippet children(ctx)}
							<Input
								id={ctx.id}
								bind:value={quantity}
								ariaDescribedby={ctx.describedby}
								class="text-right tabular-nums"
							/>
						{/snippet}
					</FormField>
					<FormField label="Price per BTC" id="va-price">
						{#snippet children(ctx)}
							<CurrencyInput id={ctx.id} bind:value={unitPrice} />
						{/snippet}
					</FormField>
				</div>
				<div
					class="mt-3 flex items-baseline justify-between gap-3 border-t border-slate-200 pt-3"
				>
					<Text as="span" size="sm" tone="muted">Paid</Text>
					<Money amount={paidPreview} currency="EUR" size="md" weight="semibold" />
				</div>
			</div>
		{:else}
			<FormField label="Name" id="va-name">
				{#snippet children(ctx)}
					<Input id={ctx.id} value="" placeholder="e.g. Gold coins" />
				{/snippet}
			</FormField>
			<FormField label="Worth" id="va-worth">
				{#snippet children(ctx)}
					<CurrencyInput id={ctx.id} value="" />
				{/snippet}
			</FormField>
		{/if}
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (addOpen = false)}>Cancel</Button>
		<Button intent="success">Save</Button>
	{/snippet}
</Dialog>
