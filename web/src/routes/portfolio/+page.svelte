<script lang="ts">
	import { onMount, tick } from 'svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import { chartHeights } from '$lib/components/organisms/charts/charts.variants';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import RankedBarChart from '$lib/components/organisms/charts/RankedBarChart.svelte';
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
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import TransactionDateHeader from '$lib/components/molecules/transaction-date-header/TransactionDateHeader.svelte';
	import PortfolioTransactionDrawer from '$lib/components/organisms/portfolio-transaction-drawer/PortfolioTransactionDrawer.svelte';
	import FooterBar from '$lib/components/organisms/footer-bar/FooterBar.svelte';
	import {
		portfolioTransactionTypeOptions,
		typeNeedsListing,
		typeNeedsQuantity
	} from '$lib/components/organisms/portfolio-transaction-drawer/portfolio-transaction-drawer.types';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import ListingSearchSelect from '$lib/components/molecules/listing-search-select/ListingSearchSelect.svelte';
	import ImportDialog from '$lib/components/organisms/import-dialog/ImportDialog.svelte';
	import {
		listPortfolioPositions,
		getPortfolioSnapshots,
		listPortfolioTransactions,
		createManualPortfolioTransaction,
		updatePortfolioTransaction,
		deletePortfolioTransaction,
		rebuildPortfolio
	} from '$lib/services/portfolio';
	import { listVendors } from '$lib/services/vendors';
	import { connectRealtime } from '$lib/services/realtime';
	import { accountStore } from '$lib/stores/account.svelte';
	import { periodStore } from '$lib/stores/period.svelte';
	import { createDraggedRange } from '$lib/stores/dragged-range.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { isManualPortfolioTransaction, portfolioOriginLabel } from '$lib/api/transactions';
	import { ApiError } from '$lib/api/client';
	import { chartColors } from '$lib/charts/theme';
	import { formatDisplayDate, todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';
	import type {
		ListingSearchRow,
		PortfolioPosition,
		PortfolioRebuildOutcome,
		PortfolioSnapshotPoint,
		PortfolioTransaction,
		PortfolioTransactionType,
		PortfolioTransactionsQuery,
		UpdatePortfolioTransactionRequest,
		Vendor
	} from '$lib/api/types';

	let positions = $state<PortfolioPosition[]>([]);
	let snapshots = $state<PortfolioSnapshotPoint[]>([]);
	let transactions = $state<PortfolioTransaction[]>([]);
	let txLimit = $state(25);
	let txOffset = $state(0);
	let txTotal = $state(0);
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
	let txToday = $state(todayISO());
	let txDate = $state(todayISO());
	let txPickerOpen = $state(false);
	let txDescription = $state('');
	let txError = $state<string | null>(null);
	/**
	 * Only a refusal by the server gets the heading: it states that a save was attempted and failed.
	 * A validation message is said on its own, because nothing was sent yet.
	 */
	let txErrorTitle = $state<string | undefined>(undefined);
	let creatingTx = $state(false);
	let rebuilding = $state(false);
	let importOpen = $state(false);

	let detailRow = $state<PortfolioTransaction | null>(null);
	let detailOpen = $state(false);
	// Bumped on every open, so reopening the same row gives a fresh form rather than one
	// carrying discarded edits or the values from before the last save.
	let detailOpenCount = $state(0);
	let savingDetail = $state(false);
	let deletingDetail = $state(false);
	let detailError = $state<string | null>(null);
	// Save and Delete report in the same band, so the heading says which of the two was refused.
	let detailErrorTitle = $state('This transaction was not changed');

	// Set while the background rebuild a write started is running, so what is on screen
	// (positions, performance, net worth) is known to still be the previous result. Cleared
	// by the portfolio.rebuilt event that refreshes the page.
	let updating = $state(false);
	// When the last portfolio.rebuilt event arrived. A small account can finish rebuilding
	// before the write's response is handled, and that write must not then raise the banner.
	let lastRebuiltAt = 0;

	const tabs = [
		{ value: 'positions', label: 'Positions' },
		{ value: 'transactions', label: 'Transactions' }
	];

	const statusTabs = [
		{ value: 'open', label: 'Open' },
		{ value: 'closed', label: 'Closed' },
		{ value: 'all', label: 'All' }
	];

	const txTypeOptions = [...portfolioTransactionTypeOptions];

	// Backend rules: non-CASH types require a listing; BUY/SELL additionally require a quantity.
	const needsListing = $derived(typeNeedsListing(txType));
	const needsQuantity = $derived(typeNeedsQuantity(txType));
	const vendorOptions = $derived(vendors.map((v) => ({ value: v.id, label: v.name })));

	const monthShort = (iso: string) =>
		new Date(iso).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	const latest = $derived(snapshots.at(-1) ?? null);
	const first = $derived(snapshots[0] ?? null);
	// The cost basis is on the chart already; what moved in the period you picked is not.
	const periodChange = $derived(
		latest && first ? scaledToNumber(latest.market_value) - scaledToNumber(first.market_value) : 0
	);
	const kpis = $derived<KpiItem[]>(
		latest && first
			? [
					{
						label: 'Market value',
						amount: scaledToNumber(latest.market_value),
						currency: 'EUR',
						change: latest.total_pnl_pct
					},
					{
						label: 'Change in period',
						amount: periodChange,
						currency: 'EUR',
						change:
							first.market_value > 0
								? (periodChange / scaledToNumber(first.market_value)) * 100
								: undefined
					},
					{
						label: 'Total P&L',
						amount: scaledToNumber(latest.total_pnl),
						currency: 'EUR',
						change: latest.return_vs_cost_basis_pct
					}
				]
			: []
	);

	/** The chart shows money (value over cost, the gain shaded) or the time-weighted return. */
	let chartMode = $state<'value' | 'return'>('value');
	const chartModes = [
		{ value: 'value', label: '€ value' },
		{ value: 'return', label: '% return' }
	];
	// Dragging across the chart zooms the period to that stretch; a click alone keeps it.
	const draggedRange = createDraggedRange();
	function onRangeSelect(rangeFrom: string, rangeTo: string) {
		if (rangeFrom !== rangeTo) draggedRange.select(rangeFrom, rangeTo);
	}

	// `time_weighted_return_pct` is each snapshot's market return against the one before it, with
	// money put in or taken out removed. Chained, it is the growth since the first day shown.
	const dailyReturns = $derived(snapshots.map((s) => s.time_weighted_return_pct));
	const cumulativeReturns = $derived.by(() => {
		let growth = 1;
		return dailyReturns.map((r, i) => {
			if (i > 0) growth *= 1 + r / 100;
			return (growth - 1) * 100;
		});
	});

	// Weight is measured against the open positions only: a closed one holds no money.
	const openValue = $derived(
		positions.reduce((sum, p) => sum + (p.is_closed ? 0 : scaledToNumber(p.market_value ?? 0)), 0)
	);
	const allocation = $derived(
		positions
			.filter((p) => !p.is_closed && (p.market_value ?? 0) > 0)
			.map((p) => ({
				key: p.id,
				label: p.symbol ?? p.name ?? '—',
				value: scaledToNumber(p.market_value ?? 0)
			}))
	);
	const euro = (n: number) =>
		n.toLocaleString('en', { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 });

	/** The table is open in the full-table slide-over. */
	let ledgerExpanded = $state(false);

	// Closing the slide-over puts focus back on the button that opened it. The page renders the
	// table, and with it that button, again only once the drawer has closed, so wait a tick.
	async function focusFullTableButton() {
		await tick();
		document.querySelector<HTMLElement>('[aria-label="Open the full table"]')?.focus();
	}
	const ledgerTitle = $derived(tab === 'positions' ? 'Positions' : 'Transactions');

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
		return `${txTotal} ${txTotal === 1 ? 'row' : 'rows'}`;
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
			const [snaps] = await Promise.all([
				getPortfolioSnapshots({
					from: from || undefined,
					to: to || undefined
				}),
				loadTransactions()
			]);
			snapshots = snaps;
			await loadPositions();
		} catch {
			error = 'Failed to load portfolio';
		} finally {
			loading = false;
		}
	}

	// One page of transactions. Paging calls this on its own, so turning a page does not
	// reload the chart and positions above it.
	let transactionsRequest = 0;
	async function loadTransactions(): Promise<void> {
		// Paging and a full reload can overlap; only the latest request may write the table.
		const request = ++transactionsRequest;
		const result = await listPortfolioTransactions({
			// FooterBar only offers the page sizes the API accepts.
			limit: txLimit as PortfolioTransactionsQuery['limit'],
			offset: txOffset,
			q: transactionQuery || undefined,
			from: from || undefined,
			to: to || undefined
		});
		if (request !== transactionsRequest) return;
		txTotal = result.pagination.total;
		// Deleting the last row of the last page leaves the offset past the end; step back to
		// the page that now holds the final rows instead of showing an empty table.
		const lastPageOffset = Math.max(0, Math.floor((txTotal - 1) / txLimit) * txLimit);
		if (result.data.length === 0 && txOffset > lastPageOffset) {
			txOffset = lastPageOffset;
			return loadTransactions();
		}
		transactions = result.data;
	}

	async function changeTransactionsPage() {
		try {
			await loadTransactions();
		} catch {
			toast.error('Failed to load transactions');
		}
	}

	// Reload (zooming the value chart, filtering transactions) on date/search change; also initial
	// load. A different period or search starts the transactions from their first page.
	$effect(() => {
		void from;
		void to;
		void transactionQuery;
		txOffset = 0;
		void loadAll();
	});

	$effect(() => {
		// `loadPositions` reads `includeClosed` only after an await, where the effect no longer
		// tracks it, so it is read here to reload positions when the status filter changes.
		void includeClosed;
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
		txErrorTitle = undefined;
		txOpen = true;
		await loadVendors();
	}

	// Manual portfolio transactions and broker imports both need a brokerage vendor.
	async function loadVendors() {
		if (vendors.length > 0) return;
		try {
			const all = await listVendors();
			vendors = all.filter((v) => v.active && v.type === 'brokerage');
			if (vendors.length > 0) vendorId = vendors[0].id;
		} catch {
			// Leave the vendor list empty; the form will warn on submit.
		}
	}

	async function openImport() {
		await loadVendors();
		importOpen = true;
	}

	async function submitTx() {
		// The dialog is open, so what is missing is said in the dialog rather than on a toast behind it.
		// Nothing has been sent at this point, so these carry no "was not saved" heading.
		txErrorTitle = undefined;
		if (!vendorId || txAmount.trim() === '') {
			txError = 'Vendor and amount are required';
			return;
		}
		if (needsListing && !txListing) {
			txError = 'Select a listing';
			return;
		}
		if (needsQuantity && txQuantity.trim() === '') {
			txError = 'Quantity is required';
			return;
		}
		creatingTx = true;
		txError = null;
		try {
			await accountStore.ensureLoaded();
			const sentAt = Date.now();
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
			noteRebuild(created.rebuild, sentAt);
			void loadAll();
		} catch (err) {
			// The dialog stays open with the entry, so the refusal belongs in it, not on a toast as well.
			txError = createMessage(err, txDate);
			txErrorTitle = 'This transaction was not saved';
		} finally {
			creatingTx = false;
		}
	}

	function openDetail(row: PortfolioTransaction) {
		detailRow = row;
		detailError = null;
		detailOpenCount += 1;
		detailOpen = true;
	}

	// A 404 naming the transaction (not the listing) means it was removed elsewhere.
	function isMissingTransaction(err: unknown): boolean {
		return (
			err instanceof ApiError &&
			err.status === 404 &&
			typeof err.body === 'object' &&
			err.body !== null &&
			'transaction_id' in err.body
		);
	}

	async function handleDetailSave(body: UpdatePortfolioTransactionRequest) {
		const row = detailRow;
		if (!row) return;
		savingDetail = true;
		detailError = null;
		try {
			const sentAt = Date.now();
			const result = await updatePortfolioTransaction(row.id, body);
			detailOpen = false;
			toast.success('Transaction updated');
			noteRebuild(result.rebuild, sentAt);
			void loadAll();
		} catch (err) {
			detailError = editMessage(err, body.occurred_at, genericWriteMessage);
			detailErrorTitle = 'This transaction was not changed';
			if (isMissingTransaction(err)) void loadAll();
		} finally {
			savingDetail = false;
		}
	}

	async function handleDetailDelete() {
		const row = detailRow;
		if (!row) return;
		const day = row.occurred_at.slice(0, 10);
		deletingDetail = true;
		detailError = null;
		try {
			const sentAt = Date.now();
			const result = await deletePortfolioTransaction(row.id);
			detailOpen = false;
			toast.success('Transaction deleted');
			noteRebuild(result.rebuild, sentAt);
			void loadAll();
		} catch (err) {
			detailError = editMessage(err, day, genericDeleteMessage);
			detailErrorTitle = 'This transaction was not deleted';
			if (isMissingTransaction(err)) void loadAll();
		} finally {
			deletingDetail = false;
		}
	}

	// A write answers before its rebuild runs, so until the rebuild reports back the page
	// says that positions, performance and net worth are still being brought up to date.
	function noteRebuild(outcome: PortfolioRebuildOutcome, sentAt: number) {
		updating = outcome === 'started' && lastRebuiltAt < sentAt;
	}

	// A finished rebuild, whoever started it, is announced over the account socket; that is
	// when the page catches up with the writes made on it.
	onMount(() => {
		let realtime: { disconnect: () => void } | null = null;
		void accountStore.ensureLoaded().then(() => {
			if (!accountStore.hasAccount) return;
			realtime = connectRealtime({
				accountId: accountStore.activeId,
				events: ['portfolio.rebuilt'],
				onEvent: () => {
					lastRebuiltAt = Date.now();
				},
				onRefresh: () => {
					updating = false;
					void loadAll();
				}
			});
		});
		return () => realtime?.disconnect();
	});

	const genericWriteMessage = 'Could not save the transaction. Try again.';
	const genericDeleteMessage = 'Could not delete the transaction. Try again.';

	function duplicateMessage(date: string): string {
		return `A transaction with the same listing and amount already exists on ${formatDisplayDate(date)}. Pick another day.`;
	}

	// A refused write answers with one field and its reason, e.g. a vendor that is not a broker
	// or a date in the future. That reason is the only thing that tells the reader what to change.
	function fieldMessage(err: unknown): string | null {
		if (!(err instanceof ApiError) || typeof err.body !== 'object' || err.body === null) {
			return null;
		}
		const reason = Object.values(err.body as Record<string, unknown>).find(
			(value) => typeof value === 'string' && value.trim() !== ''
		);
		return typeof reason === 'string' ? reason : null;
	}

	// The two paths share statuses but not their meaning: a 422 on the create is an unsupported
	// vendor type, on an edit or delete it is a row that came from an import.
	function createMessage(err: unknown, date: string): string {
		if (err instanceof ApiError && err.status === 409) return duplicateMessage(date);
		if (err instanceof ApiError && (err.status === 400 || err.status === 422)) {
			return fieldMessage(err) ?? genericWriteMessage;
		}
		return genericWriteMessage;
	}

	// An edit and a delete share these statuses but not the sentence that is left when none of them
	// matches, so the caller hands in the one that names the button it was pressed on.
	function editMessage(err: unknown, date: string, generic: string): string {
		if (err instanceof ApiError && err.status === 409) return duplicateMessage(date);
		if (err instanceof ApiError && err.status === 422) {
			return 'This transaction came from an import, so it stays as the statement records it.';
		}
		if (isMissingTransaction(err)) {
			return 'This transaction no longer exists; the list has been refreshed.';
		}
		if (err instanceof ApiError && err.status === 404) {
			return 'That listing no longer exists. Pick another one.';
		}
		if (err instanceof ApiError && err.status === 400) {
			return fieldMessage(err) ?? generic;
		}
		return generic;
	}

	async function handleRebuild() {
		if (rebuilding) return;
		rebuilding = true;
		try {
			await accountStore.ensureLoaded();
			await rebuildPortfolio();
			updating = false;
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

{#snippet costBasisCell(row: PortfolioPosition)}
	{#if row.is_closed}
		<span class="text-slate-500">—</span>
	{:else}
		<Money amount={scaledToNumber(row.cost_basis)} currency="EUR" size="sm" />
	{/if}
{/snippet}

<!-- Unrealised P&L in money; a closed position shows what it realised instead. -->
{#snippet pnlAmountCell(row: PortfolioPosition)}
	{#if row.is_closed}
		<Money
			amount={scaledToNumber(row.realized_pnl)}
			currency="EUR"
			size="sm"
			colored
			signDisplay="exceptZero"
		/>
	{:else if row.market_value !== null && row.market_value !== undefined}
		<Money
			amount={scaledToNumber(row.market_value) - scaledToNumber(row.cost_basis)}
			currency="EUR"
			size="sm"
			colored
			signDisplay="exceptZero"
		/>
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

{#snippet weightCell(row: PortfolioPosition)}
	{#if !row.is_closed && row.market_value && openValue > 0}
		<span class="text-sm text-slate-700 tabular-nums">
			{((scaledToNumber(row.market_value) / openValue) * 100).toFixed(1)}%
		</span>
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

	<PageContentTemplate title="Portfolio" fill={false}>
		{#snippet analytics()}
			<div class="flex flex-col gap-3">
				<KpiRow items={kpis} columns={3} />
				<!-- Rebuilding recomputes the series this chart draws, so it hangs on the chart's
				     own rule rather than in a menu above the page. -->
				<div class="grid gap-x-6 gap-y-3 lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
					<AnalyticsCard title={chartMode === 'value' ? 'Value vs cost basis' : 'Return'}>
						{#snippet actions()}
							{#if draggedRange.active}
								<Badge intent="info" variant="soft" size="sm">{periodStore.rangeLabel}</Badge>
								<Button size="sm" variant="ghost" onclick={draggedRange.reset}>Reset</Button>
							{/if}
							<Tabs tabs={chartModes} bind:value={chartMode} size="sm" ariaLabel="Chart shows" />
							<Button variant="ruled" loading={rebuilding} onclick={handleRebuild}>
								<Icon icon="heroicons:arrow-path" />
								Rebuild portfolio
							</Button>
						{/snippet}
						{#if chartMode === 'return'}
							<ul class="mb-2 flex gap-4 text-xs text-slate-600" aria-label="Legend">
								<li class="inline-flex items-center gap-1.5">
									<span class="w-3 border-t-2 border-sky-600"></span>Growth over the period
								</li>
								<li class="inline-flex items-center gap-1.5">
									<span class="size-2.5 rounded-sm bg-emerald-600"></span><span
										class="-ml-1 size-2.5 rounded-sm bg-red-600"
									></span>Change that day
								</li>
							</ul>
						{/if}
						<!-- The gap between value and cost is the gain, so it is shaded: green above
						     the cost line, red below it. -->
						<TimeSeriesChart
							height={chartHeights.tallUnderKpis}
							{loading}
							labels={snapshots.map((s) => s.occurred_at.slice(0, 10))}
							xTickFormat={monthShort}
							enableRangeSelect
							{onRangeSelect}
							ariaLabel={chartMode === 'value'
								? 'Market value against cost basis, the gain shaded between them'
								: 'Growth over the period as a line, each day’s change as bars'}
							datasets={chartMode === 'value'
								? [
										{
											label: 'Cost basis',
											data: snapshots.map((s) => scaledToNumber(s.total_cost_basis)),
											color: chartColors.net,
											dashed: true
										},
										{
											label: 'Market value',
											data: snapshots.map((s) => scaledToNumber(s.market_value)),
											color: chartColors.positive,
											gapFill: true
										}
									]
								: [
										{
											label: 'Growth %',
											data: cumulativeReturns,
											color: chartColors.info,
											signed: true
										},
										{
											label: 'Change that day %',
											data: dailyReturns,
											color: chartColors.net,
											barColors: dailyReturns.map((r) =>
												r < 0 ? chartColors.negative : chartColors.positive
											),
											type: 'bar'
										}
									]}
						/>
					</AnalyticsCard>
					<AnalyticsCard title="Allocation">
						<RankedBarChart
							data={allocation}
							{loading}
							formatValue={euro}
							ariaLabel="Open positions by market value, with their weight"
						/>
					</AnalyticsCard>
				</div>
			</div>
		{/snippet}

		{#if updating}
			<NoticeBand intent="info" gutter="panel" title="Updating your portfolio">
				The transaction is saved. Positions, performance and net worth are being recalculated, and
				this page refreshes as soon as that is done.
			</NoticeBand>
		{/if}

		{#if ledgerExpanded}
			<p class="px-4 py-6 text-sm text-slate-500">
				The {ledgerTitle.toLowerCase()} are open in the full table.
			</p>
		{:else}
			{@render ledger(false)}
		{/if}
	</PageContentTemplate>
</AppShellTemplate>

<Drawer
	bind:open={ledgerExpanded}
	title={ledgerTitle}
	width="max-w-[96vw]"
	padded={false}
	onClose={focusFullTableButton}
>
	{@render ledger(true)}
</Drawer>

<!-- The table renders in one place at a time: on the page, or in the full-table slide-over,
     which is the same toolbar and table with the whole screen to work in. -->
{#snippet ledger(full: boolean)}
	<div class="flex min-h-0 flex-1 flex-col">
		<!-- The status filter belongs to the positions, so it is absent on the other tab. -->
		<LedgerToolbar
			title={full ? undefined : ledgerTitle}
			meta={tableMeta}
			showSearch
			searchValue={searchQuery}
			onSearch={(q) => (searchQuery = q)}
			searchPlaceholder={tab === 'positions' ? 'Search symbol or name…' : 'Search transactions…'}
			searchAriaLabel={tab === 'positions' ? 'Search positions' : 'Search transactions'}
			before={viewTabs}
			filters={tab === 'positions' ? statusFilter : undefined}
		>
			{#snippet actions()}
				<Button variant="ruled" onclick={() => void openImport()}>
					<Icon icon="heroicons:cloud-arrow-up" />
					Import CSV
				</Button>
				<Button shape="default" onclick={openTx}>
					<Icon icon="heroicons:plus" />
					Add transaction
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

		{#if tab === 'positions'}
			<DataTable
				rows={visiblePositions}
				{loading}
				{error}
				emptyText={positionsEmptyText}
				onRetry={() => void loadAll()}
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
					{ key: 'cost_basis', header: 'Cost basis', align: 'right', cell: costBasisCell },
					{ key: 'pnl_amount', header: 'P&L', align: 'right', cell: pnlAmountCell },
					{ key: 'pnl', header: 'Unrealized', align: 'right', cell: pnlCell },
					{ key: 'weight', header: 'Weight', align: 'right', cell: weightCell }
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
				onRowClick={openDetail}
				onRetry={() => void loadAll()}
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
			>
				{#snippet footer()}
					<FooterBar
						total={txTotal}
						bind:limit={txLimit}
						bind:offset={txOffset}
						onPageChange={() => void changeTransactionsPage()}
						onLimitChange={() => void changeTransactionsPage()}
					/>
				{/snippet}
			</DataTable>
		{/if}
	</div>
{/snippet}

<ImportDialog bind:open={importOpen} {vendors} onFinished={() => void loadAll()} />

<Dialog
	bind:open={txOpen}
	title="New transaction"
	size="md"
	closeOnEscape={!creatingTx}
	closeOnBackdrop={!creatingTx}
>
	<!-- A refused save stays in the dialog, directly under its header and full-bleed, with the entry
	     untouched behind it. The band carries role="alert", so it is announced and not just coloured. -->
	{#if txError}
		<NoticeBand
			intent="error"
			surface="inset"
			gutter="dialog"
			title={txErrorTitle}
			class="-mx-5 -mt-4 mb-4"
		>
			{txError}
		</NoticeBand>
	{/if}

	<div class="space-y-3">
		<TransactionDateHeader
			bind:value={txDate}
			bind:open={txPickerOpen}
			today={txToday}
			disabled={creatingTx}
		/>

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

{#if detailRow}
	<!-- Keyed on every open so each one starts from the row as stored, not from earlier edits. -->
	{#key detailOpenCount}
		<PortfolioTransactionDrawer
			bind:open={detailOpen}
			transaction={detailRow}
			editable={isManualPortfolioTransaction(detailRow)}
			originLabel={portfolioOriginLabel(detailRow)}
			saving={savingDetail}
			deleting={deletingDetail}
			error={detailError}
			errorTitle={detailErrorTitle}
			onSave={handleDetailSave}
			onDelete={handleDetailDelete}
		/>
	{/key}
{/if}
