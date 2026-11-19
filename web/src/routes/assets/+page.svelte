<script lang="ts">
	import { tick } from 'svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import { chartHeights } from '$lib/components/organisms/charts/charts.variants';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import RankedBarChart from '$lib/components/organisms/charts/RankedBarChart.svelte';
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import AssetClassDrawer from '$lib/components/organisms/asset-class-drawer/AssetClassDrawer.svelte';
	import AddAssetItemDialog from '$lib/components/organisms/add-asset-item-dialog/AddAssetItemDialog.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import {
		listAssetClasses,
		getAssetClassGrowth,
		getAssetClassDetails,
		createAssetClass
	} from '$lib/services/assets';
	import { goto } from '$app/navigation';
	import { accountStore } from '$lib/stores/account.svelte';
	import { periodStore } from '$lib/stores/period.svelte';
	import { createDraggedRange } from '$lib/stores/dragged-range.svelte';
	import { todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import { toast } from '$lib/stores/toast.svelte';
	import { decimalStringToNumber } from '$lib/api/money';
	import { categoryRamp } from '$lib/charts/theme';
	import type { NoticeMessage } from '$lib/components/molecules/notice-band/notice-band.types';
	import type { AssetClass, AssetClassDetails, AssetClassGrowthSeries } from '$lib/api/types';

	let classes = $state<AssetClass[]>([]);
	let growth = $state<AssetClassGrowthSeries[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let drawerOpen = $state(false);
	let details = $state<AssetClassDetails | null>(null);
	let detailsLoading = $state(false);

	// Adding an item leaves the class drawer open, and a drawer draws its own scrim over the notice
	// region under the masthead. So that outcome is reported in the drawer it was started from
	// rather than as a toast nobody would see.
	let classNotice = $state<NoticeMessage | null>(null);

	// One period for the whole app, chosen in the account overview.
	const from = $derived(periodStore.from);
	const to = $derived(periodStore.to);
	// Dragging across the chart zooms the period to that stretch; a click alone keeps it.
	const draggedRange = createDraggedRange();
	function onRangeSelect(rangeFrom: string, rangeTo: string) {
		if (rangeFrom !== rangeTo) draggedRange.select(rangeFrom, rangeTo);
	}

	let createOpen = $state(false);
	let className = $state('');
	let creatingClass = $state(false);
	let createError = $state<string | null>(null);

	let addItemOpen = $state(false);

	const euro = (n: number) => `€${n.toLocaleString('en', { maximumFractionDigits: 0 })}`;
	const monthShort = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	// Live classes ranked by worth. A class keeps its colour by rank, in the list and in the
	// stacked chart alike, so the two read as one key.
	const ranked = $derived(
		classes
			.filter((c) => !c.archived)
			.map((c) => ({ id: c.id, name: c.name, worth: decimalStringToNumber(c.current_worth) }))
			.sort((a, b) => b.worth - a.worth)
	);
	const classColor = $derived(
		new Map(ranked.map((c, i) => [c.id, categoryRamp[i % categoryRamp.length]]))
	);
	const total = $derived(ranked.reduce((sum, c) => sum + c.worth, 0));

	// Each class changes worth on its own days. The chart needs one set of dates, so every class
	// carries its last known worth forward to the dates the other classes moved on, and is nothing
	// before its first. The period's own ends are dates too — its end held at today while the
	// period runs on — so the chart spans the whole period even where nothing changed in it.
	const periodEnd = $derived(to && to < todayISO() ? to : todayISO());
	const stackDates = $derived(
		[...new Set([from, ...growth.flatMap((g) => g.growth.map((p) => p.date)), periodEnd])]
			.filter((date) => !!date && date >= from && date <= periodEnd)
			.sort()
	);
	const classSeries = $derived(
		ranked
			.map((c) => growth.find((g) => g.class_id === c.id))
			.filter((g): g is AssetClassGrowthSeries => !!g && g.growth.length > 0)
			.map((g) => {
				let last = 0;
				let i = 0;
				const data = stackDates.map((date) => {
					while (i < g.growth.length && g.growth[i].date <= date) {
						last = decimalStringToNumber(g.growth[i].total_worth);
						i += 1;
					}
					return last;
				});
				return { id: g.class_id, name: g.name, data };
			})
	);
	const stacked = $derived(
		classSeries.map((s) => ({
			label: s.name,
			data: s.data,
			color: classColor.get(s.id) ?? categoryRamp[0],
			fill: true
		}))
	);

	// The distribution reads the period too: worth at its end, growth from the first worth in it.
	const trendOf = $derived(new Map(classSeries.map((s) => [s.id, s.data])));
	const growthOf = $derived(
		new Map(
			classSeries.map((s) => {
				const first = s.data.find((v) => v !== 0) ?? 0;
				const last = s.data.at(-1) ?? 0;
				return [s.id, first > 0 ? (last / first - 1) * 100 : null];
			})
		)
	);
	const distribution = $derived(
		classSeries
			.map((s) => ({
				key: s.id,
				label: s.name,
				value: s.data.at(-1) ?? 0,
				color: classColor.get(s.id)
			}))
			.filter((row) => row.value > 0)
			.sort((a, b) => b.value - a.value)
	);

	/** The table is open in the full-table slide-over. */
	let ledgerExpanded = $state(false);

	// Closing the slide-over puts focus back on the button that opened it. The page renders the
	// table, and with it that button, again only once the drawer has closed, so wait a tick.
	async function focusFullTableButton() {
		await tick();
		document.querySelector<HTMLElement>('[aria-label="Open the full table"]')?.focus();
	}

	// A drag and its Reset change the period in quick succession; only the newest load may write.
	let requestId = 0;

	async function loadAll() {
		const id = ++requestId;
		loading = true;
		error = null;
		try {
			await accountStore.ensureLoaded();
			if (id !== requestId) return;
			// No account yet is an empty state, not a failure.
			if (!accountStore.hasAccount) {
				classes = [];
				growth = [];
				return;
			}
			const [cls, series] = await Promise.all([
				listAssetClasses({}),
				getAssetClassGrowth({
					from: from || undefined,
					to: to || undefined
				})
			]);
			if (id !== requestId) return;
			classes = cls;
			growth = series;
		} catch {
			if (id !== requestId) return;
			error = 'Failed to load assets';
		} finally {
			if (id === requestId) loading = false;
		}
	}

	// Reload (and zoom the worth chart) whenever the date range changes; also the initial load.
	$effect(() => {
		void from;
		void to;
		void loadAll();
	});

	async function createClass() {
		if (className.trim() === '') return;
		creatingClass = true;
		createError = null;
		try {
			await accountStore.ensureLoaded();
			await createAssetClass({ name: className.trim() });
			createOpen = false;
			className = '';
			toast.success('Asset class created');
			void loadAll();
		} catch {
			// The dialog stays open with the entry, so the refusal is stated in it, not behind it.
			createError = 'Failed to create asset class';
		} finally {
			creatingClass = false;
		}
	}

	async function openClass(row: AssetClass) {
		drawerOpen = true;
		detailsLoading = true;
		details = null;
		classNotice = null;
		try {
			await accountStore.ensureLoaded();
			details = await getAssetClassDetails(row.id);
		} finally {
			detailsLoading = false;
		}
	}

	// A new item changes the class total, so the table behind the drawer is reloaded too.
	async function reloadOpenClass() {
		if (!details) return;
		const classId = details.class.id;
		detailsLoading = true;
		try {
			details = await getAssetClassDetails(classId);
		} catch {
			// The item was saved; only the refresh failed, so say that rather than let the
			// drawer silently keep showing the figures from before it.
			classNotice = { intent: 'error', message: 'Saved, but the class could not be refreshed' };
		} finally {
			detailsLoading = false;
		}
		void loadAll();
	}
</script>

{#snippet worthCell(row: AssetClass)}
	<Money amount={decimalStringToNumber(row.current_worth)} currency="EUR" size="sm" />
{/snippet}

{#snippet shareCell(row: AssetClass)}
	{#if !row.archived && total > 0}
		<span class="text-sm text-slate-700 tabular-nums">
			{((decimalStringToNumber(row.current_worth) / total) * 100).toFixed(1)}%
		</span>
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

<!-- The distribution list's extra cells: the class's growth and how its worth moved. -->
{#snippet distributionExtra(row: { key: string })}
	{@const pct = growthOf.get(row.key)}
	<span
		class={[
			'w-14 shrink-0 text-right text-xs tabular-nums',
			pct === null || pct === undefined
				? 'text-slate-400'
				: pct >= 0
					? 'text-emerald-700'
					: 'text-red-700'
		].join(' ')}
	>
		{pct === null || pct === undefined ? '—' : `${pct >= 0 ? '+' : ''}${pct.toFixed(1)}%`}
	</span>
	<span class="hidden w-20 shrink-0 justify-end sm:flex">
		{#if (trendOf.get(row.key)?.length ?? 0) > 1}
			<Sparkline data={trendOf.get(row.key) ?? []} width={80} height={22} />
		{/if}
	</span>
{/snippet}

{#snippet growthCell(row: AssetClass)}
	{#if row.growth_pct !== null && row.growth_pct !== undefined}
		<Badge intent={row.growth_pct >= 0 ? 'success' : 'error'} variant="soft" size="sm">
			{row.growth_pct >= 0 ? '+' : ''}{row.growth_pct.toFixed(2)}%
		</Badge>
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar />
	{/snippet}

	<PageContentTemplate title="Assets" fill={false}>
		{#snippet analytics()}
			<div class="grid grid-cols-1 gap-x-6 gap-y-3 lg:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
				<!-- The total is net worth in the masthead already; stacked by class, the chart
				     shows what moved it. -->
				<AnalyticsCard title="Worth by class">
					{#snippet actions()}
						{#if draggedRange.active}
							<Badge intent="info" variant="soft" size="sm">{periodStore.rangeLabel}</Badge>
							<Button size="sm" variant="ghost" onclick={draggedRange.reset}>Reset</Button>
						{/if}
					{/snippet}
					<ul
						class="mb-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-600"
						aria-label="Legend"
					>
						{#each stacked as series (series.label)}
							<li class="inline-flex items-center gap-1.5">
								<span class="size-2.5 rounded-sm" style:background-color={series.color}></span>
								{series.label}
							</li>
						{/each}
					</ul>
					<TimeSeriesChart
						height={chartHeights.tallUnderLegend}
						{loading}
						stacked
						labels={stackDates}
						xTickFormat={monthShort}
						enableRangeSelect
						{onRangeSelect}
						ariaLabel="Worth per asset class over time, stacked"
						datasets={stacked}
					/>
				</AnalyticsCard>
				<AnalyticsCard title="Distribution">
					<RankedBarChart
						data={distribution}
						{loading}
						formatValue={euro}
						extra={distributionExtra}
						onSelect={(key) => {
							const row = classes.find((c) => c.id === key);
							if (row) void openClass(row);
						}}
						ariaLabel="Asset classes by worth — select one to open it"
					/>
				</AnalyticsCard>
			</div>
		{/snippet}

		{#if ledgerExpanded}
			<p class="px-4 py-6 text-sm text-slate-500">The asset classes are open in the full table.</p>
		{:else}
			{@render ledger(false)}
		{/if}
	</PageContentTemplate>
</AppShellTemplate>

<Drawer
	bind:open={ledgerExpanded}
	title="Asset classes"
	width="max-w-[96vw]"
	padded={false}
	onClose={focusFullTableButton}
>
	{@render ledger(true)}
</Drawer>

<!-- The table renders in one place at a time: on the page, or in the full-table slide-over,
     which is the same toolbar and table with the whole screen to work in. -->
{#snippet ledger(full: boolean)}
	<!-- The asset-class table has never had a search, and this change does not give it one. -->
	<LedgerToolbar
		title={full ? undefined : 'Asset classes'}
		meta={loading ? 'Loading…' : error ? 'Could not load' : `${classes.length} classes`}
	>
		{#snippet actions()}
			<!-- Reset on opening, not on closing: Cancel sets `createOpen` itself, which never reaches
				     the dialog's `onClose`, and a stale band would greet the next empty form. -->
			<Button
				shape="default"
				onclick={() => {
					createError = null;
					createOpen = true;
				}}
			>
				<Icon icon="heroicons:plus" />
				Add asset class
			</Button>
			{#if !full}
				<IconButton
					icon="heroicons:arrows-pointing-out"
					ariaLabel="Open the full table"
					onclick={() => (ledgerExpanded = true)}
				/>
			{/if}
		{/snippet}
	</LedgerToolbar>
	<DataTable
		rows={classes}
		{loading}
		{error}
		emptyText="No asset classes yet. Add one to start tracking what you own."
		onRowClick={openClass}
		onRetry={() => void loadAll()}
		columns={[
			{ key: 'name', header: 'Class', value: (r: AssetClass) => r.name },
			{ key: 'source', header: 'Source', value: (r: AssetClass) => r.source },
			{ key: 'worth', header: 'Current worth', align: 'right', cell: worthCell },
			{ key: 'share', header: 'Share', align: 'right', cell: shareCell },
			{ key: 'growth', header: 'Growth', align: 'right', cell: growthCell }
		]}
	/>
{/snippet}

<AssetClassDrawer
	bind:open={drawerOpen}
	{details}
	loading={detailsLoading}
	notice={classNotice}
	onDismissNotice={() => (classNotice = null)}
	onOpenClass={() => details && goto(`/assets/${details.class.id}`)}
	onAddItem={() => (addItemOpen = true)}
	onOpenItem={(assetId) => details && goto(`/assets/${details.class.id}?item=${assetId}`)}
/>

{#if details}
	<AddAssetItemDialog
		bind:open={addItemOpen}
		classId={details.class.id}
		assetClassName={details.class.name}
		onSaved={reloadOpenClass}
	/>
{/if}

<Dialog bind:open={createOpen} title="New asset class" size="sm">
	{#if createError}
		<NoticeBand intent="error" surface="inset" gutter="dialog" class="-mx-5 -mt-4 mb-4">
			{createError}
		</NoticeBand>
	{/if}
	<FormField label="Name" id="class-name">
		{#snippet children(ctx)}
			<Input id={ctx.id} bind:value={className} placeholder="e.g. Real estate" />
		{/snippet}
	</FormField>
	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (createOpen = false)}>Cancel</Button>
		<Button intent="success" onclick={createClass} loading={creatingClass}>Create</Button>
	{/snippet}
</Dialog>
