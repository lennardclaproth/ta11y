<script lang="ts">
	// An asset class as its own page: a narrow item ledger on the left, the selected daily-priced
	// item as an article on the right — its price, three figures, value against paid, and the
	// purchases that produced them.
	//
	// Below lg there is no room for two columns, so the page lands on the ledger and every item
	// opens its performance in place, exactly like the class drawer on Assets; "Open <item>"
	// follows through to the article and "All items" steps back.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import AddAssetItemDialog from '$lib/components/organisms/add-asset-item-dialog/AddAssetItemDialog.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Breadcrumb from '$lib/components/molecules/breadcrumb/Breadcrumb.svelte';
	import HoldingPerformanceRow from '$lib/components/molecules/holding-performance-row/HoldingPerformanceRow.svelte';
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
	import { untrack } from 'svelte';
	import { page } from '$app/state';
	import { getAssetClassDetails, getHolding } from '$lib/services/assets';
	import { accountStore } from '$lib/stores/account.svelte';
	import { decimalStringToNumber, formatQuantity } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { chartColors } from '$lib/charts/theme';
	import type { AssetClassDetails, AssetHolding, HoldingPurchase } from '$lib/api/types';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';

	const classId = $derived(page.params.classId ?? '');

	let details = $state<AssetClassDetails | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let selectedId = $state<string | null>(null);
	let view = $state<'ledger' | 'item'>('ledger');
	let expandedId = $state<string | null>(null);
	let addOpen = $state(false);
	/** Set while adding a purchase to an item that already exists, cleared for a new item. */
	let addTo = $state<AssetHolding | null>(null);

	const holdings = $derived(details?.holdings ?? []);
	const manual = $derived(details?.assets ?? []);
	const itemCount = $derived(holdings.length + manual.length);
	const summary = $derived(holdings.find((h) => h.id === selectedId) ?? holdings[0] ?? null);

	// The class payload carries every holding's figures but not its purchases — those are one
	// item's detail, not a class's. The article reads from the summary until the detail lands,
	// so the figures appear with the rest of the page and only the table fills in after.
	let detail = $state<AssetHolding | null>(null);
	const selected = $derived(detail?.id === summary?.id ? (detail ?? summary) : summary);
	const unit = $derived(selected ? selected.symbol.split('/')[0] || selected.name : '');
	const series = $derived(selected?.series ?? []);
	const purchases = $derived(selected?.purchases ?? []);

	// Paid in and Worth are real sums over every item in the class. Unrealized only means
	// something where something was paid for a daily price, so it is summed over the holdings
	// and dropped entirely when there are none — a zero there would read as "no gain" rather
	// than "nothing to gain on yet".
	const paidIn = $derived(
		holdings.reduce((total, h) => total + decimalStringToNumber(h.paid), 0) +
			manual.reduce((total, a) => total + decimalStringToNumber(a.current_worth), 0)
	);
	const worth = $derived(
		holdings.reduce((total, h) => total + decimalStringToNumber(h.value), 0) +
			manual.reduce((total, a) => total + decimalStringToNumber(a.current_worth), 0)
	);
	const unrealized = $derived(
		holdings.reduce((total, h) => total + decimalStringToNumber(h.unrealized), 0)
	);
	const holdingsPaid = $derived(
		holdings.reduce((total, h) => total + decimalStringToNumber(h.paid), 0)
	);
	// Nothing paid means no percentage to report, not a zero percent.
	const unrealizedPct = $derived(holdingsPaid > 0 ? (unrealized / holdingsPaid) * 100 : null);

	const classSummary = $derived(
		holdings.length === 0
			? [
					{ label: 'Paid in', amount: paidIn, colored: false },
					{ label: 'Worth', amount: worth, colored: false }
				]
			: [
					{ label: 'Paid in', amount: paidIn, colored: false },
					{ label: 'Worth', amount: worth, colored: false },
					{ label: 'Unrealized', amount: unrealized, colored: true }
				]
	);

	const holdingKpis = $derived<KpiItem[]>(
		selected
			? [
					{ label: 'Paid', amount: decimalStringToNumber(selected.paid), currency: 'EUR' },
					{ label: 'Value now', amount: decimalStringToNumber(selected.value), currency: 'EUR' },
					{
						label: 'Unrealized',
						amount: decimalStringToNumber(selected.unrealized),
						currency: 'EUR',
						change: selected.unrealized_pct ?? undefined
					}
				]
			: []
	);

	const dayTick = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', {
			month: 'short',
			day: 'numeric',
			timeZone: 'UTC'
		});

	async function load() {
		loading = true;
		error = null;
		try {
			await accountStore.ensureLoaded();
			details = await getAssetClassDetails(classId);
			// A deep link from the drawer names the item it was opened on.
			const wanted = page.url.searchParams.get('item');
			if (wanted && details.holdings.some((h) => h.id === wanted)) {
				selectedId = wanted;
				expandedId = wanted;
				view = 'item';
			} else if (!selectedId) {
				selectedId = details.holdings[0]?.id ?? null;
			}
		} catch {
			error = 'The items and their daily prices are unavailable right now.';
			details = null;
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		void classId;
		void load();
	});

	// Fetch the selected item's purchases, both when the selection changes and when the class is
	// reloaded — a purchase that was just saved is exactly what this table has to show.
	$effect(() => {
		const id = summary?.id;
		if (!id) {
			detail = null;
			return;
		}
		void loadDetail(id);
	});

	async function loadDetail(assetId: string) {
		try {
			const loaded = await getHolding(assetId);
			// A slow response for an item the user has already navigated away from is stale.
			if (loaded.id === untrack(() => summary?.id)) detail = loaded;
		} catch {
			// The figures come from the class payload and are already on screen; only the
			// purchase list is missing, and the table says so on its own.
			detail = null;
		}
	}

	function openItem(id: string) {
		selectedId = id;
		view = 'item';
	}
