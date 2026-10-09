<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import StatCard from '$lib/components/organisms/stat-card/StatCard.svelte';
	import RecurringItemsTable from '$lib/components/organisms/recurring-items-table/RecurringItemsTable.svelte';
	import RecurringItemDrawer from '$lib/components/organisms/recurring-item-drawer/RecurringItemDrawer.svelte';
	import RecurringSuggestionsDrawer from '$lib/components/organisms/recurring-suggestions-drawer/RecurringSuggestionsDrawer.svelte';
	import EndRecurringDialog from '$lib/components/organisms/end-recurring-dialog/EndRecurringDialog.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import {
		confirmRecurringSuggestion,
		dismissRecurringSuggestion,
		endRecurringItem,
		getRecurringItem,
		getRecurringOverview,
		getRecurringSuggestions,
		unlinkRecurringTransaction
	} from '$lib/services/recurring';
	import { connectRealtime } from '$lib/services/realtime';
	import { accountStore } from '$lib/stores/account.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { recurringMonthTick } from '$lib/api/recurring';
	import { chartColors } from '$lib/charts/theme';
	import { parseQuery, serializeQuery, type QuerySchema } from '$lib/url/routeQuery';
	import { pushQuery } from '$lib/url/queryState';
	import type {
		RecurringItem,
		RecurringItemDetail,
		RecurringOverviewResponse,
		RecurringRhythm,
		RecurringSuggestion,
		RecurringTransaction
	} from '$lib/api/types';

	const schema: QuerySchema = {
		q: { type: 'string' },
		group: { type: 'string' }
	};

	const initial = parseQuery(page.url.searchParams, schema);

	let search = $state((initial.q as string) || '');
	let group = $state(((initial.group as string) || 'expenses') as 'expenses' | 'income' | 'ended');

	let overview = $state<RecurringOverviewResponse | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let suggestions = $state<RecurringSuggestion[]>([]);
	let suggestionsLoading = $state(true);
	let suggestionsError = $state<string | null>(null);
	let suggestionsOpen = $state(false);
	let pendingSuggestion = $state<string | null>(null);

	let itemOpen = $state(false);
	let item = $state<RecurringItemDetail | null>(null);
	let itemLoading = $state(false);
	let itemError = $state<string | null>(null);
	let unlinking = $state<string | null>(null);

	let endOpen = $state(false);
	let ending = $state(false);
	let endError = $state<string | null>(null);

	const rows = $derived.by(() => {
		if (!overview) return [];
		if (group === 'income') return overview.income;
		if (group === 'ended') return overview.ended;
		return overview.expenses;
	});

	// While the list is unknown the tabs carry no counts: a number here would be a claim
	// the page cannot yet back up.
	const known = $derived(!loading && !error && overview !== null);
	const groups = $derived([
		{ value: 'expenses', label: known ? `Expenses (${overview!.expenses.length})` : 'Expenses' },
		{ value: 'income', label: known ? `Income (${overview!.income.length})` : 'Income' },
		{ value: 'ended', label: known ? `Ended (${overview!.ended.length})` : 'Ended' }
	]);

	const runningCount = $derived(
		overview ? overview.expenses.length + overview.income.length : 0
	);
	const hasItems = $derived(
		overview !== null &&
			overview.expenses.length + overview.income.length + overview.ended.length > 0
	);
	// A search that found nothing is not an empty account, so the chart stays: the totals
	// describe the account either way.
	const searching = $derived(search.trim() !== '');
	const hasChart = $derived(known && (hasItems || searching) && overview!.series.length > 1);

	const showingIncome = $derived(group === 'income');
	const chartTitle = $derived(
		showingIncome ? 'Recurring income per month' : 'Recurring expenses per month'
	);

	// An empty account and a search that found nothing are different situations with
	// different ways out, so they never share a sentence.
	const emptyText = $derived(
		searching
			? 'No recurring items match this search.'
			: 'No recurring items yet. Mark a transaction in Cashflow to start one.'
	);

	async function load() {
		loading = true;
		error = null;
		try {
			overview = await getRecurringOverview({ q: search || undefined });
		} catch {
			// The amounts become unknown, not nil, so the totals disappear rather than
			// showing a zero that would be a lie.
			overview = null;
			error = 'Could not load your recurring items. Try again.';
		} finally {
			loading = false;
		}
	}

	async function loadSuggestions() {
		suggestionsLoading = true;
		suggestionsError = null;
		try {
			suggestions = (await getRecurringSuggestions()).data;
		} catch {
			suggestions = [];
			suggestionsError = 'Could not read your history right now. Try again.';
		} finally {
			suggestionsLoading = false;
		}
	}

	// Reload whenever the search changes (also covers the initial load).
	$effect(() => {
		void search;
		void load();
	});

	$effect(() => {
		void loadSuggestions();
	});

	onMount(() => {
		let realtime: { disconnect: () => void } | null = null;
		void accountStore.ensureLoaded().then(() => {
			if (!accountStore.hasAccount) return;
			// A finished import links its rows to confirmed items, so the page re-reads
			// rather than showing counts that the import has already moved on from.
			realtime = connectRealtime({
				accountId: accountStore.activeId,
				events: ['import.completed'],
				onRefresh: () => {
					void load();
					void loadSuggestions();
				}
			});
		});
		return () => realtime?.disconnect();
	});

	function syncUrl() {
		void pushQuery('/recurring', serializeQuery({ q: search, group }, schema), { replace: true });
	}

	async function openItem(row: RecurringItem) {
		item = null;
		itemError = null;
		itemLoading = true;
		itemOpen = true;
		try {
			item = await getRecurringItem(row.id);
		} catch {
			itemError = 'Could not load this recurring item. Try again.';
		} finally {
			itemLoading = false;
		}
	}

	async function handleUnlink(transaction: RecurringTransaction) {
		if (!item) return;
		unlinking = transaction.id;
		try {
			await unlinkRecurringTransaction(item.id, transaction.id);
			toast.success('Transaction unlinked — it stays in Cashflow');
			item = await getRecurringItem(item.id);
			void load();
		} catch {
			toast.error('Could not unlink the transaction');
		} finally {
			unlinking = null;
		}
	}

	async function handleEnd(month: string) {
		if (!item) return;
		ending = true;
		endError = null;
		try {
			await endRecurringItem(item.id, { from: month });
			endOpen = false;
			itemOpen = false;
			toast.success(`${item.name} ends from ${month}`);
			group = 'ended';
			syncUrl();
			void load();
		} catch {
			endError = 'Could not end this item. Try again.';
		} finally {
			ending = false;
		}
	}

	async function handleConfirm(
		suggestion: RecurringSuggestion,
		name: string,
		rhythm: RecurringRhythm
	) {
		pendingSuggestion = suggestion.match_key;
		try {
			await confirmRecurringSuggestion({
				match_key: suggestion.match_key,
				name,
				direction: suggestion.direction,
				rhythm
			});
			suggestions = suggestions.filter((entry) => entry.match_key !== suggestion.match_key);
			toast.success(`${name} added, with ${suggestion.matches} transactions`);
			void load();
		} catch {
			toast.error('Could not confirm this suggestion');
		} finally {
			pendingSuggestion = null;
		}
	}

	async function handleDismiss(suggestion: RecurringSuggestion) {
		pendingSuggestion = suggestion.match_key;
		try {
			await dismissRecurringSuggestion({
				match_key: suggestion.match_key,
				direction: suggestion.direction
			});
			suggestions = suggestions.filter((entry) => entry.match_key !== suggestion.match_key);
			toast.info('Dismissed — this pattern will not be suggested again');
		} catch {
			toast.error('Could not dismiss this suggestion');
		} finally {
			pendingSuggestion = null;
		}
	}
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Recurring"
			showSearch
			searchValue={search}
			searchPlaceholder="Search name…"
			onSearch={(q) => {
				search = q;
				syncUrl();
			}}
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			{#if error}
				<!-- No totals rather than a zero that would be a lie: the amounts are unknown, not nil. -->
				<section
					class="flex flex-wrap items-center justify-between gap-3 border-t border-slate-400 py-3"
				>
					<Text as="span" size="sm" tone="muted">
						Totals are unavailable while the list cannot be loaded.
					</Text>
					<Button size="sm" variant="outline" shape="default" onclick={() => void load()}>
						<Icon icon="heroicons:arrow-path" />
						Try again
					</Button>
				</section>
			{:else}
				<div class={hasChart ? 'grid grid-cols-2 gap-3 lg:grid-cols-4' : 'grid grid-cols-2 gap-3'}>
					<!-- The chart is context, not the answer. On a phone it would push every item
					     below the fold, so there it steps aside and the two totals lead. An account
					     without items has no line to draw, so there it is absent rather than empty. -->
					{#if hasChart}
						<AnalyticsCard title={chartTitle} class="hidden lg:col-span-2 lg:block">
							<TimeSeriesChart
								height="h-44"
								labels={overview!.series.map((point) => point.month)}
								xTickFormat={recurringMonthTick}
								{loading}
								ariaLabel={`Total of your running ${showingIncome ? 'recurring income' : 'recurring expenses'} per month`}
								datasets={[
									{
										label: chartTitle,
										data: overview!.series.map((point) =>
											scaledToNumber(showingIncome ? point.incomeCents : point.expenseCents)
										),
										color: showingIncome ? chartColors.positive : chartColors.negative,
										fill: true
									}
								]}
							/>
						</AnalyticsCard>
					{/if}

					{#if loading}
						{#each ['expenses', 'income'] as key (key)}
							<section class="flex min-w-0 flex-col gap-2 border-t border-slate-400 py-3">
								<Skeleton width="9rem" />
								<Skeleton variant="rect" width="7rem" height="2rem" />
							</section>
						{/each}
					{:else}
						<StatCard
							label="Expenses per month"
							amount={scaledToNumber(overview?.monthlyExpenseCents ?? 0)}
							currency="EUR"
						/>
						<StatCard
							label="Income per month"
							amount={scaledToNumber(overview?.monthlyIncomeCents ?? 0)}
							currency="EUR"
						/>
					{/if}
				</div>

				<Text as="p" size="xs" tone="muted" class="mt-2">
					{#if loading}
						Loading your recurring items…
					{:else if runningCount > 0}
						Monthly equivalent of {runningCount} running items — a quarterly amount counts as a third,
						a yearly amount as a twelfth.
					{:else}
						No running items yet, so there is nothing to total.
					{/if}
				</Text>
			{/if}
		{/snippet}

		<LedgerToolbar
			title="Recurring items"
			actionLabel={known && suggestions.length > 0
				? `Review suggestions (${suggestions.length})`
				: undefined}
			actionIcon="heroicons:light-bulb"
			onAdd={() => (suggestionsOpen = true)}
		>
			<Tabs
				tabs={groups}
				bind:value={
					() => group,
					(next) => {
						group = next as 'expenses' | 'income' | 'ended';
						syncUrl();
					}
				}
				size="sm"
				ariaLabel="Recurring item groups"
			/>
		</LedgerToolbar>

		<!-- Named, so the strip says what the button would open. It stays out of a search
		     result, where listing items that are not in the list below would only confuse. -->
		{#if known && suggestions.length > 0 && !searching}
			<div
				class="flex shrink-0 flex-wrap items-center gap-x-2 gap-y-1 border-b border-slate-200 bg-white px-4 py-2 text-sm"
			>
				<Icon icon="heroicons:light-bulb" size="sm" class="text-slate-500" />
				<span class="hidden text-slate-700 sm:inline">
					{suggestions.map((suggestion) => suggestion.name).join(' · ')}
				</span>
				<span class="hidden text-slate-500 sm:inline">
					— possible recurring items from your history. Nothing is added until you confirm it.
				</span>
				<span class="text-slate-700 sm:hidden">
					{suggestions.length} possible recurring items from your history
				</span>
			</div>
		{/if}

		<RecurringItemsTable {rows} {group} {loading} {error} {emptyText} onOpen={openItem} />
	</PageContentTemplate>
</AppShellTemplate>

<RecurringSuggestionsDrawer
	bind:open={suggestionsOpen}
	{suggestions}
	loading={suggestionsLoading}
	error={suggestionsError}
	pending={pendingSuggestion}
	onConfirm={handleConfirm}
	onDismiss={handleDismiss}
/>

<RecurringItemDrawer
	bind:open={itemOpen}
	{item}
	loading={itemLoading}
	error={itemError}
	{unlinking}
	onUnlink={handleUnlink}
	onEnd={() => {
		endError = null;
		endOpen = true;
	}}
/>

{#if item}
	<EndRecurringDialog
		bind:open={endOpen}
		name={item.name}
		saving={ending}
		error={endError}
		onConfirm={handleEnd}
	/>
{/if}
