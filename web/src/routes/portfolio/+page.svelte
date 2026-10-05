<script lang="ts">
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import ListingSearchSelect from '$lib/components/molecules/listing-search-select/ListingSearchSelect.svelte';
	import {
		listPortfolioPositions,
		getPortfolioSnapshots,
		listPortfolioTransactions,
		createManualPortfolioTransaction,
		rebuildPortfolio
	} from '$lib/services/portfolio';
	import { listVendors } from '$lib/services/vendors';
	import { accountStore } from '$lib/stores/account.svelte';
	import { periodStore } from '$lib/stores/period.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { chartColors } from '$lib/charts/theme';
	import { formatDisplayDate, todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';
	import type {
		ListingSearchRow,
		PortfolioPosition,
		PortfolioSnapshotPoint,
		PortfolioTransaction,
		PortfolioTransactionType,
		Vendor
	} from '$lib/api/types';

	let positions = $state<PortfolioPosition[]>([]);
	let snapshots = $state<PortfolioSnapshotPoint[]>([]);
	let transactions = $state<PortfolioTransaction[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let tab = $state('positions');
	let searchQuery = $state('');

	/**
	 * The row filter that replaced the "Include closed" switch. `include_closed` is the only
	 * thing the API offers, so `Closed` asks for everything and keeps the closed rows here --
	 * the positions list is not paginated, so narrowing it in the browser costs nothing.
	 */
	let positionStatus = $state('open');
	const includeClosed = $derived(positionStatus !== 'open');

	// One period for the whole app, chosen in the account overview.
	const from = $derived(periodStore.from);
	const to = $derived(periodStore.to);

	let txOpen = $state(false);
	let vendors = $state<Vendor[]>([]);
	let vendorId = $state('');
	let txType = $state<PortfolioTransactionType>('BUY');
	let txListing = $state<ListingSearchRow | null>(null);
	let txQuantity = $state('');
	let txAmount = $state('');
	let txDate = $state(todayISO());
	let txDescription = $state('');
	let creatingTx = $state(false);
	let rebuilding = $state(false);

	const tabs = [
		{ value: 'positions', label: 'Positions' },
		{ value: 'transactions', label: 'Transactions' }
	];

	const statusTabs = [
		{ value: 'open', label: 'Open' },
		{ value: 'closed', label: 'Closed' },
		{ value: 'all', label: 'All' }
	];

	const txTypeOptions = [
		{ value: 'BUY', label: 'Buy' },
		{ value: 'SELL', label: 'Sell' },
		{ value: 'DIVIDEND', label: 'Dividend' },
		{ value: 'TAX', label: 'Tax' },
		{ value: 'FEE', label: 'Fee' },
		{ value: 'CASH', label: 'Cash' }
	];

	// Backend rules: non-CASH types require a listing; BUY/SELL additionally require a quantity.
	const needsListing = $derived(txType !== 'CASH');
	const needsQuantity = $derived(txType === 'BUY' || txType === 'SELL');
	const vendorOptions = $derived(vendors.map((v) => ({ value: v.id, label: v.name })));

	const monthShort = (iso: string) =>
		new Date(iso).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	const latest = $derived(snapshots.at(-1) ?? null);
	const kpis = $derived<KpiItem[]>(
		latest
			? [
					{
						label: 'Market value',
						amount: scaledToNumber(latest.market_value),
						currency: 'EUR',
						change: latest.total_pnl_pct
					},
					{
						label: 'Total P&L',
						amount: scaledToNumber(latest.total_pnl),
						currency: 'EUR',
						change: latest.return_vs_cost_basis_pct
					},
					{ label: 'Cost basis', amount: scaledToNumber(latest.total_cost_basis), currency: 'EUR' }
				]
			: []
	);

	/**
	 * One search box serves both views, so what it means follows the tab. Transactions are
	 * searched by the API; positions come back as one unpaginated list, so the same words
	 * narrow them here without a round trip.
	 */
	const transactionQuery = $derived(tab === 'transactions' ? searchQuery.trim() : '');

	const visiblePositions = $derived.by(() => {
		const needle = tab === 'positions' ? searchQuery.trim().toLowerCase() : '';
		return positions.filter((position) => {
			if (positionStatus === 'open' && position.is_closed) return false;
			if (positionStatus === 'closed' && !position.is_closed) return false;
			if (!needle) return true;
			return (
				(position.symbol ?? '').toLowerCase().includes(needle) ||
				(position.name ?? '').toLowerCase().includes(needle)
			);
		});
	});

	const tableMeta = $derived.by(() => {
		if (loading) return 'Loading…';
		if (error) return 'Could not load';
		if (tab === 'positions') {
			return `${visiblePositions.length} ${visiblePositions.length === 1 ? 'position' : 'positions'}`;
		}
		return `${transactions.length} ${transactions.length === 1 ? 'row' : 'rows'}`;
	});

	// An empty list after a search is a different situation from an account with no positions.
	const positionsEmptyText = $derived(
		searchQuery.trim()
			? `No positions match “${searchQuery.trim()}”. Clear the search to see them all.`
			: positionStatus === 'closed'
				? 'No closed positions.'
				: 'No positions yet. Add a transaction or import a file to build them.'
	);

	async function loadPositions() {
		await accountStore.ensureLoaded();
		if (!accountStore.hasAccount) {
			positions = [];
			return;
		}
		positions = (
			await listPortfolioPositions({
				include_closed: includeClosed
			})
		).data;
	}

	async function loadAll() {
		loading = true;
		error = null;
		try {
			await accountStore.ensureLoaded();
			// No account yet is an empty state, not a failure: firing account-scoped
			// requests with the placeholder id would 4xx and read as a load error.
			if (!accountStore.hasAccount) {
				snapshots = [];
				transactions = [];
				positions = [];
				return;
			}
			const [snaps, txs] = await Promise.all([
				getPortfolioSnapshots({
					from: from || undefined,
					to: to || undefined
				}),
				listPortfolioTransactions({
					limit: 25,
					q: transactionQuery || undefined,
					from: from || undefined,
					to: to || undefined
				})
			]);
			snapshots = snaps;
			transactions = txs.data;
			await loadPositions();
		} catch {
			error = 'Failed to load portfolio';
		} finally {
			loading = false;
		}
	}

	// Reload (zooming the value chart, filtering transactions) on date/search change; also initial load.
	$effect(() => {
		void from;
		void to;
		void transactionQuery;
		void loadAll();
	});

	$effect(() => {
		// Reads `includeClosed` synchronously, so it reloads positions when the toggle changes.
		void loadPositions();
	});

	async function openTx() {
		txType = 'BUY';
		txListing = null;
		txQuantity = '';
		txAmount = '';
		txDate = todayISO();
		txDescription = '';
		txOpen = true;
		if (vendors.length === 0) {
			try {
				const all = await listVendors();
				// Manual portfolio transactions require a brokerage/portfolio vendor.
				vendors = all.filter((v) => v.active && v.type === 'portfolio');
				if (vendors.length > 0) vendorId = vendors[0].id;
			} catch {
				// Leave the vendor list empty; the form will warn on submit.
			}
		}
	}

	async function submitTx() {
		if (!vendorId || txAmount.trim() === '') {
			toast.error('Vendor and amount are required');
			return;
		}
		if (needsListing && !txListing) {
			toast.error('Select a listing');
			return;
		}
		if (needsQuantity && txQuantity.trim() === '') {
			toast.error('Quantity is required');
			return;
		}
		creatingTx = true;
		try {
			await accountStore.ensureLoaded();
			await createManualPortfolioTransaction({
				vendor_id: vendorId,
				occurred_at: txDate,
				type: txType,
				listing_id: needsListing && txListing ? txListing.id : undefined,
				amount: txAmount.trim(),
				quantity: needsQuantity ? txQuantity.trim() : undefined,
				description: txDescription.trim() || undefined
			});
			txOpen = false;
			toast.success('Transaction added');
			void loadAll();
		} catch {
			toast.error('Failed to add transaction');
		} finally {
			creatingTx = false;
		}
	}

	async function handleRebuild() {
		if (rebuilding) return;
		rebuilding = true;
		try {
			await accountStore.ensureLoaded();
			await rebuildPortfolio();
			toast.success('Portfolio rebuild started');
		} catch {
			toast.error('Failed to start rebuild');
		} finally {
			rebuilding = false;
		}
	}
</script>

{#snippet marketValueCell(row: PortfolioPosition)}
	{#if row.market_value !== null && row.market_value !== undefined}
		<Money amount={scaledToNumber(row.market_value)} currency="EUR" size="sm" />
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

{#snippet pnlCell(row: PortfolioPosition)}
	{#if row.unrealized_pnl_pct !== null && row.unrealized_pnl_pct !== undefined}
		<Badge intent={row.unrealized_pnl_pct >= 0 ? 'success' : 'error'} variant="soft" size="sm">
			{row.unrealized_pnl_pct >= 0 ? '+' : ''}{row.unrealized_pnl_pct.toFixed(2)}%
		</Badge>
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

{#snippet txTypeCell(row: PortfolioTransaction)}
	<Badge intent="neutral" variant="soft" size="sm">{row.type}</Badge>
{/snippet}

{#snippet viewTabs()}
	<Tabs {tabs} bind:value={tab} size="sm" ariaLabel="Portfolio view" />
{/snippet}

{#snippet statusFilter()}
	<Tabs tabs={statusTabs} bind:value={positionStatus} size="sm" ariaLabel="Position status" />
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar />
	{/snippet}

	<PageContentTemplate title="Portfolio">
		{#snippet analytics()}
			<div class="flex flex-col gap-3">
				<KpiRow items={kpis} columns={3} />
				<!-- Rebuilding recomputes the series this chart draws, so it hangs on the chart's
				     own rule rather than in a menu above the page. -->
				<AnalyticsCard title="Value vs cost basis">
					{#snippet actions()}
						<Button variant="ruled" loading={rebuilding} onclick={handleRebuild}>
							<Icon icon="heroicons:arrow-path" />
							Rebuild portfolio
						</Button>
					{/snippet}
					<TimeSeriesChart
						height="h-52"
						{loading}
						labels={snapshots.map((s) => s.occurred_at.slice(0, 10))}
						xTickFormat={monthShort}
						datasets={[
							{
								label: 'Market value',
								data: snapshots.map((s) => scaledToNumber(s.market_value)),
								color: chartColors.positive,
								fill: true
							},
							{
								label: 'Cost basis',
								data: snapshots.map((s) => scaledToNumber(s.total_cost_basis)),
								color: chartColors.net,
								dashed: true
							}
						]}
					/>
				</AnalyticsCard>
			</div>
		{/snippet}

		<div class="flex min-h-0 flex-1 flex-col">
			<!-- The status filter belongs to the positions, so it is absent on the other tab. -->
			<LedgerToolbar
				title={tab === 'positions' ? 'Positions' : 'Transactions'}
				meta={tableMeta}
				showSearch
				searchValue={searchQuery}
				onSearch={(q) => (searchQuery = q)}
				searchPlaceholder={tab === 'positions'
					? 'Search symbol or name…'
					: 'Search transactions…'}
				searchAriaLabel={tab === 'positions' ? 'Search positions' : 'Search transactions'}
				before={viewTabs}
				filters={tab === 'positions' ? statusFilter : undefined}
			>
				{#snippet actions()}
					<Button shape="default" onclick={openTx}>
						<Icon icon="heroicons:plus" />
						Add transaction
					</Button>
				{/snippet}
			</LedgerToolbar>

			{#if tab === 'positions'}
				<DataTable
					rows={visiblePositions}
					{loading}
					{error}
					emptyText={positionsEmptyText}
					columns={[
						{ key: 'symbol', header: 'Symbol', value: (r: PortfolioPosition) => r.symbol ?? '—' },
						{ key: 'name', header: 'Name', value: (r: PortfolioPosition) => r.name ?? '—' },
						{
							key: 'quantity',
							header: 'Qty',
							align: 'right',
							value: (r: PortfolioPosition) => r.quantity
						},
						{ key: 'market_value', header: 'Market value', align: 'right', cell: marketValueCell },
						{ key: 'pnl', header: 'Unrealized', align: 'right', cell: pnlCell }
					]}
				/>
			{:else}
				<DataTable
					rows={transactions}
					{loading}
					{error}
					emptyText={transactionQuery
						? `No transactions match “${transactionQuery}”. Clear the search to see them all.`
						: 'No transactions in this period.'}
					columns={[
						{
							key: 'date',
							header: 'Date',
							value: (r: PortfolioTransaction) => formatDisplayDate(r.occurred_at.slice(0, 10))
						},
						{ key: 'type', header: 'Type', cell: txTypeCell },
						{
							key: 'symbol',
							header: 'Symbol',
							value: (r: PortfolioTransaction) => r.symbol ?? '—'
						},
						{
							key: 'quantity',
							header: 'Qty',
							align: 'right',
							value: (r: PortfolioTransaction) => r.quantity
						},
						{
							key: 'amount',
							header: 'Amount',
							align: 'right',
							value: (r: PortfolioTransaction) => r.amount
						}
					]}
				/>
			{/if}
		</div>
	</PageContentTemplate>
</AppShellTemplate>

<Dialog bind:open={txOpen} title="New transaction" size="md">
	<div class="space-y-3">
		<div class="grid grid-cols-2 gap-3">
			<FormField label="Date" id="ptx-date">
				<DatePicker value={txDate} onChange={(v) => (txDate = v ?? txDate)} class="w-full" />
			</FormField>
			<FormField label="Type" id="ptx-type">
				{#snippet children(ctx)}
					<Select
						id={ctx.id}
						bind:value={txType}
						options={txTypeOptions}
						ariaLabel="Transaction type"
					/>
				{/snippet}
			</FormField>
		</div>

		<FormField label="Vendor" id="ptx-vendor" hint="Brokerage account">
			{#snippet children(ctx)}
				<Select
					id={ctx.id}
					bind:value={vendorId}
					options={vendorOptions}
					placeholder="Select a vendor"
					ariaLabel="Vendor"
				/>
			{/snippet}
		</FormField>

		{#if needsListing}
			<FormField label="Listing" id="ptx-listing">
				<ListingSearchSelect bind:value={txListing} ariaLabel="Listing" />
			</FormField>
		{/if}

		<div class="grid grid-cols-2 gap-3">
			{#if needsQuantity}
				<FormField label="Quantity" id="ptx-quantity">
					{#snippet children(ctx)}
						<Input id={ctx.id} bind:value={txQuantity} placeholder="e.g. 10" />
					{/snippet}
				</FormField>
			{/if}
			<FormField label="Amount" id="ptx-amount">
				{#snippet children(ctx)}
					<CurrencyInput id={ctx.id} bind:value={txAmount} ariaDescribedby={ctx.describedby} />
				{/snippet}
			</FormField>
		</div>

		<FormField label="Description" id="ptx-description" hint="Optional">
			{#snippet children(ctx)}
				<Input id={ctx.id} bind:value={txDescription} />
			{/snippet}
		</FormField>
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (txOpen = false)}>Cancel</Button>
		<Button intent="success" onclick={submitTx} loading={creatingTx}>Save</Button>
	{/snippet}
</Dialog>