</script>

{#snippet quantityCell(row: HoldingPurchase)}
	<span class="tabular-nums">{formatQuantity(row.quantity)} {unit}</span>
{/snippet}

{#snippet unitPriceCell(row: HoldingPurchase)}
	<Money amount={decimalStringToNumber(row.unit_price)} currency="EUR" size="sm" />
{/snippet}

{#snippet paidCell(row: HoldingPurchase)}
	<Money amount={decimalStringToNumber(row.paid)} currency="EUR" size="sm" />
{/snippet}

{#snippet valueCell(row: HoldingPurchase)}
	<Money
		amount={decimalStringToNumber(row.quantity) * decimalStringToNumber(selected?.price ?? '0')}
		currency="EUR"
		size="sm"
	/>
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar title={details?.class.name ?? 'Asset class'} />
	{/snippet}

	<div class="relative flex min-h-full flex-col gap-5 px-4 pb-6 lg:h-full lg:min-h-0 lg:px-8">
		<!-- Where you are, and what the whole class is worth. The figures that matter per item live
		     in the reading column, so this stays one rule high. -->
		<div
			class="flex shrink-0 flex-wrap items-baseline justify-between gap-x-6 gap-y-2 border-b border-slate-300 pb-2"
		>
			<Breadcrumb
				items={[
					{ label: 'Assets', href: '/assets' },
					{ label: details?.class.name ?? 'Asset class' }
				]}
			/>
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
					{#if unrealizedPct !== null}
						<TrendIndicator value={unrealizedPct} size="sm" />
					{/if}
				</dl>
			{/if}
		</div>

		{#if !loading && !error && selected?.price_carried_forward}
			<!-- The figures stay on screen, so the warning says which price they are based on rather
			     than claiming nothing could be read. -->
			<Alert intent="warning" title="Prices are not up to date">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<span>
						Today's daily prices have not come in yet. The values below still use the price of
						{formatDisplayDate(selected.price_date ?? null)}.
					</span>
					<Button
						size="sm"
						variant="outline"
						intent="secondary"
						shape="default"
						onclick={() => void load()}
					>
						Try again
					</Button>
				</div>
			</Alert>
		{/if}

		{#if error}
			<!-- Nothing could be read, so the page does not pretend to show figures. -->
			<Alert intent="error" title="This asset class could not be loaded">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<span>{error} Nothing was changed; your purchases are still recorded.</span>
					<Button
						size="sm"
						variant="outline"
						intent="secondary"
						shape="default"
						onclick={() => void load()}
					>
						Try again
					</Button>
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
					<LedgerToolbar
						title="Items"
						actionLabel="Add item"
						onAdd={() => {
							addTo = null;
							addOpen = true;
						}}
					/>
					{#if loading}
						<div class="space-y-2 p-3" aria-busy="true">
							{#each [0, 1, 2] as i (i)}
								<Skeleton variant="rect" class="h-12 w-full" />
							{/each}
						</div>
					{:else if itemCount === 0}
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
							{#each holdings as holding (holding.id)}
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
							{#each manual as asset (asset.id)}
								<li
									class="flex items-center gap-3 border-b border-slate-200 px-3 py-2.5 last:border-b-0"
								>
									<span class="min-w-0 flex-1">
										<span class="flex flex-wrap items-center gap-2">
											<span class="truncate text-sm text-slate-900">{asset.name}</span>
											<Badge intent="neutral" variant="soft" size="sm">Manual</Badge>
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
					{:else if !selected}
						<div
							class="flex flex-1 flex-col items-center justify-center gap-3 border-t border-slate-400 px-6 py-16 text-center"
						>
							<Heading level="h2" size="lg" class="text-slate-900">Nothing to follow yet</Heading>
							<Text as="p" size="sm" tone="muted" class="max-w-sm">
								Link an item to a daily price and record what you bought. From that day on its value
								follows the price, back in time as well, and counts towards your net worth.
							</Text>
							<Button
								shape="default"
								onclick={() => {
									addTo = null;
									addOpen = true;
								}}
							>
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
									{selected.symbol} · {formatQuantity(selected.quantity)}
									{unit} · average cost
									<Money
										amount={decimalStringToNumber(selected.avg_unit_price)}
										currency="EUR"
										size="sm"
									/> per {unit}
								</Text>
							</div>
							<!-- The price block only hangs on the right once there is a column to hang it on;
							     stacked, a right-aligned label reads as a stray fragment. -->
							<div class="text-left sm:text-right">
								<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">
									Price per {unit}
								</p>
								<p class="leading-tight">
									<Money
										amount={decimalStringToNumber(selected.price)}
										currency="EUR"
										size="xl"
										weight="semibold"
									/>
								</p>
								<Text as="span" size="sm" tone="muted">
									{#if selected.price_date}
										{formatDisplayDate(selected.price_date)}{selected.price_carried_forward
											? ' · last known price, carried forward'
											: ''}
									{:else}
										No price yet
									{/if}
								</Text>
							</div>
						</header>

						<KpiRow columns={3} items={holdingKpis} />

						{#if series.length > 0}
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
										Since {formatDisplayDate(series[0].date)} you paid
										<Money
											amount={decimalStringToNumber(selected.paid)}
											currency="EUR"
											size="sm"
										/> and it is worth
										<Money
											amount={decimalStringToNumber(selected.value)}
											currency="EUR"
											size="sm"
										/> today.
									</span>
								</div>
								<TimeSeriesChart
									height="h-44"
									labels={series.map((p) => p.date)}
									xTickFormat={dayTick}
									datasets={[
										{
											label: 'Value',
											data: series.map((p) => decimalStringToNumber(p.value)),
											color: chartColors.positive,
											fill: true
										},
										{
											label: 'Paid',
											data: series.map((p) => decimalStringToNumber(p.paid)),
											color: chartColors.net,
											dashed: true
										}
									]}
									ariaLabel={`Value of ${selected.name} against what was paid for it`}
								/>
							</AnalyticsCard>
						{/if}

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
									addTo = selected;
									addOpen = true;
								}}
							/>
							<DataTable
								rows={purchases}
								emptyText="No purchases yet"
								columns={[
									{
										key: 'date',
										header: 'Date',
										value: (r: HoldingPurchase) => formatDisplayDate(r.date)
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

<AddAssetItemDialog
	bind:open={addOpen}
	{classId}
	className={details?.class.name ?? ''}
	holding={addTo}
	onSaved={() => void load()}
/>
