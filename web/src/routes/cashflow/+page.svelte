<script lang="ts">
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import DonutChart from '$lib/components/organisms/charts/DonutChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import CashflowTransactionsTable from '$lib/components/organisms/cashflow-transactions-table/CashflowTransactionsTable.svelte';
	import TransactionFormModal from '$lib/components/organisms/transaction-form-modal/TransactionFormModal.svelte';
	import TransactionDetailDrawer from '$lib/components/organisms/transaction-detail-drawer/TransactionDetailDrawer.svelte';
	import MarkRecurringDialog from '$lib/components/organisms/mark-recurring-dialog/MarkRecurringDialog.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import {
		listCashflowTransactions,
		getCashflowMonthly,
		getCashflowTagDistribution,
		createCashflowTransactions,
		changeCashflowTransactionDate,
		tagCashflowTransactionsBySelection
	} from '$lib/services/cashflow';
	import {
		createRecurringItem,
		getRecurringOverview,
		linkRecurringTransactions
	} from '$lib/services/recurring';
	import { connectRealtime } from '$lib/services/realtime';
	import { toast } from '$lib/stores/toast.svelte';
	import { accountStore } from '$lib/stores/account.svelte';
	import type { CashflowTransactionFormValue } from '$lib/components/organisms/transaction-form-modal/transaction-form-modal.types';
	import {
		parseQuery,
		serializeQuery,
		type QuerySchema,
		type QueryState
	} from '$lib/url/routeQuery';
	import { pushQuery } from '$lib/url/queryState';
	import { scaledToNumber } from '$lib/api/money';
	import { cashflowOriginLabel, isManualCashflowTransaction } from '$lib/api/transactions';
	import { ApiError } from '$lib/api/client';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { chartColors, donutRamps } from '$lib/charts/theme';
	import type {
		CashflowDirection,
		CashflowTransaction,
		CashflowTransactionsQuery,
		CashflowMonthlyPoint,
		RecurringItem,
		TagDistributionEntry
	} from '$lib/api/types';
	import type { MarkRecurringValue } from '$lib/components/organisms/mark-recurring-dialog/mark-recurring-dialog.types';
	import type { SortDirection } from '$lib/components/organisms/data-table/data-table.types';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	const schema: QuerySchema = {
		description: { type: 'string' },
		tags: { type: 'string[]' },
		direction: { type: 'string' },
		sort_by: { type: 'string' },
		sort_order: { type: 'string' },
		limit: { type: 'number' },
		offset: { type: 'number' },
		from: { type: 'string' },
		to: { type: 'string' }
	};

	const initial = parseQuery(page.url.searchParams, schema);

	let descriptionFilter = $state((initial.description as string) || '');
	let tagFilter = $state(initial.tags as string[]);
	let directionFilter = $state(((initial.direction as string) || null) as CashflowDirection | null);
	let sortKey = $state((initial.sort_by as string) || 'date');
	let sortDirection = $state(((initial.sort_order as string) || 'desc') as SortDirection);
	let limit = $state((initial.limit as number) || 25);
	let offset = $state((initial.offset as number) || 0);
	let from = $state((initial.from as string) || '');
	let to = $state((initial.to as string) || '');

	let rows = $state<CashflowTransaction[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let selectedIds = $state<string[]>([]);

	let monthly = $state<CashflowMonthlyPoint[]>([]);
	let incoming = $state<TagDistributionEntry[]>([]);
	let outgoing = $state<TagDistributionEntry[]>([]);
	let analyticsLoading = $state(true);

	let createOpen = $state(false);
	let creating = $state(false);
	let createError = $state<string | null>(null);

	let markOpen = $state(false);
	let marking = $state(false);
	let markError = $state<string | null>(null);
	let recurringItems = $state<RecurringItem[]>([]);

	let detailRow = $state<CashflowTransaction | null>(null);
	let detailOpen = $state(false);
	let detailDate = $state('');
	let savingDate = $state(false);
	let dateError = $state<string | null>(null);

	const tagOptions = $derived(
		incoming
			.concat(outgoing)
			.map((entry) => entry.tag)
			.filter((tag, index, all) => tag !== '' && all.indexOf(tag) === index)
			.map((tag) => ({ value: tag, label: tag }))
	);

	// An empty ledger and an empty result set are different situations, so they read differently.
	const filtering = $derived(
		Boolean(descriptionFilter || tagFilter.length > 0 || directionFilter || from || to)
	);
	const emptyText = $derived(
		filtering
			? 'No transactions match your filters'
			: 'No transactions yet. Add one to start your ledger.'
	);

	const euro = (n: number) => `€${n.toLocaleString('en', { maximumFractionDigits: 0 })}`;
	const monthShort = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	function currentQuery(): CashflowTransactionsQuery {
		return {
			description: descriptionFilter || undefined,
			tags: tagFilter.join(',') || undefined,
			direction: directionFilter ?? undefined,
			sort_by: (sortKey || 'date') as CashflowTransactionsQuery['sort_by'],
			sort_order: sortDirection,
			limit,
			offset,
			from: from || undefined,
			to: to || undefined,
			hide_ignored: true
		};
	}

	function urlState(): QueryState {
		return {
			description: descriptionFilter,
			tags: tagFilter,
			direction: directionFilter ?? '',
			sort_by: sortKey,
			sort_order: sortDirection,
			limit,
			offset,
			from,
			to
		};
	}

	async function load(query: CashflowTransactionsQuery) {
		loading = true;
		error = null;
		try {
			const result = await listCashflowTransactions(query);
			rows = result.data;
			total = result.pagination.total;
		} catch {
			error = 'Failed to load transactions';
		} finally {
			loading = false;
		}
	}

	function syncUrl() {
		void pushQuery('/cashflow', serializeQuery(urlState(), schema), { replace: true });
	}

	// Reload whenever the working state changes (initial + every filter/sort/page change).
	$effect(() => {
		void load(currentQuery());
	});

	async function loadAnalytics() {
		analyticsLoading = true;
		try {
			const range = { from: from || undefined, to: to || undefined };
			const [monthlyRes, dist] = await Promise.all([
				getCashflowMonthly(range),
				getCashflowTagDistribution(range)
			]);
			monthly = monthlyRes.data;
			incoming = dist.incoming;
			outgoing = dist.outgoing;
		} catch {
			// Analytics are supplementary: leave the charts to render their own empty
			// state rather than letting an unhandled rejection escape the page.
			monthly = [];
			incoming = [];
			outgoing = [];
		} finally {
			analyticsLoading = false;
		}
	}

	// Re-fetch the trend + donuts whenever the date range changes (also covers the initial load).
	$effect(() => {
		void from;
		void to;
		void loadAnalytics();
	});

	onMount(() => {
		let realtime: { disconnect: () => void } | null = null;
		void accountStore.ensureLoaded().then(() => {
			if (!accountStore.hasAccount) return;
			realtime = connectRealtime({
				accountId: accountStore.activeId,
				events: ['import.completed', 'bulk_tag.completed'],
				onRefresh: () => {
					void load(currentQuery());
					void loadAnalytics();
				}
			});
		});
		return () => realtime?.disconnect();
	});

	function onSort(key: string, direction: SortDirection) {
		sortKey = key;
		sortDirection = direction;
		syncUrl();
	}
	function onPageChange() {
		syncUrl();
	}
	function onLimitChange() {
		offset = 0;
		syncUrl();
	}
	function onFilterChange() {
		offset = 0;
		syncUrl();
	}
	function onRangeSelect(rangeFrom: string, rangeTo: string) {
		from = rangeFrom;
		to = rangeTo;
		offset = 0;
		syncUrl();
		toast.info(`Filtered to ${rangeFrom} – ${rangeTo}`);
	}

	async function handleCreate(value: CashflowTransactionFormValue) {
		creating = true;
		createError = null;
		try {
			await accountStore.ensureLoaded();
			await createCashflowTransactions({
				transactions: [
					{
						date: value.date,
						amount: value.amount,
						type: value.type,
						description: value.description,
						// The backend requires a non-blank note and tag per row.
						note: value.note || value.description,
						tag: value.tag || 'Uncategorized'
					}
				]
			});
			createOpen = false;
			toast.success('Transaction created');
			void load(currentQuery());
			void loadAnalytics();
		} catch (err) {
			createError =
				err instanceof ApiError && err.status === 409
					? `A transaction with the same amount and description already exists on ${formatDisplayDate(value.date)}. Pick another day.`
					: 'Failed to create transaction';
			toast.error('Failed to create transaction');
		} finally {
			creating = false;
		}
	}

	function openDetail(row: CashflowTransaction) {
		detailRow = row;
		detailDate = row.date.slice(0, 10);
		dateError = null;
		detailOpen = true;
	}

	async function handleDateChange(date: string) {
		const row = detailRow;
		if (!row) return;
		savingDate = true;
		dateError = null;
		try {
			await changeCashflowTransactionDate({ id: row.id, date });
			detailOpen = false;
			toast.success(`Moved to ${formatDisplayDate(date)}`);
			void load(currentQuery());
			void loadAnalytics();
		} catch (err) {
			dateError = dateChangeMessage(err, date);
		} finally {
			savingDate = false;
		}
	}

	// The refusals worth naming are the ones the reader can act on: a date that already holds an
	// identical transaction, or a row that came from a statement and keeps its date.
	function dateChangeMessage(err: unknown, date: string): string {
		if (err instanceof ApiError && err.status === 409) {
			return `A transaction with the same amount and description already exists on ${formatDisplayDate(date)}. Pick another day.`;
		}
		if (err instanceof ApiError && err.status === 422) {
			return 'This transaction came from an import, so it keeps its statement date.';
		}
		return 'Could not change the date. Try again.';
	}

	async function handleBulkTag() {
		const tag = window.prompt('Tag for the selected transactions')?.trim();
		if (!tag) return;
		const ids = selectedIds;
		try {
			await tagCashflowTransactionsBySelection({ tag, ids });
			toast.success(`Tagged ${ids.length} transactions`);
			selectedIds = [];
			void load(currentQuery());
			void loadAnalytics();
		} catch {
			toast.error('Failed to tag transactions');
		}
	}

	const selectedRows = $derived(rows.filter((row) => selectedIds.includes(row.id)));

	// The dialog offers "add to an existing item" first, so the items are read when it
	// opens rather than on every page load: most visits never mark anything.
	async function openMark() {
		markError = null;
		markOpen = true;
		try {
			const overview = await getRecurringOverview();
			recurringItems = [...overview.expenses, ...overview.income];
		} catch {
			// Starting a new item still works without the list, so this is not reported as a
			// failure — the dialog simply offers only that route.
			recurringItems = [];
		}
	}

	async function handleMarkRecurring(value: MarkRecurringValue) {
		const ids = selectedIds;
		marking = true;
		markError = null;
		try {
			if (value.mode === 'existing') {
				await linkRecurringTransactions(value.itemId, { ids });
			} else {
				await createRecurringItem({
					name: value.name,
					direction: value.direction,
					rhythm: value.rhythm,
					ids
				});
			}
			markOpen = false;
			selectedIds = [];
			toast.success(
				value.mode === 'existing'
					? `Linked ${ids.length} transactions`
					: `${value.name} added, with ${ids.length} transactions`
			);
		} catch (err) {
			markError = markRecurringMessage(err);
		} finally {
			marking = false;
		}
	}

	// The refusals worth naming are the ones the reader can act on: a name already taken,
	// or an item that was ended and no longer takes transactions.
	function markRecurringMessage(err: unknown): string {
		if (err instanceof ApiError && err.status === 409) {
			return 'A recurring item with that name already exists. Pick another name, or add these to it.';
		}
		if (err instanceof ApiError && err.status === 422) {
			return 'That item was ended, so it takes no new transactions.';
		}
		return 'Could not mark these as recurring. Try again.';
	}

	const navActions: MenuItem[] = [{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' }];
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Cashflow"
			showSearch
			searchValue={descriptionFilter}
			searchPlaceholder="Search description…"
			onSearch={(q) => {
				descriptionFilter = q;
				onFilterChange();
			}}
			showDateRange
			dateFrom={from || null}
			dateTo={to || null}
			onDateChange={(r) => {
				from = r.from ?? '';
				to = r.to ?? '';
				offset = 0;
				syncUrl();
			}}
			actions={navActions}
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
				<AnalyticsCard title="Net trend" class="lg:col-span-2">
					<TimeSeriesChart
						height="h-44"
						labels={monthly.map((m) => m.month)}
						xTickFormat={monthShort}
						loading={analyticsLoading}
						enableRangeSelect
						{onRangeSelect}
						datasets={[
							{
								label: 'Net',
								data: monthly.map((m) => scaledToNumber(m.net_cents)),
								color: chartColors.net,
								signed: true
							}
						]}
					/>
				</AnalyticsCard>
				<AnalyticsCard title="Incoming">
					<DonutChart
						data={incoming.map((e) => ({ label: e.tag, value: scaledToNumber(e.totalCents) }))}
						ramp={donutRamps.incoming}
						loading={analyticsLoading}
						formatValue={euro}
						centerLabel="In"
					/>
				</AnalyticsCard>
				<AnalyticsCard title="Outgoing">
					<DonutChart
						data={outgoing.map((e) => ({ label: e.tag, value: scaledToNumber(e.totalCents) }))}
						ramp={donutRamps.outgoing}
						loading={analyticsLoading}
						formatValue={euro}
						centerLabel="Out"
					/>
				</AnalyticsCard>
			</div>
		{/snippet}

		<LedgerToolbar
			title="Transactions"
			actionLabel="Add transaction"
			onAdd={() => (createOpen = true)}
		/>
		<CashflowTransactionsTable
			{rows}
			{loading}
			{error}
			{total}
			bind:limit
			bind:offset
			bind:selectedIds
			{sortKey}
			{sortDirection}
			bind:descriptionFilter
			bind:tagFilter
			bind:directionFilter
			{tagOptions}
			{emptyText}
			{onSort}
			{onPageChange}
			{onLimitChange}
			{onFilterChange}
			onRowClick={openDetail}
		>
			{#snippet bulkActions()}
				<Button size="sm" variant="ghost" intent="secondary" onclick={handleBulkTag}>Tag</Button>
				<Button size="sm" variant="ghost" intent="secondary" onclick={openMark}>
					Mark as recurring
				</Button>
			{/snippet}
		</CashflowTransactionsTable>
	</PageContentTemplate>
</AppShellTemplate>

<TransactionFormModal
	bind:open={createOpen}
	onSubmit={handleCreate}
	submitting={creating}
	error={createError}
/>

<MarkRecurringDialog
	bind:open={markOpen}
	selection={selectedRows}
	items={recurringItems}
	saving={marking}
	error={markError}
	onSubmit={handleMarkRecurring}
/>

{#snippet detailFields()}
	{#if detailRow}
		<div class="flex items-center justify-between gap-3 py-3">
			<dt class="text-sm text-slate-500">Amount</dt>
			<dd><Money amount={scaledToNumber(detailRow.amountCents)} currency="EUR" size="sm" /></dd>
		</div>
		<div class="flex items-center justify-between gap-3 py-3">
			<dt class="text-sm text-slate-500">Note</dt>
			<dd class="text-sm text-slate-800">{detailRow.note || '—'}</dd>
		</div>
	{/if}
{/snippet}

{#if detailRow}
	<TransactionDetailDrawer
		bind:open={detailOpen}
		title={detailRow.description}
		originLabel={cashflowOriginLabel(detailRow)}
		subtitle={detailRow.tag || 'Untagged'}
		editable={isManualCashflowTransaction(detailRow)}
		bind:date={detailDate}
		originalDate={detailRow.date.slice(0, 10)}
		saving={savingDate}
		error={dateError}
		onSave={handleDateChange}
		details={detailFields}
	/>
{/if}
