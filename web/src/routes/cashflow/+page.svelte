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
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import CardRail from '$lib/components/molecules/card-rail/CardRail.svelte';
	import { cardRailWidths } from '$lib/components/molecules/card-rail/card-rail.variants';
	import ActionMenu from '$lib/components/molecules/action-menu/ActionMenu.svelte';
	import CashflowTransactionsTable from '$lib/components/organisms/cashflow-transactions-table/CashflowTransactionsTable.svelte';
	import TransactionFormModal from '$lib/components/organisms/transaction-form-modal/TransactionFormModal.svelte';
	import ImportDialog from '$lib/components/organisms/import-dialog/ImportDialog.svelte';
	import { goto } from '$app/navigation';
	import { listVendors } from '$lib/services/vendors';
	import TransactionDetailDrawer from '$lib/components/organisms/transaction-detail-drawer/TransactionDetailDrawer.svelte';
	import RunningMonthCard from '$lib/components/organisms/wealth-goal-cards/RunningMonthCard.svelte';
	import StreakCard from '$lib/components/organisms/wealth-goal-cards/StreakCard.svelte';
	import MonthlyStandingCard from '$lib/components/organisms/wealth-goal-cards/MonthlyStandingCard.svelte';
	import NoGoalCard from '$lib/components/organisms/wealth-goal-cards/NoGoalCard.svelte';
	import GoalErrorCard from '$lib/components/organisms/wealth-goal-cards/GoalErrorCard.svelte';
	import GoalLoadingCard from '$lib/components/organisms/wealth-goal-cards/GoalLoadingCard.svelte';
	import SetGoalDialog from '$lib/components/organisms/set-goal-dialog/SetGoalDialog.svelte';
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
		getWealthGoalStanding,
		markCashflowPurposeByFilter,
		markCashflowPurposeBySelection,
		setWealthGoal
	} from '$lib/services/wealthgoal';
	import { monthLabel, monthRange } from '$lib/api/wealthgoal';
	import { connectRealtime } from '$lib/services/realtime';
	import { toast } from '$lib/stores/toast.svelte';
	import { accountStore } from '$lib/stores/account.svelte';
	import { periodStore, type PeriodPreset } from '$lib/stores/period.svelte';
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
		MonthStanding,
		TagDistributionEntry,
		TransactionPurpose,
		Vendor,
		WealthGoalStandingResponse
	} from '$lib/api/types';
	import type { SortDirection } from '$lib/components/organisms/data-table/data-table.types';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	const schema: QuerySchema = {
		description: { type: 'string' },
		tags: { type: 'string[]' },
		direction: { type: 'string' },
		purpose: { type: 'string[]' },
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
	let purposeFilter = $state(initial.purpose as string[]);
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

	let standing = $state<WealthGoalStandingResponse | null>(null);
	let standingLoading = $state(true);
	let standingError = $state(false);
	let goalOpen = $state(false);
	let goalPercent = $state(30);
	let goalSaving = $state(false);
	let goalError = $state<string | null>(null);
	// The month the standing sent you to, shown on the ledger header so the jump is undoable
	// by eye. The period is app-wide, so the jump also remembers the period it replaced.
	let scopedMonth = $state<string | null>(null);
	let periodBeforeScope: { from: string; to: string; preset: PeriodPreset } | null = null;

	let createOpen = $state(false);
	let creating = $state(false);
	let createError = $state<string | null>(null);
	let importOpen = $state(false);
	let brokerageVendors = $state<Vendor[]>([]);

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
	// The period is app-wide and always set, so it is not what makes this a filtered view.
	const filtering = $derived(
		Boolean(
			descriptionFilter.trim() ||
				tagFilter.length > 0 ||
				directionFilter ||
				purposeFilter.length > 0
		)
	);
	const emptyText = $derived(
		descriptionFilter.trim()
			? `No transactions match “${descriptionFilter.trim()}”. Clear the search to see them all.`
			: filtering
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
			purpose: purposeFilter.join(',') || undefined,
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
			purpose: purposeFilter,
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

	// The standing scores whole calendar months, so it deliberately ignores the page's date
	// range: narrowing the ledger to a week must not make a month look empty.
	async function loadStanding() {
		standingLoading = true;
		standingError = false;
		try {
			standing = await getWealthGoalStanding();
			if (standing.goal) goalPercent = standing.goal.share_percent;
		} catch {
			standing = null;
			standingError = true;
		} finally {
			standingLoading = false;
		}
	}

	$effect(() => {
		void loadStanding();
	});

	const runningMonth = $derived(standing?.months.find((m) => m.result === 'in_progress') ?? null);
	const finishedMonths = $derived(standing?.months.filter((m) => m.result !== 'in_progress') ?? []);
	// A new goal applies from the month you set it in; before the first one that is this month.
	const goalEffectiveFrom = $derived(
		runningMonth?.month ?? new Date().toISOString().slice(0, 7) + '-01'
	);

	async function handleSaveGoal(percent: number) {
		goalSaving = true;
		goalError = null;
		try {
			await setWealthGoal({ share_percent: percent });
			goalOpen = false;
			toast.success(`Monthly goal set to ${percent}% of income`);
			void loadStanding();
		} catch (err) {
			goalError =
				err instanceof ApiError && err.status === 400
					? 'A goal is a whole percentage between 0 and 100.'
					: 'Could not save your monthly goal. Try again.';
		} finally {
			goalSaving = false;
		}
	}

	// The standing sends you to exactly the transactions that keep a month incomplete: the
	// month as the period, and the Purpose filter on "not assigned". The period it replaces is
	// remembered so "Show all transactions" puts it back.
	function openUnassigned(month: MonthStanding) {
		const range = monthRange(month.month);
		periodBeforeScope = {
			from: periodStore.from,
			to: periodStore.to,
			preset: periodStore.preset
		};
		periodStore.set({ from: range.from, to: range.to });
		purposeFilter = ['none'];
		scopedMonth = month.month;
		selectedIds = [];
		offset = 0;
		syncUrl();
	}

	function clearScope() {
		scopedMonth = null;
		purposeFilter = [];
		if (periodBeforeScope) {
			periodStore.set(periodBeforeScope);
			periodBeforeScope = null;
		}
		offset = 0;
		syncUrl();
	}

	// Marking is reported honestly: income only sticks to incoming money and a contribution
	// only to outgoing money, so a mixed selection says how much of it actually moved.
	function reportMarked(updated: number, matched: number, purpose: TransactionPurpose) {
		const what = purpose === 'income' ? 'income' : purpose === 'wealth' ? 'wealth' : 'unassigned';
		if (updated === 0) {
			toast.info(
				purpose === 'income'
					? 'Nothing marked: only incoming money can be income.'
					: purpose === 'wealth'
						? 'Nothing marked: only outgoing money can be a contribution.'
						: 'Nothing to clear.'
			);
			return;
		}
		if (updated < matched) {
			toast.success(
				`Marked ${updated} of ${matched} transactions as ${what}; the rest are the other direction.`
			);
			return;
		}
		toast.success(`Marked ${updated} transactions as ${what}`);
	}

	// A mark is idempotent, so a double click corrupts nothing -- but it does fire a second
	// mutation, a second toast and a second pair of reloads that race each other.
	let marking = $state(false);

	async function handleMarkSelection(purpose: TransactionPurpose) {
		const ids = selectedIds;
		if (ids.length === 0 || marking) return;
		marking = true;
		try {
			const result = await markCashflowPurposeBySelection({
				purpose: purpose === '' ? 'none' : purpose,
				ids
			});
			reportMarked(result.updated_count, result.matched_count, purpose);
			selectedIds = [];
			void load(currentQuery());
			void loadStanding();
		} catch {
			toast.error('Failed to mark transactions');
		} finally {
			marking = false;
		}
	}

	async function handleMarkFilter(purpose: TransactionPurpose) {
		if (marking) return;
		marking = true;
		try {
			const result = await markCashflowPurposeByFilter({
				purpose: purpose === '' ? 'none' : purpose,
				filters: {
					description: descriptionFilter || undefined,
					tags: tagFilter.join(',') || undefined,
					direction: directionFilter ?? undefined,
					purpose: purposeFilter.join(',') || undefined,
					from: from || undefined,
					to: to || undefined,
					hide_ignored: true
				}
			});
			reportMarked(result.updated_count, result.matched_count, purpose);
			selectedIds = [];
			void load(currentQuery());
			void loadStanding();
		} catch {
			toast.error('Failed to mark transactions');
		} finally {
			marking = false;
		}
	}

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
					void loadStanding();
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
		// Changing a filter by hand leaves the month the standing sent you to, so the ledger
		// header stops claiming you are still looking at it.
		scopedMonth = null;
		offset = 0;
		syncUrl();
	}
	function onRangeSelect(rangeFrom: string, rangeTo: string) {
		// Dragging the chart picks a period like any other, so it goes through the same store.
		scopedMonth = null;
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

	const tableMeta = $derived.by(() => {
		if (loading) return 'Loading…';
		if (error) return 'Could not load';
		const selected = selectedIds.length > 0 ? ` · ${selectedIds.length} selected` : '';
		return `${total} ${total === 1 ? 'row' : 'rows'}${selected}`;
	});

	// Imports need a brokerage vendor, which the cashflow page does not otherwise load,
	// so the list is fetched when the dialog is first opened rather than on every visit.
	async function openImport() {
		if (brokerageVendors.length === 0) {
			try {
				brokerageVendors = (await listVendors()).filter((v) => v.active && v.type === 'portfolio');
			} catch {
				// Leave the list empty; the dialog says there is no brokerage account.
			}
		}
		importOpen = true;
	}

	const cardGroups = [
		{ id: 'cashflow', label: 'Cashflow overview' },
		{ id: 'goal', label: 'Wealth goal' }
	];

	// Marking a purpose is an action on these rows, so it sits with the other row actions on the
	// ledger header. Three of them beside Tag, Import and Add would crowd the rule, so the
	// purposes themselves live in one menu.
	const selectionMarkItems: MenuItem[] = $derived([
		{ label: 'Mark as income', disabled: marking, onSelect: () => handleMarkSelection('income') },
		{ label: 'Mark as wealth', disabled: marking, onSelect: () => handleMarkSelection('wealth') },
		{ label: 'Clear purpose', disabled: marking, onSelect: () => handleMarkSelection('') }
	]);

	const filterMarkItems: MenuItem[] = $derived([
		{ label: 'Mark as income', disabled: marking, onSelect: () => handleMarkFilter('income') },
		{ label: 'Mark as wealth', disabled: marking, onSelect: () => handleMarkFilter('wealth') },
		{ label: 'Clear purpose', disabled: marking, onSelect: () => handleMarkFilter('') }
	]);
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar />
	{/snippet}

	<PageContentTemplate title="Cashflow">
		{#snippet analytics()}
			<!-- The band is one scrolling rail in two groups: the cashflow charts, and the wealth
			     goal. The ledger underneath never changes, so you mark transactions and read the
			     score on the same page. Which cards you see becomes a setting later. -->
			<CardRail
				groups={cardGroups}
				ariaLabel="Cashflow and wealth-goal cards"
				hint="Scroll sideways for the rest — which cards you see becomes a setting later."
			>
				<div data-card="trend" data-group="cashflow" class={cardRailWidths.wide}>
					<AnalyticsCard title="Net trend">
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
				</div>
				<div data-card="incoming" data-group="cashflow" class={cardRailWidths.default}>
					<AnalyticsCard title="Incoming">
						<DonutChart
							data={incoming.map((e) => ({ label: e.tag, value: scaledToNumber(e.totalCents) }))}
							ramp={donutRamps.incoming}
							loading={analyticsLoading}
							formatValue={euro}
							centerLabel="In"
						/>
					</AnalyticsCard>
				</div>
				<div data-card="outgoing" data-group="cashflow" class={cardRailWidths.default}>
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

				{#if standingLoading}
					<div data-card="month" data-group="goal" class={cardRailWidths.wide}>
						<AnalyticsCard title="This month against your goal">
							<GoalLoadingCard shape="month" />
						</AnalyticsCard>
					</div>
					<div data-card="streak" data-group="goal" class={cardRailWidths.default}>
						<AnalyticsCard title="Streak">
							<GoalLoadingCard shape="streak" />
						</AnalyticsCard>
					</div>
					<div data-card="standing" data-group="goal" class={cardRailWidths.wide}>
						<AnalyticsCard title="Monthly standing">
							<GoalLoadingCard shape="rows" />
						</AnalyticsCard>
					</div>
				{:else if standingError}
					<div data-card="month" data-group="goal" class={cardRailWidths.solo}>
						<AnalyticsCard title="Wealth goal">
							<GoalErrorCard onRetry={() => loadStanding()} />
						</AnalyticsCard>
					</div>
				{:else if !standing?.goal || !runningMonth}
					<div data-card="month" data-group="goal" class={cardRailWidths.solo}>
						<AnalyticsCard title="Wealth goal">
							<NoGoalCard onSet={() => (goalOpen = true)} />
						</AnalyticsCard>
					</div>
				{:else}
					<div data-card="month" data-group="goal" class={cardRailWidths.wide}>
						<AnalyticsCard title="{monthLabel(runningMonth.month)} against your goal">
							<RunningMonthCard
								month={runningMonth}
								stacked
								onOpenUnassigned={openUnassigned}
								onAdjust={() => (goalOpen = true)}
							/>
						</AnalyticsCard>
					</div>
					<div data-card="streak" data-group="goal" class={cardRailWidths.default}>
						<AnalyticsCard title="Streak">
							<StreakCard
								months={standing.months}
								currentStreak={standing.current_streak}
								bestStreak={standing.best_streak}
							/>
						</AnalyticsCard>
					</div>
					<div data-card="standing" data-group="goal" class={cardRailWidths.wide}>
						<AnalyticsCard title="Monthly standing">
							<MonthlyStandingCard months={finishedMonths} onOpenUnassigned={openUnassigned} />
						</AnalyticsCard>
					</div>
				{/if}
			</CardRail>
		{/snippet}

		<!-- Everything that acts on these rows lives here: searching, tagging or marking a
		     selection, importing a statement and adding one. -->
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
			{#snippet before()}
				{#if scopedMonth}
					<Badge intent="info" variant="soft" size="sm">{monthLabel(scopedMonth)}</Badge>
					<Button size="sm" variant="ghost" intent="secondary" onclick={clearScope}>
						Show all transactions
					</Button>
				{/if}
			{/snippet}
			{#snippet actions()}
				{#if selectedIds.length > 0}
					<Button variant="ruled" onclick={handleBulkTag}>
						<Icon icon="heroicons:tag" />
						Tag {selectedIds.length} selected
					</Button>
					<ActionMenu label="Mark {selectedIds.length} selected…" items={selectionMarkItems} />
				{/if}
				<!-- "Mark all matches" acts on the filter, so it names the number the filter returns. -->
				{#if !loading && !error && filtering && total > 0}
					<ActionMenu label="Mark all {total} matches…" items={filterMarkItems} />
				{/if}
				<Button variant="ruled" onclick={() => void openImport()}>
					<Icon icon="heroicons:cloud-arrow-up" />
					Import CSV
				</Button>
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
			{total}
			bind:limit
			bind:offset
			bind:selectedIds
			{sortKey}
			{sortDirection}
			bind:descriptionFilter
			bind:tagFilter
			bind:directionFilter
			bind:purposeFilter
			{tagOptions}
			{emptyText}
			{onSort}
			{onPageChange}
			{onLimitChange}
			{onFilterChange}
			onRowClick={openDetail}
		/>
	</PageContentTemplate>
</AppShellTemplate>

<TransactionFormModal
	bind:open={createOpen}
	onSubmit={handleCreate}
	submitting={creating}
	error={createError}
/>

<SetGoalDialog
	bind:open={goalOpen}
	bind:percent={goalPercent}
	exampleIncome={scaledToNumber(runningMonth?.income_cents ?? 0)}
	effectiveFrom={goalEffectiveFrom}
	saving={goalSaving}
	error={goalError}
	onSave={handleSaveGoal}
/>

<ImportDialog
	bind:open={importOpen}
	vendors={brokerageVendors}
	onFinished={() => {
		void load(currentQuery());
		void loadAnalytics();
		void loadStanding();
	}}
	onGoToPortfolio={() => void goto('/portfolio')}
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
