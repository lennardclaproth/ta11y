<script lang="ts">
	// An asset class as its own page: a narrow item ledger on the left and a reading column on the
	// right. The column opens on the whole class — what it is worth against what was paid in over
	// the period in the top bar — and switches to one item when it is picked: a daily-priced item
	// with its price and purchases, or an item with a worth you set yourself with its changes.
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
	import AssetWorthDialog from '$lib/components/organisms/asset-worth-dialog/AssetWorthDialog.svelte';
	import NameDialog from '$lib/components/organisms/name-dialog/NameDialog.svelte';
	import ConfirmDialog from '$lib/components/organisms/confirm-dialog/ConfirmDialog.svelte';
	import ActionMenu from '$lib/components/molecules/action-menu/ActionMenu.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
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
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		deleteAsset,
		deleteAssetClass,
		deleteMutation,
		deletePurchase,
		getAssetClassDetails,
		getHolding,
		renameAsset,
		updateAssetClass
	} from '$lib/services/assets';
	import { getPortfolioSnapshots } from '$lib/services/portfolio';
	import { accountStore } from '$lib/stores/account.svelte';
	import { periodStore } from '$lib/stores/period.svelte';
	import { decimalStringToNumber, formatQuantity } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { chartColors } from '$lib/charts/theme';
	import type {
		Asset,
		AssetClassDetails,
		AssetHolding,
		AssetMutation,
		HoldingPurchase,
		PortfolioSnapshotPoint
	} from '$lib/api/types';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';
	import {
		dailyPoints,
		figuresFor,
		holdingPoints,
		manualItemPoints,
		portfolioPoints,
		sumPoints,
		trimLeading,
		type WorthFigures,
		type WorthPoint
	} from './class-series';

	/** The selection that shows the whole class rather than one item. */
	const OVERVIEW = 'overview';

	const classId = $derived(page.params.classId ?? '');
	const from = $derived(periodStore.from);
	const to = $derived(periodStore.to);

	let details = $state<AssetClassDetails | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let snapshots = $state<PortfolioSnapshotPoint[]>([]);
	let snapshotsError = $state<string | null>(null);

	let selectedId = $state<string>(OVERVIEW);
	let view = $state<'ledger' | 'item'>('ledger');
	let expandedId = $state<string | null>(null);
	let addOpen = $state(false);
	/** Set while adding a purchase to an item that already exists, cleared for a new item. */
	let addTo = $state<AssetHolding | null>(null);
	let worthOpen = $state(false);
	/** Set while correcting a recorded change instead of adding one. */
	let worthEditing = $state<AssetMutation | null>(null);
	/** Set while correcting a recorded purchase instead of adding one. */
	let editPurchase = $state<HoldingPurchase | null>(null);

	// One rename and one delete dialog serve the class, its items, changes and purchases; each
	// action fills in what is renamed or deleted and how.
	let nameOpen = $state(false);
	let nameDialog = $state<{
		title: string;
		value: string;
		onSave: (name: string) => Promise<void>;
	}>({ title: '', value: '', onSave: async () => {} });
	let confirmOpen = $state(false);
	let confirmDialog = $state<{
		title: string;
		message: string;
		confirmLabel: string;
		onConfirm: () => Promise<void>;
	}>({ title: '', message: '', confirmLabel: 'Delete', onConfirm: async () => {} });

	/** With two columns a row selects; on a phone it opens in place first, like the drawer. */
	let wide = $state(false);

	$effect(() => {
		const query = window.matchMedia('(min-width: 1024px)');
		wide = query.matches;
		const onChange = (event: MediaQueryListEvent) => (wide = event.matches);
		query.addEventListener('change', onChange);
		return () => query.removeEventListener('change', onChange);
	});
	const rowMode = $derived(wide || view === 'item' ? 'select' : 'expand');

	// The Portfolio class is kept by the portfolio sync; the API refuses items added to it, and
	// its worth against cost basis comes from the portfolio itself.
	const canAdd = $derived(details?.class.source.toUpperCase() === 'MANUAL');
	const isPortfolio = $derived(details?.class.source.toUpperCase() === 'PORTFOLIO');
	const holdings = $derived(details?.holdings ?? []);
	const manual = $derived(details?.assets ?? []);
	const mutations = $derived(details?.mutations ?? []);
	const itemCount = $derived(holdings.length + manual.length);

	const selectedHolding = $derived(holdings.find((h) => h.id === selectedId) ?? null);
	const selectedManual = $derived(
		isPortfolio ? null : (manual.find((a) => a.id === selectedId) ?? null)
	);

	// The class payload carries every holding's figures but not its purchases — those are one
	// item's detail, not a class's. The article reads from the summary until the detail lands,
	// so the figures appear with the rest of the page and only the table fills in after.
	let detail = $state<AssetHolding | null>(null);
	const selected = $derived(
		detail?.id === selectedHolding?.id ? (detail ?? selectedHolding) : selectedHolding
	);
	const unit = $derived(selected ? selected.symbol.split('/')[0] || selected.name : '');
	const purchases = $derived(selected?.purchases ?? []);

	// Every series is laid out day by day over the period, so items can be added together and the
	// first day of the period is what growth is measured from.
	const manualDaily = $derived(
		new Map(manual.map((a) => [a.id, dailyPoints(manualItemPoints(mutations, a.id), from, to)]))
	);
	const holdingDaily = $derived(
		new Map(holdings.map((h) => [h.id, dailyPoints(holdingPoints(h), from, to)]))
	);
	const classDaily = $derived(
		isPortfolio
			? dailyPoints(portfolioPoints(snapshots), from, to)
			: sumPoints([...manualDaily.values(), ...holdingDaily.values()])
	);
	const classFigures = $derived(figuresFor(classDaily));

	const selectionDaily = $derived<WorthPoint[]>(
		selectedHolding
			? (holdingDaily.get(selectedHolding.id) ?? [])
			: selectedManual
				? (manualDaily.get(selectedManual.id) ?? [])
				: classDaily
	);
	const selectionFigures = $derived(figuresFor(selectionDaily));
	const chartPoints = $derived(trimLeading(selectionDaily));

	const manualChanges = $derived(
		selectedManual
			? mutations
					.filter((m) => m.item_id === selectedManual.id)
					.sort(
						(a, b) =>
							b.effective_date.localeCompare(a.effective_date) ||
							b.created_at.localeCompare(a.created_at)
					)
			: []
	);
	const openingChangeId = $derived(manualChanges.at(-1)?.id ?? null);

	const periodLabel = $derived(periodStore.presetLabel);
	const longRange = $derived(classDaily.length > 400);

	function kpisFor(figures: WorthFigures | null, paidLabel = 'Paid in'): KpiItem[] {
		if (!figures) return [];
		return [
			{ label: paidLabel, amount: figures.paid, currency: 'EUR' },
			{ label: 'Worth', amount: figures.worth, currency: 'EUR' },
			{
				label: 'Unrealized',
				amount: figures.unrealized,
				currency: 'EUR',
				change: figures.unrealizedPct ?? undefined
			},
			{
				label: `Growth · ${periodLabel}`,
				amount: figures.growth,
				currency: 'EUR',
				change: figures.growthPct ?? undefined
			}
		];
	}

	const paidLabel = $derived(isPortfolio ? 'Cost basis' : 'Paid in');

	const dayTick = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString(
			'en',
			longRange
				? { month: 'short', year: '2-digit', timeZone: 'UTC' }
				: { month: 'short', day: 'numeric', timeZone: 'UTC' }
		);

	async function load() {
		loading = true;
		error = null;
		try {
			await accountStore.ensureLoaded();
			details = await getAssetClassDetails(classId);
			// A deep link from the drawer names the item it was opened on.
			const wanted = page.url.searchParams.get('item');
			const known =
				wanted &&
				(details.holdings.some((h) => h.id === wanted) ||
					details.assets.some((a) => a.id === wanted));
			if (wanted && known) {
				selectedId = wanted;
				expandedId = wanted;
				view = 'item';
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
		untrack(() => {
			selectedId = OVERVIEW;
			void load();
		});
	});

	// The portfolio's worth against cost basis comes from its own snapshots, for the period shown.
	$effect(() => {
		if (!isPortfolio) {
			snapshots = [];
			return;
		}
		void loadSnapshots(from, to);
	});

	async function loadSnapshots(rangeFrom: string, rangeTo: string) {
		snapshotsError = null;
		try {
			snapshots = await getPortfolioSnapshots({ from: rangeFrom, to: rangeTo });
		} catch {
			snapshots = [];
			snapshotsError = 'The portfolio history could not be loaded.';
		}
	}

	// Fetch the selected item's purchases, both when the selection changes and when the class is
	// reloaded — a purchase that was just saved is exactly what this table has to show.
	$effect(() => {
		const id = selectedHolding?.id;
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
			if (loaded.id === untrack(() => selectedHolding?.id)) detail = loaded;
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

	function changeLabel(row: AssetMutation): string {
		if (row.id === openingChangeId) return 'Opening worth';
		if (row.change_type.toUpperCase() === 'ADJUST') {
			return row.direction?.toUpperCase() === 'DECREASE' ? 'Withdrawn' : 'Added';
		}
		return 'Set worth';
	}

	/** What the change did to the worth, signed. */
	function changeAmount(row: AssetMutation): number {
		return decimalStringToNumber(row.new_worth) - decimalStringToNumber(row.previous_worth);
	}

	function askRename(title: string, value: string, onSave: (name: string) => Promise<void>) {
		nameDialog = { title, value, onSave };
		nameOpen = true;
	}

	function askDelete(
		title: string,
		message: string,
		confirmLabel: string,
		onConfirm: () => Promise<void>
	) {
		confirmDialog = { title, message, confirmLabel, onConfirm };
		confirmOpen = true;
	}

	const classMenu = $derived(
		details && canAdd
			? [
					{
						label: 'Rename class',
						icon: 'heroicons:pencil',
						onSelect: () =>
							askRename('Rename class', details?.class.name ?? '', async (name) => {
								await updateAssetClass({ id: classId, name });
								await load();
							})
					},
					{
						label: 'Delete class',
						icon: 'heroicons:trash',
						intent: 'danger' as const,
						divider: true,
						onSelect: () =>
							askDelete(
								`Delete ${details?.class.name ?? 'this class'}?`,
								`The class, its ${itemCount} ${itemCount === 1 ? 'item' : 'items'} and their whole history are removed, and your net worth is recalculated without them. This cannot be undone.`,
								'Delete class',
								async () => {
									await deleteAssetClass(classId, {});
									await goto('/assets');
								}
							)
					}
				]
			: []
	);

	function itemMenu(item: { id: string; name: string }) {
		return [
			{
				label: 'Rename',
				icon: 'heroicons:pencil',
				onSelect: () =>
					askRename('Rename item', item.name, async (name) => {
						await renameAsset(item.id, { name });
						await load();
					})
			},
			{
				label: 'Delete',
				icon: 'heroicons:trash',
				intent: 'danger' as const,
				divider: true,
				onSelect: () =>
					askDelete(
						`Delete ${item.name}?`,
						`${item.name} and its whole history are removed, and the class and your net worth are recalculated without it. This cannot be undone.`,
						'Delete item',
						async () => {
							await deleteAsset(item.id);
							selectedId = OVERVIEW;
							await load();
						}
					)
			}
		];
	}

	function editChange(row: AssetMutation) {
		worthEditing = row;
		worthOpen = true;
	}

	function removeChange(row: AssetMutation) {
		if (!selectedManual) return;
		const assetId = selectedManual.id;
		askDelete(
			'Delete this change?',
			`${changeLabel(row)} on ${formatDisplayDate(row.effective_date.slice(0, 10))} is removed and the history after it recalculated.`,
			'Delete change',
			async () => {
				await deleteMutation(assetId, row.id);
				await load();
			}
		);
	}

	function editPurchaseRow(row: HoldingPurchase) {
		addTo = selected;
		editPurchase = row;
		addOpen = true;
	}

	function removePurchase(row: HoldingPurchase) {
		if (!selected) return;
		const assetId = selected.id;
		askDelete(
			'Delete this purchase?',
			`The purchase of ${formatQuantity(row.quantity)} ${unit} on ${formatDisplayDate(row.date)} is removed and the value of ${selected.name} recalculated without it.`,
			'Delete purchase',
			async () => {
				await deletePurchase(assetId, row.id);
				await load();
			}
		);
	}

	function itemGrowthPct(asset: Asset): number | null {
		return figuresFor(manualDaily.get(asset.id) ?? [])?.growthPct ?? null;
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

{#snippet changeAmountCell(row: AssetMutation)}
	<Money
		amount={changeAmount(row)}
		currency="EUR"
		size="sm"
		colored
		signDisplay="exceptZero"
	/>
{/snippet}

{#snippet worthAfterCell(row: AssetMutation)}
	<Money amount={decimalStringToNumber(row.new_worth)} currency="EUR" size="sm" />
{/snippet}

{#snippet changeActionsCell(row: AssetMutation)}
	{#if canAdd}
		<span class="flex justify-end gap-1">
			<IconButton
				icon="heroicons:pencil"
				ariaLabel={`Edit ${changeLabel(row).toLowerCase()} on ${formatDisplayDate(row.effective_date.slice(0, 10))}`}
				variant="ghost"
				intent="secondary"
				size="sm"
				onclick={() => editChange(row)}
			/>
			{#if manualChanges.length > 1}
				<IconButton
					icon="heroicons:trash"
					ariaLabel={`Delete ${changeLabel(row).toLowerCase()} on ${formatDisplayDate(row.effective_date.slice(0, 10))}`}
					variant="ghost"
					intent="error"
					size="sm"
					onclick={() => removeChange(row)}
				/>
			{/if}
		</span>
	{/if}
{/snippet}

{#snippet purchaseActionsCell(row: HoldingPurchase)}
	<span class="flex justify-end gap-1">
		<IconButton
			icon="heroicons:pencil"
			ariaLabel={`Edit purchase on ${formatDisplayDate(row.date)}`}
			variant="ghost"
			intent="secondary"
			size="sm"
			onclick={() => editPurchaseRow(row)}
		/>
		{#if purchases.length > 1}
			<IconButton
				icon="heroicons:trash"
				ariaLabel={`Delete purchase on ${formatDisplayDate(row.date)}`}
				variant="ghost"
				intent="error"
				size="sm"
				onclick={() => removePurchase(row)}
			/>
		{/if}
	</span>
{/snippet}

{#snippet valueChart(title: string, ariaLabel: string)}
	<AnalyticsCard {title}>
		<!-- Two series need naming: the chart's own tooltip is hover-only, so the key and the
		     sentence under it carry the same reading without a pointer. -->
		<div
			class="mb-3 flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1 text-xs text-slate-600"
		>
			<span class="flex flex-wrap items-center gap-x-5 gap-y-1">
				<span class="flex items-center gap-2">
					<span class="inline-block h-0.5 w-6 bg-emerald-600"></span>Worth
				</span>
				<span class="flex items-center gap-2">
					<span class="inline-block w-6 border-t-2 border-dashed border-slate-700"></span>
					{paidLabel}
				</span>
			</span>
			{#if selectionFigures}
				<span class="text-slate-500">
					{periodLabel}: {selectionFigures.growth >= 0 ? 'grew' : 'shrank'}
					<Money
						amount={Math.abs(selectionFigures.growth)}
						currency="EUR"
						size="sm"
					/> on top of what was paid in.
				</span>
			{/if}
		</div>
		<TimeSeriesChart
			height="h-44"
			labels={chartPoints.map((p) => p.date)}
			xTickFormat={dayTick}
			datasets={[
				{
					label: 'Worth',
					data: chartPoints.map((p) => p.value),
					color: chartColors.positive,
					fill: true
				},
				{
					label: paidLabel,
					data: chartPoints.map((p) => p.paid),
					color: chartColors.net,
					dashed: true
				}
			]}
			{ariaLabel}
		/>
	</AnalyticsCard>
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar />
	{/snippet}

	<div class="relative flex min-h-full flex-col gap-5 px-4 pb-6 lg:h-full lg:min-h-0 lg:px-8">
		<!-- Where you are, and what the whole class is worth over the period. The figures that
		     matter per item live in the reading column, so this stays one line high. -->
		<div class="flex shrink-0 flex-wrap items-baseline justify-between gap-x-6 gap-y-2">
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
			{:else if !error && classFigures}
				<dl class="flex flex-wrap items-baseline gap-x-6 gap-y-1">
					{#each [{ label: paidLabel, amount: classFigures.paid, colored: false }, { label: 'Worth', amount: classFigures.worth, colored: false }, { label: 'Unrealized', amount: classFigures.unrealized, colored: true }] as entry (entry.label)}
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
					{#if classFigures.unrealizedPct !== null}
						<TrendIndicator value={classFigures.unrealizedPct} size="sm" />
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
					<span>{error} Nothing was changed; your items are still recorded.</span>
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
					<LedgerToolbar title="Items">
						{#snippet actions()}
							{#if canAdd}
								<Button
									shape="default"
									onclick={() => {
										addTo = null;
										editPurchase = null;
										addOpen = true;
									}}
								>
									<Icon icon="heroicons:plus" />
									Add item
								</Button>
							{/if}
						{/snippet}
					</LedgerToolbar>
					{#if loading}
						<div class="space-y-2 p-3" aria-busy="true">
							{#each [0, 1, 2] as i (i)}
								<Skeleton variant="rect" class="h-12 w-full" />
							{/each}
						</div>
					{:else}
						<ul class="min-h-0 flex-1 overflow-y-auto">
							<li class="border-b border-slate-200">
								<button
									type="button"
									aria-current={selectedId === OVERVIEW ? 'true' : undefined}
									class={[
										'flex w-full items-center gap-3 px-3 py-2.5 text-left transition-colors hover:bg-slate-50',
										'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
										selectedId === OVERVIEW ? 'bg-amber-50' : ''
									]
										.filter(Boolean)
										.join(' ')}
									onclick={() => openItem(OVERVIEW)}
								>
									<span class="min-w-0 flex-1">
										<span class="block text-sm font-medium text-slate-900">Whole class</span>
										<span class="mt-0.5 block text-xs text-slate-500">
											{itemCount}
											{itemCount === 1 ? 'item' : 'items'} · {periodLabel}
										</span>
									</span>
									{#if classFigures}
										<span class="shrink-0 text-right">
											<Money
												amount={classFigures.worth}
												currency="EUR"
												size="sm"
												weight="semibold"
											/>
											{#if classFigures.growthPct !== null}
												<span class="mt-0.5 block">
													<TrendIndicator value={classFigures.growthPct} size="sm" />
												</span>
											{/if}
										</span>
									{/if}
								</button>
							</li>
							{#each holdings as holding (holding.id)}
								<HoldingPerformanceRow
									{holding}
									mode={rowMode}
									open={expandedId === holding.id}
									current={selectedId === holding.id}
									showKind={rowMode === 'expand' || wide}
									onActivate={() =>
										rowMode === 'expand'
											? (expandedId = expandedId === holding.id ? null : holding.id)
											: openItem(holding.id)}
									onOpenItem={() => openItem(holding.id)}
								/>
							{/each}
							{#each manual as asset (asset.id)}
								{@const growthPct = itemGrowthPct(asset)}
								<li class="border-b border-slate-200 last:border-b-0">
									{#if isPortfolio}
										<!-- The portfolio's one item is the class itself; the overview above reads it. -->
										<div class="flex items-center gap-3 px-3 py-2.5">
											<span class="min-w-0 flex-1 truncate text-sm text-slate-900">{asset.name}</span>
											<Money
												amount={decimalStringToNumber(asset.current_worth)}
												currency="EUR"
												size="sm"
												weight="semibold"
											/>
										</div>
									{:else}
										<button
											type="button"
											aria-current={selectedId === asset.id ? 'true' : undefined}
											class={[
												'flex w-full items-center gap-3 px-3 py-2.5 text-left transition-colors hover:bg-slate-50',
												'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
												selectedId === asset.id ? 'bg-amber-50' : ''
											]
												.filter(Boolean)
												.join(' ')}
											onclick={() => openItem(asset.id)}
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
												{#if growthPct !== null}
													<span class="mt-0.5 block">
														<TrendIndicator value={growthPct} size="sm" />
													</span>
												{/if}
											</span>
										</button>
									{/if}
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
						<!-- The whole column loads as one: a figure that is already on screen while its chart
						     is still empty reads as a result, not as a wait. -->
						<div class="space-y-4" aria-busy="true">
							<Skeleton variant="rect" class="h-9 w-56" />
							<div class="grid gap-3 sm:grid-cols-4">
								<Skeleton variant="rect" class="h-16 w-full" />
								<Skeleton variant="rect" class="h-16 w-full" />
								<Skeleton variant="rect" class="h-16 w-full" />
								<Skeleton variant="rect" class="h-16 w-full" />
							</div>
							<Skeleton variant="rect" class="h-44 w-full rounded-xl" />
						</div>
					{:else if selected}
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
							<div class="flex items-end gap-4">
								<ActionMenu label="Item" items={itemMenu(selected)} />
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
							</div>
						</header>

						<KpiRow columns={4} items={kpisFor(selectionFigures)} />

						{#if chartPoints.length > 0}
							{@render valueChart(
								'Value against paid',
								`Value of ${selected.name} against what was paid for it`
							)}
						{/if}

						<Panel
							variant="muted"
							shape="square"
							padding="none"
							class="flex min-h-56 flex-col overflow-hidden"
						>
							<LedgerToolbar title="Purchases">
								{#snippet actions()}
									<Button
										shape="default"
										onclick={() => {
											addTo = selected;
											editPurchase = null;
											addOpen = true;
										}}
									>
										<Icon icon="heroicons:plus" />
										Add purchase
									</Button>
								{/snippet}
							</LedgerToolbar>
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
									{ key: 'value', header: 'Value now', align: 'right', cell: valueCell },
									{ key: 'actions', header: '', align: 'right', cell: purchaseActionsCell }
								]}
							/>
						</Panel>
					{:else if selectedManual}
						<header
							class="flex flex-wrap items-end justify-between gap-x-6 gap-y-2 border-t border-slate-400 pt-3"
						>
							<div class="min-w-0">
								<div class="flex flex-wrap items-center gap-2">
									<Heading level="h2" size="xl" class="text-slate-900">
										{selectedManual.name}
									</Heading>
									<Badge intent="neutral" variant="soft">Worth you set</Badge>
								</div>
								<Text as="p" size="sm" tone="muted">
									Last changed on {formatDisplayDate(selectedManual.updated_at.slice(0, 10))}
								</Text>
							</div>
							{#if canAdd}
								<div class="flex items-center gap-2">
									<ActionMenu label="Item" items={itemMenu(selectedManual)} />
									<Button
										shape="default"
										onclick={() => {
											worthEditing = null;
											worthOpen = true;
										}}
									>
										<Icon icon="heroicons:pencil-square" />Update worth
									</Button>
								</div>
							{/if}
						</header>

						<KpiRow columns={4} items={kpisFor(selectionFigures)} />

						{#if chartPoints.length > 0}
							{@render valueChart(
								'Worth against paid in',
								`Worth of ${selectedManual.name} against what was paid in`
							)}
						{:else}
							<Text as="p" size="sm" tone="muted">
								{selectedManual.name} has no worth recorded in this period.
							</Text>
						{/if}

						<Panel
							variant="muted"
							shape="square"
							padding="none"
							class="flex min-h-56 flex-col overflow-hidden"
						>
							<LedgerToolbar title="Changes" />
							<DataTable
								rows={manualChanges}
								emptyText="No changes recorded"
								columns={[
									{
										key: 'date',
										header: 'Date',
										value: (r: AssetMutation) => formatDisplayDate(r.effective_date.slice(0, 10))
									},
									{ key: 'kind', header: 'Change', value: (r: AssetMutation) => changeLabel(r) },
									{ key: 'note', header: 'Note', value: (r: AssetMutation) => r.note ?? '' },
									{ key: 'amount', header: 'Amount', align: 'right', cell: changeAmountCell },
									{ key: 'worth', header: 'Worth after', align: 'right', cell: worthAfterCell },
									{ key: 'actions', header: '', align: 'right', cell: changeActionsCell }
								]}
							/>
						</Panel>
					{:else}
						<header
							class="flex flex-wrap items-end justify-between gap-x-6 gap-y-2 border-t border-slate-400 pt-3"
						>
							<div class="min-w-0">
								<div class="flex flex-wrap items-center gap-2">
									<Heading level="h2" size="xl" class="text-slate-900">
										{details?.class.name ?? 'Asset class'}
									</Heading>
									{#if isPortfolio}
										<Badge intent="info" variant="soft">From your portfolio</Badge>
									{/if}
								</div>
								<Text as="p" size="sm" tone="muted">
									{periodLabel} · {periodStore.rangeLabel}
								</Text>
							</div>
							{#if classMenu.length > 0}
								<ActionMenu label="Class" items={classMenu} />
							{/if}
							{#if isPortfolio}
								<Button
									variant="outline"
									intent="secondary"
									shape="default"
									onclick={() => void goto('/portfolio')}
								>
									Open Portfolio<Icon icon="heroicons:arrow-right" size="sm" />
								</Button>
							{/if}
						</header>

						{#if snapshotsError}
							<Alert intent="error" title="The portfolio history could not be loaded">
								{snapshotsError} The worth on the left is still current.
							</Alert>
						{/if}

						{#if itemCount === 0}
							<div
								class="flex flex-1 flex-col items-center justify-center gap-3 px-6 py-16 text-center"
							>
								<Heading level="h2" size="lg" class="text-slate-900">Nothing in this class yet</Heading>
								<Text as="p" size="sm" tone="muted" class="max-w-sm">
									Add an item with a worth you set yourself, like savings, or link one to a daily
									price. Its worth over time shows up here.
								</Text>
								{#if canAdd}
									<Button
										shape="default"
										onclick={() => {
											addTo = null;
											editPurchase = null;
											addOpen = true;
										}}
									>
										<Icon icon="heroicons:plus" size="sm" />Add item
									</Button>
								{/if}
							</div>
						{:else}
							<KpiRow columns={4} items={kpisFor(selectionFigures, paidLabel)} />
							{#if chartPoints.length > 0}
								{@render valueChart(
									isPortfolio ? 'Value against cost basis' : 'Worth against paid in',
									`Worth of ${details?.class.name ?? 'this class'} against what was paid in`
								)}
							{:else}
								<Text as="p" size="sm" tone="muted">Nothing was held in this period.</Text>
							{/if}
						{/if}
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
	purchase={editPurchase}
	onSaved={() => void load()}
/>

<AssetWorthDialog
	bind:open={worthOpen}
	asset={selectedManual}
	editing={worthEditing}
	opening={worthEditing !== null && worthEditing.id === openingChangeId}
	onSaved={() => void load()}
/>

<NameDialog
	bind:open={nameOpen}
	title={nameDialog.title}
	value={nameDialog.value}
	onSave={nameDialog.onSave}
/>

<ConfirmDialog
	bind:open={confirmOpen}
	title={confirmDialog.title}
	message={confirmDialog.message}
	confirmLabel={confirmDialog.confirmLabel}
	onConfirm={confirmDialog.onConfirm}
/>
