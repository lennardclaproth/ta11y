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
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import CashflowTransactionsTable from '$lib/components/organisms/cashflow-transactions-table/CashflowTransactionsTable.svelte';
	import TransactionFormModal from '$lib/components/organisms/transaction-form-modal/TransactionFormModal.svelte';
	import {
		listCashflowTransactions,
		getCashflowMonthly,
		getCashflowTagDistribution,
		createCashflowTransactions,
		tagCashflowTransactionsBySelection
	} from '$lib/services/cashflow';
	import { connectRealtime } from '$lib/services/realtime';
	import { toast } from '$lib/stores/toast.svelte';
	import { accountStore } from '$lib/stores/account.svelte';
	import { periodStore } from '$lib/stores/period.svelte';
	import type { CashflowTransactionFormValue } from '$lib/components/organisms/transaction-form-modal/transaction-form-modal.types';
	import {
		parseQuery,
		serializeQuery,
		type QuerySchema,
		type QueryState
	} from '$lib/url/routeQuery';
	import { pushQuery } from '$lib/url/queryState';
	import { scaledToNumber } from '$lib/api/money';
	import { chartColors, donutRamps } from '$lib/charts/theme';
	import type {
		CashflowDirection,
		CashflowTransaction,
		CashflowTransactionsQuery,
		CashflowMonthlyPoint,
		TagDistributionEntry
	} from '$lib/api/types';
	import type { SortDirection } from '$lib/components/organisms/data-table/data-table.types';

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

	// The period is the app's, not this page's: a link that carries one seeds it, but only
	// while nothing has chosen a period yet, so returning here never undoes a later choice.
	periodStore.seed((initial.from as string) || null, (initial.to as string) || null);
	const from = $derived(periodStore.from);
	const to = $derived(periodStore.to);

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

	const tagOptions = $derived(
		incoming
			.concat(outgoing)
			.map((entry) => entry.tag)
			.filter((tag, index, all) => tag !== '' && all.indexOf(tag) === index)
			.map((tag) => ({ value: tag, label: tag }))
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

	// A period change invalidates the load effect and then resets `offset`, so two queries can be
	// in flight at once. Only the newest may write the table, or a slow response for the page you
	// just left can land last and show rows the URL and the pagination control disagree with.
	let requestId = 0;

	async function load(query: CashflowTransactionsQuery) {
		const id = ++requestId;
		loading = true;
		error = null;
		try {
			const result = await listCashflowTransactions(query);
			if (id !== requestId) return;
			rows = result.data;
			total = result.pagination.total;
		} catch {
			if (id !== requestId) return;
			error = 'Failed to load transactions';
		} finally {
			if (id === requestId) loading = false;
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

	// Keep the URL in step with the period after it is changed elsewhere (the overview, or a
	// drag on the chart). The first run is the page's own initial state, which the URL already
	// holds -- rewriting it there would discard a deep-linked page offset.
	let periodSynced = false;
	$effect(() => {
		void periodStore.from;
		void periodStore.to;
		if (!periodSynced) {
			periodSynced = true;
			return;
		}
		offset = 0;
		syncUrl();
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
		// Dragging the chart picks a period like any other, so it goes through the same store.
		periodStore.set({ from: rangeFrom, to: rangeTo });
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
		} catch {
			createError = 'Failed to create transaction';
			toast.error('Failed to create transaction');
		} finally {
			creating = false;
		}
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

	const tableMeta = $derived.by(() => {
		if (loading) return 'Loading…';
		if (error) return 'Could not load';
		const selected = selectedIds.length > 0 ? ` · ${selectedIds.length} selected` : '';
		return `${total} ${total === 1 ? 'row' : 'rows'}${selected}`;
	});

	// An account with no transactions and a search that matched none are different situations
	// with different next steps, so they do not read the same.
	const emptyText = $derived(
		descriptionFilter.trim()
			? `No transactions match “${descriptionFilter.trim()}”. Clear the search to see them all.`
			: 'No transactions yet. Add one to get started.'
	);
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar />
	{/snippet}

	<PageContentTemplate title="Cashflow">
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

		<!-- Everything that acts on these rows lives here: searching, tagging a selection and
		     adding one. -->
		<LedgerToolbar
			title="Transactions"
			meta={tableMeta}
			showSearch
			searchValue={descriptionFilter}
			searchPlaceholder="Search description…"
			searchAriaLabel="Search transactions by description"
			onSearch={(q) => {
				descriptionFilter = q;
				onFilterChange();
			}}
		>
			{#snippet actions()}
				{#if selectedIds.length > 0}
					<Button variant="ruled" onclick={handleBulkTag}>
						<Icon icon="heroicons:tag" />
						Tag {selectedIds.length} selected
					</Button>
				{/if}
				<Button shape="default" onclick={() => (createOpen = true)}>
					<Icon icon="heroicons:plus" />
					Add transaction
				</Button>
			{/snippet}
		</LedgerToolbar>
		<CashflowTransactionsTable
			{rows}
			{loading}
			{error}
			{emptyText}
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
			{onSort}
			{onPageChange}
			{onLimitChange}
			{onFilterChange}
		/>
	</PageContentTemplate>
</AppShellTemplate>

<TransactionFormModal
	bind:open={createOpen}
	onSubmit={handleCreate}
	submitting={creating}
	error={createError}
/>
