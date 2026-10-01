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
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Switch from '$lib/components/atoms/switch/Switch.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import TransactionDateHeader from '$lib/components/molecules/transaction-date-header/TransactionDateHeader.svelte';
	import TransactionDetailDrawer from '$lib/components/organisms/transaction-detail-drawer/TransactionDetailDrawer.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import ListingSearchSelect from '$lib/components/molecules/listing-search-select/ListingSearchSelect.svelte';
	import {
		listPortfolioPositions,
		getPortfolioSnapshots,
		listPortfolioTransactions,
		createManualPortfolioTransaction,
		changePortfolioTransactionDate,
		rebuildPortfolio
	} from '$lib/services/portfolio';
	import { listVendors } from '$lib/services/vendors';
	import { accountStore } from '$lib/stores/account.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { isManualPortfolioTransaction, portfolioOriginLabel } from '$lib/api/transactions';
	import { ApiError } from '$lib/api/client';
	import { chartColors } from '$lib/charts/theme';
	import { formatDisplayDate, todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';
	import type {
		ListingSearchRow,
		PortfolioPosition,
		PortfolioRebuildOutcome,
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
	let includeClosed = $state(false);
	let tab = $state('positions');
	let from = $state('');
	let to = $state('');
	let searchQuery = $state('');

	let txOpen = $state(false);
	let vendors = $state<Vendor[]>([]);
	let vendorId = $state('');
	let txType = $state<PortfolioTransactionType>('BUY');
	let txListing = $state<ListingSearchRow | null>(null);
	let txQuantity = $state('');
	let txAmount = $state('');
	let txToday = $state(todayISO());
	let txDate = $state(todayISO());
	let txPickerOpen = $state(false);
	let txDescription = $state('');
	let txError = $state<string | null>(null);
	let creatingTx = $state(false);
	let rebuilding = $state(false);

	let detailRow = $state<PortfolioTransaction | null>(null);
	let detailOpen = $state(false);
	let detailDate = $state('');
	let savingDate = $state(false);
	let dateError = $state<string | null>(null);

	// Set when a write was saved but its rebuild could not run, so what is on screen (positions,
	// performance, net worth) still reflects the previous result.
	let staleSince = $state<string | null>(null);

	const tabs = [
		{ value: 'positions', label: 'Positions' },
		{ value: 'transactions', label: 'Transactions' }
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

	const navActions: MenuItem[] = [
		{
			label: 'Rebuild portfolio',
			icon: 'heroicons:arrow-path',
			onSelect: () => void handleRebuild()
		}
	];

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
					q: searchQuery || undefined,
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
		void searchQuery;
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
		// The form always opens on today; a backdated entry is a deliberate step away from it.
		txToday = todayISO();
		txDate = txToday;
		txPickerOpen = false;
		txDescription = '';
		txError = null;
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
		txError = null;
		try {
			await accountStore.ensureLoaded();
			const created = await createManualPortfolioTransaction({
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
			noteRebuild(created.rebuild, txDate);
			void loadAll();
		} catch (err) {
			txError = writeMessage(err, txDate);
			toast.error('Failed to add transaction');
		} finally {
			creatingTx = false;
		}
	}

	function openDetail(row: PortfolioTransaction) {
		detailRow = row;
		detailDate = row.occurred_at.slice(0, 10);
		dateError = null;
		detailOpen = true;
	}

	async function handleDateChange(date: string) {
		const row = detailRow;
		if (!row) return;
		savingDate = true;
		dateError = null;
		try {
			const result = await changePortfolioTransactionDate({ id: row.id, occurred_at: date });
			detailOpen = false;
			toast.success(`Moved to ${formatDisplayDate(date)}`);
			noteRebuild(result.rebuild, date);
			void loadAll();
		} catch (err) {
			dateError = writeMessage(err, date);
		} finally {
			savingDate = false;
		}
	}

	// A write that could not be followed by a rebuild leaves the page showing the previous
	// result, so it says so and keeps the existing rebuild action as the way out.
	function noteRebuild(outcome: PortfolioRebuildOutcome, date: string) {
		staleSince = outcome === 'completed' || outcome === 'skipped' ? null : date;
	}

	// The refusals worth naming are the ones the reader can act on: a date that already holds an
	// identical transaction, or a row that came from an import and keeps its statement date.
	function writeMessage(err: unknown, date: string): string {
		if (err instanceof ApiError && err.status === 409) {
			return `A transaction with the same listing and amount already exists on ${formatDisplayDate(date)}. Pick another day.`;
		}
		if (err instanceof ApiError && err.status === 422) {
			return 'This transaction came from an import, so it keeps its statement date.';
		}
		return 'Could not save the transaction. Try again.';
	}

	async function handleRebuild() {
		if (rebuilding) return;
		rebuilding = true;
		try {
			await accountStore.ensureLoaded();
			await rebuildPortfolio();
			staleSince = null;
			toast.success('Portfolio rebuild started');
			void loadAll();
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

<!-- Where the row came from, as a word rather than only a badge colour: it is what decides
     whether its date can still be changed. -->
{#snippet txOriginCell(row: PortfolioTransaction)}
	<Badge intent={isManualPortfolioTransaction(row) ? 'info' : 'neutral'} variant="soft" size="sm">
		{portfolioOriginLabel(row)}
	</Badge>
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Portfolio"
			showSearch
			searchValue={searchQuery}
			searchPlaceholder="Search transactions…"
			onSearch={(q) => (searchQuery = q)}
			showDateRange
			dateFrom={from || null}
			dateTo={to || null}
			onDateChange={(r) => {
				from = r.from ?? '';
				to = r.to ?? '';
			}}
			actions={navActions}
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			<div class="flex flex-col gap-3">
				<KpiRow items={kpis} columns={3} />
				<AnalyticsCard title="Value vs cost basis">
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

		{#if staleSince}
			<Alert intent="warning" title="Portfolio not updated yet">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<span>
						The transaction was saved on {formatDisplayDate(staleSince)}, but a rebuild was already
						running. Positions, performance and net worth still show the previous result.
					</span>
					<Button
						size="sm"
						variant="outline"
						intent="secondary"
						shape="default"
						loading={rebuilding}
						onclick={() => void handleRebuild()}
					>
						Rebuild portfolio
					</Button>
				</div>
			</Alert>
		{/if}

		<div class="flex min-h-0 flex-1 flex-col">
			<LedgerToolbar actionLabel="Add transaction" onAdd={openTx}>
				<Tabs {tabs} bind:value={tab} ariaLabel="Portfolio view" />
				{#if tab === 'positions'}
					<label class="flex items-center gap-2 text-sm text-slate-600">
						<span>Include closed</span>
						<Switch bind:checked={includeClosed} aria-label="Include closed positions" />
					</label>
				{/if}
			</LedgerToolbar>

			{#if tab === 'positions'}
				<DataTable
					rows={positions}
					{loading}
					{error}
					emptyText="No positions"
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
					emptyText="No transactions"
					onRowClick={openDetail}
					columns={[
						{
							key: 'date',
							header: 'Date',
							value: (r: PortfolioTransaction) => formatDisplayDate(r.occurred_at.slice(0, 10))
						},
						{ key: 'type', header: 'Type', cell: txTypeCell },
						{ key: 'origin', header: 'Entered', width: 'w-28', cell: txOriginCell },
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

<!-- While saving, Escape and the backdrop stay inert so a pending create cannot be sent twice. -->
<Dialog
	bind:open={txOpen}
	title="New transaction"
	size="md"
	closeOnEscape={!creatingTx}
	closeOnBackdrop={!creatingTx}
>
	<div class="space-y-3">
		<TransactionDateHeader
			bind:value={txDate}
			bind:open={txPickerOpen}
			today={txToday}
			disabled={creatingTx}
		/>

		{#if txError}
			<!-- role="alert" so a refusal that appears after Save is announced, not just coloured. -->
			<div role="alert">
				<Alert intent="error" title="This transaction was not saved">{txError}</Alert>
			</div>
		{/if}

		<FormField label="Type" id="ptx-type">
			{#snippet children(ctx)}
				<Select id={ctx.id} bind:value={txType} options={txTypeOptions} ariaLabel="Transaction type" />
			{/snippet}
		</FormField>

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
		<Button variant="ghost" intent="secondary" disabled={creatingTx} onclick={() => (txOpen = false)}>
			Cancel
		</Button>
		<Button intent="success" onclick={submitTx} loading={creatingTx}>
			{creatingTx ? 'Saving' : 'Save'}
		</Button>
	{/snippet}
</Dialog>

{#snippet detailFields()}
	{#if detailRow}
		<div class="flex items-center justify-between gap-3 py-3">
			<dt class="text-sm text-slate-500">Amount</dt>
			<dd class="text-sm text-slate-800">{detailRow.amount}</dd>
		</div>
		<div class="flex items-center justify-between gap-3 py-3">
			<dt class="text-sm text-slate-500">Quantity</dt>
			<dd class="text-sm text-slate-800">{detailRow.quantity}</dd>
		</div>
	{/if}
{/snippet}

{#if detailRow}
	<TransactionDetailDrawer
		bind:open={detailOpen}
		title={detailRow.symbol ?? detailRow.description ?? 'Transaction'}
		originLabel={portfolioOriginLabel(detailRow)}
		subtitle={detailRow.type}
		editable={isManualPortfolioTransaction(detailRow)}
		bind:date={detailDate}
		originalDate={detailRow.occurred_at.slice(0, 10)}
		saving={savingDate}
		error={dateError}
		onSave={handleDateChange}
		details={detailFields}
	/>
{/if}
