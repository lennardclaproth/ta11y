<script lang="ts">
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import { onMount, untrack } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
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
	import CashflowImportDialog from '$lib/components/organisms/cashflow-import-dialog/CashflowImportDialog.svelte';
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
	import MarkRecurringDialog from '$lib/components/organisms/mark-recurring-dialog/MarkRecurringDialog.svelte';
	import TagTransactionsDialog from '$lib/components/organisms/tag-transactions-dialog/TagTransactionsDialog.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import {
		listCashflowTransactions,
		getCashflowMonthly,
		getCashflowTagDistribution,
		createCashflowTransactions,
		changeCashflowTransactionDate,
		ignoreCashflowTransactionsBySelection,
		listCashflowTags,
		tagCashflowTransactionsBySelection
	} from '$lib/services/cashflow';
	import { listIgnoreRules } from '$lib/services/ignoreRules';
	import Switch from '$lib/components/atoms/switch/Switch.svelte';
	import {
		getWealthGoalStanding,
		markCashflowPurposeByFilter,
		markCashflowPurposeBySelection,
		setWealthGoal
	} from '$lib/services/wealthgoal';
	import { monthLabel, monthRange } from '$lib/api/wealthgoal';
	import {
		createRecurringItem,
		getRecurringOverview,
		linkRecurringTransactions
	} from '$lib/services/recurring';
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
	import {
		cashflowOriginLabel,
		cashflowTagQuery,
		isManualCashflowTransaction,
		UNTAGGED_FILTER
	} from '$lib/api/transactions';
	import { ApiError } from '$lib/api/client';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { chartColors, donutRamps } from '$lib/charts/theme';
	import type {
		CashflowDirection,
		CashflowTag,
		CashflowTransaction,
		CashflowTransactionsQuery,
		CashflowMonthlyPoint,
		MonthStanding,
		IgnoreRule,
		RecurringItem,
		TagDistributionEntry,
		TransactionPurpose,
		Vendor,
		WealthGoalStandingResponse
	} from '$lib/api/types';
	import type { MarkRecurringValue } from '$lib/components/organisms/mark-recurring-dialog/mark-recurring-dialog.types';
	import type { TagTransactionsValue } from '$lib/components/organisms/tag-transactions-dialog/tag-transactions-dialog.types';
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
	// The rows behind the selection, by id. A selection outlives the page it was made on, and
	// "select all" reaches rows no page has shown, so the current page alone cannot answer what
	// is selected.
	const knownRows = new SvelteMap<string, CashflowTransaction>();
	let selectingAll = $state(false);

	/**
	 * Ignored rows are out of the ledger by default — that is what ignoring them is for.
	 * Switching them back on is how you check what a rule did and put something back.
	 */
	let showIgnored = $state(false);
	/** Named on every ignored row, so the rule that did it is visible rather than implied. */
	let ignoreRules = $state<IgnoreRule[]>([]);

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
	let brokerageImportOpen = $state(false);
	let vendorsLoaded = $state(false);
	let bankVendors = $state<Vendor[]>([]);
	let brokerageVendors = $state<Vendor[]>([]);

	let tagOpen = $state(false);
	let tagging = $state(false);
	let tagError = $state<string | null>(null);
	/** What the tag dialog tags: the selection as it was when the dialog opened. */
	let tagRows = $state<CashflowTransaction[]>([]);
	let cashflowTags = $state<CashflowTag[]>([]);
	let tagsLoading = $state(false);

	let markOpen = $state(false);
	let marking = $state(false);
	let markError = $state<string | null>(null);
	/** What the dialog marks: the table's selection, or the one row opened in the drawer. */
	let markRows = $state<CashflowTransaction[]>([]);
	let recurringItems = $state<RecurringItem[]>([]);

	let detailRow = $state<CashflowTransaction | null>(null);
	let detailOpen = $state(false);
	let detailDate = $state('');
	let savingDate = $state(false);
	let dateError = $state<string | null>(null);

	// Untagged rows are a bucket in the tag totals, offered first so it can be picked on its own
	// or next to a tag. An account that has tagged nothing yet has only this choice.
	const tagOptions = $derived.by(() => {
		const tags = incoming.concat(outgoing).map((entry) => entry.tag);
		const named = tags.filter(
			(tag, index) => tag !== '' && tag !== UNTAGGED_FILTER && tags.indexOf(tag) === index
		);
		const hasUntagged = tags.some((tag) => tag === '' || tag === UNTAGGED_FILTER);
		return [
			...(hasUntagged ? [{ value: UNTAGGED_FILTER, label: 'Untagged' }] : []),
			...named.map((tag) => ({ value: tag, label: tag }))
		];
	});

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
			...cashflowTagQuery(tagFilter),
			direction: directionFilter ?? undefined,
			purpose: purposeFilter.join(',') || undefined,
			sort_by: (sortKey || 'date') as CashflowTransactionsQuery['sort_by'],
			sort_order: sortDirection,
			limit,
			offset,
			from: from || undefined,
			to: to || undefined,
			hide_ignored: !showIgnored
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
			for (const row of result.data) knownRows.set(row.id, row);
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
	// Only the period is a dependency: `syncUrl` reads the page offset and every filter, and
	// tracking those would make each page change rerun this and send the table back to page one.
	let periodSynced = false;
	$effect(() => {
		void periodStore.from;
		void periodStore.to;
		if (!periodSynced) {
			periodSynced = true;
			return;
		}
		untrack(() => {
			offset = 0;
			selectedIds = [];
			syncUrl();
		});
	});

	// Marking, saving a goal, an import and the realtime refresh all reload the standing, so two
	// calls can be in flight at once. As with `load`, only the newest may write: a slow first
	// response landing last would put the standing from before the mutation back on screen.
	let standingRequestId = 0;

	// The standing scores whole calendar months, so it deliberately ignores the page's date
	// range: narrowing the ledger to a week must not make a month look empty.
	async function loadStanding() {
		const id = ++standingRequestId;
		standingLoading = true;
		standingError = false;
		try {
			const result = await getWealthGoalStanding();
			if (id !== standingRequestId) return;
			standing = result;
			if (standing.goal) goalPercent = standing.goal.share_percent;
		} catch {
			if (id !== standingRequestId) return;
			standing = null;
			standingError = true;
		} finally {
			if (id === standingRequestId) standingLoading = false;
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

	// A period change already resets the offset and rewrites the URL (the effect above), so the
	// two paths below leave that to it and only push the query state themselves when they leave
	// the period where it is -- otherwise their change to the Purpose filter never reaches the URL.
	const periodMoves = (next: { from: string; to: string }) =>
		next.from !== periodStore.from || next.to !== periodStore.to;

	// The standing sends you to exactly the transactions that keep a month incomplete: the
	// month as the period, and the Purpose filter on "not assigned". The period it replaces is
	// remembered so "Show all transactions" puts it back.
	function openUnassigned(month: MonthStanding) {
		const range = monthRange(month.month);
		// Only a jump from an unscoped view may record the period: a second jump would otherwise
		// remember the month scope of the first, and "Show all transactions" would restore that.
		if (!scopedMonth) {
			periodBeforeScope = {
				from: periodStore.from,
				to: periodStore.to,
				preset: periodStore.preset
			};
		}
		const moves = periodMoves(range);
		periodStore.set({ from: range.from, to: range.to });
		purposeFilter = ['none'];
		scopedMonth = month.month;
		selectedIds = [];
		if (!moves) {
			offset = 0;
			syncUrl();
		}
	}

	function clearScope() {
		scopedMonth = null;
		purposeFilter = [];
		const moves = periodBeforeScope !== null && periodMoves(periodBeforeScope);
		if (periodBeforeScope) {
			periodStore.set(periodBeforeScope);
			periodBeforeScope = null;
		}
		if (!moves) {
			offset = 0;
			syncUrl();
		}
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
	let markingPurpose = $state(false);

	async function handleMarkSelection(purpose: TransactionPurpose) {
		const ids = selectedIds;
		if (ids.length === 0 || markingPurpose) return;
		markingPurpose = true;
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
			markingPurpose = false;
		}
	}

	async function handleMarkFilter(purpose: TransactionPurpose) {
		if (markingPurpose) return;
		markingPurpose = true;
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
			markingPurpose = false;
		}
	}

	// The rules are read once and only to name the rule on an ignored row; the ledger does
	// not otherwise depend on them, so a failure leaves the rows reading "By hand" rather
	// than taking the page down with it.
	async function loadIgnoreRules() {
		try {
			ignoreRules = await listIgnoreRules();
		} catch {
			ignoreRules = [];
		}
	}

	onMount(() => {
		void loadIgnoreRules();
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
		selectedIds = [];
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
			// The modal stays open with the entry, so the refusal belongs in it. A toast on top of it
			// would say the same thing twice.
			createError =
				err instanceof ApiError && err.status === 409
					? `A transaction with the same amount and description already exists on ${formatDisplayDate(value.date)}. Pick another day.`
					: 'Failed to create transaction';
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

	// The dialog opens at once and reads the tags while it is up: they are the account's whole
	// history rather than this period's, so they are read when tagging starts, not with the page.
	async function openTag() {
		tagRows = selectedRows;
		tagError = null;
		tagOpen = true;
		tagsLoading = true;
		try {
			cashflowTags = (await listCashflowTags()).data;
		} catch {
			// Typing a tag still works without the list, so the dialog stays usable; it just
			// offers nothing to pick from.
			cashflowTags = [];
		} finally {
			tagsLoading = false;
		}
	}

	async function handleTag({ tag }: TagTransactionsValue) {
		const ids = tagRows.filter((row) => row.tag.trim() !== tag).map((row) => row.id);
		tagging = true;
		tagError = null;
		try {
			await tagCashflowTransactionsBySelection({ tag, ids });
			tagOpen = false;
			const picked = new Set(tagRows.map((row) => row.id));
			selectedIds = selectedIds.filter((id) => !picked.has(id));
			const count = `${ids.length} ${ids.length === 1 ? 'transaction' : 'transactions'}`;
			toast.success(tag ? `Tagged ${count} as “${tag}”` : `Removed the tag from ${count}`);
			void load(currentQuery());
			void loadAnalytics();
		} catch {
			tagError = 'Could not tag these transactions. Try again.';
		} finally {
			tagging = false;
		}
	}

	/**
	 * Puts ignored rows back in the ledger. The API records that the decision was made by
	 * hand, so no rule ignores them again — which is the whole point of restoring one.
	 */
	async function restore(ids: string[]) {
		if (ids.length === 0) return;
		try {
			await ignoreCashflowTransactionsBySelection({ ignored: false, ids });
			selectedIds = selectedIds.filter((id) => !ids.includes(id));
			toast.success(
				ids.length === 1
					? 'Put back in your ledger. Rules leave it alone from now on.'
					: `Put ${ids.length} transactions back in your ledger`
			);
			void load(currentQuery());
			void loadAnalytics();
		} catch {
			toast.error('Could not restore the transactions');
		}
	}

	/**
	 * Hands the selection to the rules page as a draft. The rule is still written and
	 * previewed in the one place that owns rules, rather than guessed at from here: the
	 * shared text of the selected descriptions is a starting point, not a decision.
	 */
	function makeRuleFromSelection() {
		if (selectedRows.length === 0) return;
		const contains = sharedPrefix(selectedRows.map((row) => row.description));
		void goto(`/cashflow/ignore-rules?contains=${encodeURIComponent(contains)}`);
	}

	/** The longest leading text every selected description shares, trimmed at a word. */
	function sharedPrefix(values: string[]): string {
		if (values.length === 0) return '';
		let prefix = values[0];
		for (const value of values.slice(1)) {
			while (prefix && !value.toLowerCase().startsWith(prefix.toLowerCase())) {
				prefix = prefix.slice(0, -1);
			}
		}
		const trimmed = prefix.trim();
		// A prefix that stops mid-word would make a rule nobody can read, so fall back to
		// the first description in full and let the preview show what it catches.
		return trimmed.length >= 3 ? trimmed : values[0];
	}

	const selectedRows = $derived(
		selectedIds.flatMap((id) => {
			const row = knownRows.get(id);
			return row ? [row] : [];
		})
	);

	// Select-all in the table header takes the page; past it, every row matching the filters is
	// one click away. Those are read in one request and selected by id, so every bulk action
	// keeps working on ids whatever page they were picked on.
	const pageSelected = $derived(
		rows.length > 0 && rows.every((row) => selectedIds.includes(row.id))
	);
	const allSelected = $derived(total > 0 && selectedIds.length >= total);

	async function selectAllMatching() {
		selectingAll = true;
		try {
			const result = await listCashflowTransactions({ ...currentQuery(), limit: total, offset: 0 });
			for (const row of result.data) knownRows.set(row.id, row);
			selectedIds = result.data.map((row) => row.id);
		} catch {
			toast.error('Could not select every transaction. Try again.');
		} finally {
			selectingAll = false;
		}
	}

	// The dialog offers "add to an existing item" first, so the items are read when it
	// opens rather than on every page load: most visits never mark anything. They are read
	// *before* opening, because the dialog picks its route once from the list it sees on
	// opening — arriving late would point an account that already has items at "new".
	async function openMark(selection: CashflowTransaction[]) {
		markRows = selection;
		markError = null;
		try {
			const overview = await getRecurringOverview();
			recurringItems = [...overview.expenses, ...overview.income];
		} catch {
			// Starting a new item still works without the list, so this is not reported as a
			// failure — the dialog simply offers only that route.
			recurringItems = [];
		}
		markOpen = true;
	}

	/** The row opened in the drawer goes to the same dialog as a selection of one. */
	function markDetailRow() {
		const row = detailRow;
		if (!row) return;
		detailOpen = false;
		void openMark([row]);
	}

	async function handleMarkRecurring(value: MarkRecurringValue) {
		const ids = markRows.map((row) => row.id);
		marking = true;
		markError = null;
		try {
			// A transaction belongs to one item at most; the API skips the ones that already do,
			// so the count it reports is the one worth repeating.
			let message: string;
			if (value.mode === 'existing') {
				const { linked_count } = await linkRecurringTransactions(value.itemId, { ids });
				const skipped = ids.length - linked_count;
				message = `Linked ${linked_count} ${linked_count === 1 ? 'transaction' : 'transactions'}`;
				if (skipped > 0) message += ` · ${skipped} already in another item`;
			} else {
				await createRecurringItem({
					name: value.name,
					direction: value.direction,
					rhythm: value.rhythm,
					ids
				});
				message = `${value.name} added to recurring`;
			}
			markOpen = false;
			selectedIds = selectedIds.filter((id) => !ids.includes(id));
			toast.success(message);
		} catch (err) {
			markError = markRecurringMessage(err);
		} finally {
			marking = false;
		}
	}

	// The refusals worth naming are the ones the reader can act on: a name already taken,
	// transactions that already belong to another item, one running the other way or ignored,
	// or an item that was ended and no longer takes transactions. The body's key tells them apart.
	function markRecurringMessage(err: unknown): string {
		if (!(err instanceof ApiError)) return 'Could not mark these as recurring. Try again.';
		const body = (err.body ?? {}) as Record<string, unknown>;
		if (err.status === 409 && 'ids' in body) {
			return 'These transactions already belong to another recurring item. Unlink them there first.';
		}
		if (err.status === 409) {
			return 'A recurring item with that name already exists. Pick another name, or add these to it.';
		}
		if (err.status === 422 && 'direction' in body) {
			return 'That item runs the other way. Pick an item of the same direction, or start a new one.';
		}
		if (err.status === 422 && 'ignored' in body) {
			return 'Ignored transactions are not linked to recurring items. Restore them first.';
		}
		if (err.status === 422) {
			return 'That item was ended, so it takes no new transactions.';
		}
		return 'Could not mark these as recurring. Try again.';
	}

	const tableMeta = $derived.by(() => {
		if (loading) return 'Loading…';
		if (error) return 'Could not load';
		const selected = selectedIds.length > 0 ? ` · ${selectedIds.length} selected` : '';
		return `${total} ${total === 1 ? 'row' : 'rows'}${selected}`;
	});

	// Imports need the vendors, which the cashflow page does not otherwise load, so the list
	// is fetched when an import is first opened rather than on every visit. A bank statement
	// opens the bank import; a DEGIRO statement is handed over from there to the brokerage one.
	async function openImport() {
		if (!vendorsLoaded) {
			try {
				const active = (await listVendors()).filter((v) => v.active);
				bankVendors = active.filter((v) => v.type === 'bank');
				brokerageVendors = active.filter((v) => v.type === 'brokerage');
				vendorsLoaded = true;
			} catch {
				// Retried on the next open; opening now would read as "no bank account".
				toast.error('Could not load your accounts. Try again.');
				return;
			}
		}
		importOpen = true;
	}

	// An import adds transactions, so it moves the ledger, the charts and the standing alike.
	function refreshAfterImport() {
		void load(currentQuery());
		void loadAnalytics();
		void loadStanding();
		// An import can switch a rule's harvest, and the rule names on the ledger's rows
		// come from the list this page holds.
		void loadIgnoreRules();
	}

	const cardGroups = [
		{ id: 'cashflow', label: 'Cashflow overview' },
		{ id: 'goal', label: 'Wealth goal' }
	];

	// Marking a purpose is an action on these rows, so it sits with the other row actions on the
	// ledger header. Three of them beside Tag, Import and Add would crowd the rule, so the
	// purposes themselves live in one menu.
	const selectionMarkItems: MenuItem[] = $derived([
		{
			label: 'Mark as income',
			disabled: markingPurpose,
			onSelect: () => handleMarkSelection('income')
		},
		{
			label: 'Mark as wealth',
			disabled: markingPurpose,
			onSelect: () => handleMarkSelection('wealth')
		},
		{ label: 'Clear purpose', disabled: markingPurpose, onSelect: () => handleMarkSelection('') }
	]);

	const filterMarkItems: MenuItem[] = $derived([
		{
			label: 'Mark as income',
			disabled: markingPurpose,
			onSelect: () => handleMarkFilter('income')
		},
		{
			label: 'Mark as wealth',
			disabled: markingPurpose,
			onSelect: () => handleMarkFilter('wealth')
		},
		{ label: 'Clear purpose', disabled: markingPurpose, onSelect: () => handleMarkFilter('') }
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
			<!-- What is in the list is a filter over these rows, so the ignored switch and the
			     page that owns the rules sit with the filters rather than with the actions. -->
			{#snippet filters()}
				<div class="flex items-center gap-2 text-sm text-slate-700">
					<Switch
						checked={showIgnored}
						aria-label="Show ignored rows"
						onchange={(event) => {
							showIgnored = (event.currentTarget as HTMLInputElement).checked;
							offset = 0;
							selectedIds = [];
						}}
					/>
					<span>Show ignored rows</span>
				</div>
				<a
					href="/cashflow/ignore-rules"
					class="inline-flex items-center gap-1 rounded-md text-sm text-sky-700 underline underline-offset-2 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
				>
					<Icon icon="heroicons:eye-slash" size="sm" />Ignore rules
				</a>
			{/snippet}
			{#snippet actions()}
				{#if selectedIds.length > 0}
					<Button variant="ruled" onclick={() => void openTag()}>
						<Icon icon="heroicons:tag" />
						Tag {selectedIds.length} selected
					</Button>
					<ActionMenu label="Mark {selectedIds.length} selected…" items={selectionMarkItems} />
					{#if showIgnored}
						<Button
							variant="ruled"
							onclick={() =>
								void restore(selectedRows.filter((row) => row.ignored).map((row) => row.id))}
						>
							<Icon icon="heroicons:arrow-uturn-left" />
							Restore
						</Button>
					{/if}
					<Button variant="ruled" onclick={makeRuleFromSelection}>
						<Icon icon="heroicons:eye-slash" />
						Make an ignore rule
					</Button>
					<Button variant="ruled" onclick={() => void openMark(selectedRows)}>
						<Icon icon="heroicons:arrow-path-rounded-square" />
						Mark as recurring
					</Button>
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
		{#if total > rows.length && (allSelected || pageSelected)}
			<div
				class="flex flex-wrap items-center gap-x-3 gap-y-1 border-b border-slate-200 bg-amber-50 px-4 py-2 text-sm text-slate-700"
				role="status"
			>
				{#if allSelected}
					<span>All {total} transactions are selected.</span>
					<button
						type="button"
						class="rounded-md text-sky-700 underline underline-offset-2 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
						onclick={() => (selectedIds = [])}
					>
						Clear selection
					</button>
				{:else}
					<span>All {rows.length} on this page are selected.</span>
					<button
						type="button"
						class="rounded-md text-sky-700 underline underline-offset-2 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none disabled:text-slate-400"
						disabled={selectingAll}
						onclick={() => void selectAllMatching()}
					>
						{selectingAll ? 'Selecting…' : `Select all ${total} transactions`}
					</button>
				{/if}
			</div>
		{/if}
		<CashflowTransactionsTable
			{rows}
			{loading}
			{error}
			{total}
			{showIgnored}
			{ignoreRules}
			onRestore={(row) => void restore([row.id])}
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
			onRetry={() => void load(currentQuery())}
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

<TagTransactionsDialog
	bind:open={tagOpen}
	selection={tagRows}
	tags={cashflowTags}
	{tagsLoading}
	saving={tagging}
	error={tagError}
	onSubmit={handleTag}
/>

<MarkRecurringDialog
	bind:open={markOpen}
	selection={markRows}
	items={recurringItems}
	saving={marking}
	error={markError}
	onSubmit={handleMarkRecurring}
/>

<CashflowImportDialog
	bind:open={importOpen}
	vendors={bankVendors}
	onFinished={refreshAfterImport}
	onBrokerageImport={() => (brokerageImportOpen = true)}
	onReviewIgnored={(id) => void goto(`/cashflow/imports/${id}`)}
/>

<ImportDialog
	bind:open={brokerageImportOpen}
	vendors={brokerageVendors}
	onFinished={refreshAfterImport}
	onGoToPortfolio={() => void goto('/portfolio')}
	onReviewIgnored={(id) => void goto(`/cashflow/imports/${id}`)}
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

<!-- An ignored row is not linked to recurring items, so it gets no button to try. -->
{#snippet detailActions()}
	<Button variant="ruled" onclick={markDetailRow}>
		<Icon icon="heroicons:arrow-path-rounded-square" />
		Mark as recurring
	</Button>
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
		actions={detailRow.ignored ? undefined : detailActions}
	/>
{/if}
