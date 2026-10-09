<script lang="ts">
	// Design prototype for issue #16 — final design, built on variant A ("Ledger").
	// Not a route: it composes existing components to show the intended page.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import StatCard from '$lib/components/organisms/stat-card/StatCard.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { chartColors } from '$lib/charts/theme';
	import AmountTrend from './AmountTrend.svelte';
	import RecurringList from './RecurringList.svelte';
	import SuggestionsDrawer from './SuggestionsDrawer.svelte';
	import RecurringItemDrawer from './RecurringItemDrawer.svelte';
	import EndRecurringDialog from './EndRecurringDialog.svelte';
	import {
		monthlyExpenseTotal,
		monthlyIncomeTotal,
		expenseTrendMonths,
		expenseTrendTotals,
		incomeTrendTotals,
		openedItem,
		recurringEnded,
		recurringExpenses,
		recurringIncome,
		recurringSuggestions,
		rhythmLabels,
		type RecurringItem
	} from '../recurring.fixture';
	import { nextExpectedLabel, shortMonth } from '../recurring.format';

	type Props = {
		/** Which situation the prototype shows. */
		situation?: 'default' | 'loading' | 'error' | 'empty' | 'no-matches';
		/** Which group is open. */
		group?: string;
		/** Secondary surfaces, opened by the stories that document them. */
		suggestionsOpen?: boolean;
		itemOpen?: boolean;
		endOpen?: boolean;
	};

	let {
		situation = 'default',
		group = $bindable('expenses'),
		suggestionsOpen = $bindable(false),
		itemOpen = $bindable(false),
		endOpen = $bindable(false)
	}: Props = $props();

	const hasItems = $derived(situation !== 'empty');
	const loading = $derived(situation === 'loading');
	const failed = $derived(situation === 'error');

	const counts = $derived({
		expenses: hasItems ? recurringExpenses.length : 0,
		income: hasItems ? recurringIncome.length : 0,
		ended: hasItems ? recurringEnded.length : 0
	});

	// While the list is unknown, the tabs carry no counts: a number here would be a claim the page
	// cannot yet back up.
	const known = $derived(!loading && !failed);
	const groups = $derived([
		{ value: 'expenses', label: known ? `Expenses (${counts.expenses})` : 'Expenses' },
		{ value: 'income', label: known ? `Income (${counts.income})` : 'Income' },
		{ value: 'ended', label: known ? `Ended (${counts.ended})` : 'Ended' }
	]);

	// The chart follows the open group, so the totals above never describe a list below them that
	// is showing something else. Ended items keep the expense series for context.
	const showingIncome = $derived(group === 'income');
	const chartTitle = $derived(
		showingIncome ? 'Recurring income per month' : 'Recurring expenses per month'
	);

	// Soonest first: the page answers "what comes off next" before it answers anything else.
	// Ended items have no expectation, so they read newest-ended first instead.
	const allRows = $derived.by(() => {
		if (group === 'ended') {
			return [...recurringEnded].sort((a, b) => (a.endedFrom! < b.endedFrom! ? 1 : -1));
		}
		const source = group === 'income' ? recurringIncome : recurringExpenses;
		return [...source].sort((a, b) => (a.nextExpected! < b.nextExpected! ? -1 : 1));
	});

	const rows = $derived(situation === 'default' ? allRows : []);
	const error = $derived(failed ? 'Could not load your recurring items. Try again.' : null);

	// An empty account and a search that found nothing are different situations with different
	// ways out, so they never share a sentence.
	const emptyText = $derived(
		situation === 'no-matches'
			? 'No recurring items match this search.'
			: 'No recurring items yet. Mark a transaction in Cashflow to start one.'
	);

	const showSuggestionAction = $derived(situation === 'default' || situation === 'no-matches');
	const runningCount = $derived(counts.expenses + counts.income);
</script>

{#snippet nameCell(row: RecurringItem)}
	<div class="flex min-w-0 flex-col">
		<span class="truncate font-medium text-slate-900">{row.name}</span>
		<span class="text-xs text-slate-500">{row.linked} transactions</span>
	</div>
{/snippet}

{#snippet rhythmCell(row: RecurringItem)}
	<Badge intent="neutral" variant="soft" size="sm">{rhythmLabels[row.rhythm]}</Badge>
{/snippet}

{#snippet amountCell(row: RecurringItem)}
	<Money amount={row.amount} currency="EUR" size="sm" />
{/snippet}

<!-- Oldest amount first, then the shape of the amounts since: a reading of what happened, in a
     tone that carries no verdict. A rising expense is not an error, a falling one not a success. -->
{#snippet trendCell(row: RecurringItem)}
	<div class="flex items-center gap-2">
		<Money
			amount={row.history[0]}
			currency="EUR"
			size="sm"
			weight="normal"
			class="w-16 shrink-0 text-right text-slate-500"
		/>
		<Icon icon="heroicons:arrow-long-right" size="sm" class="shrink-0 text-slate-400" />
		<AmountTrend data={row.history} ariaLabel={`Amounts for ${row.name}, oldest to newest`} />
	</div>
{/snippet}

{#snippet nextCell(row: RecurringItem)}
	{#if row.endedFrom}
		<span class="text-sm text-slate-500">Ended from {row.endedFrom}</span>
	{:else}
		<span class="text-sm whitespace-nowrap text-slate-700">
			{nextExpectedLabel(row.nextExpected)}
		</span>
	{/if}
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Recurring"
			showSearch
			searchValue={situation === 'no-matches' ? 'cloud' : ''}
			searchPlaceholder="Search name…"
			accountName="Account"
			accountEmail="you@example.com"
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			{#if failed}
				<!-- No totals rather than a zero that would be a lie: the amounts are unknown, not nil. -->
				<section class="flex flex-wrap items-center justify-between gap-3 border-t border-slate-400 py-3">
					<Text as="span" size="sm" tone="muted">
						Totals are unavailable while the list cannot be loaded.
					</Text>
					<Button size="sm" variant="outline" shape="default">
						<Icon icon="heroicons:arrow-path" />
						Try again
					</Button>
				</section>
			{:else}
				<div class={hasItems ? 'grid grid-cols-2 gap-3 lg:grid-cols-4' : 'grid grid-cols-2 gap-3'}>
					<!-- The chart is context, not the answer. On a phone it would push every item below
					     the fold, so there it steps aside and the two totals lead. An account without
					     items has no line to draw, so there it is absent rather than empty. -->
					{#if hasItems}
						<AnalyticsCard title={chartTitle} class="hidden lg:col-span-2 lg:block">
							<TimeSeriesChart
								height="h-44"
								labels={expenseTrendMonths}
								xTickFormat={shortMonth}
								{loading}
								ariaLabel={`Total of your running ${showingIncome ? 'recurring income' : 'recurring expenses'} per month`}
								datasets={[
									{
										label: chartTitle,
										data: showingIncome ? incomeTrendTotals : expenseTrendTotals,
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
							amount={hasItems ? monthlyExpenseTotal : 0}
							currency="EUR"
						/>
						<StatCard
							label="Income per month"
							amount={hasItems ? monthlyIncomeTotal : 0}
							currency="EUR"
						/>
					{/if}
				</div>

				<Text as="p" size="xs" tone="muted" class="mt-2">
					{#if loading}
						Loading your recurring items…
					{:else if hasItems}
						Monthly equivalent of {runningCount} running items — a quarterly amount counts as a third,
						a yearly amount as a twelfth.
					{:else}
						No running items yet, so there is nothing to total.
					{/if}
				</Text>
			{/if}
		{/snippet}

		<!-- Ruled ledger header: title, the three groups, and the one action this page owns. -->
		<div
			class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
		>
			<div class="flex min-w-0 flex-wrap items-center gap-4">
				<h2 class="text-2xl">Recurring items</h2>
				<Tabs tabs={groups} bind:value={group} size="sm" ariaLabel="Recurring item groups" />
			</div>
			{#if showSuggestionAction}
				<Button shape="default" onclick={() => (suggestionsOpen = true)}>
					<Icon icon="heroicons:light-bulb" />
					Review suggestions ({recurringSuggestions.length})
				</Button>
			{/if}
		</div>

		<!-- Named, so the strip says what the button would open. It stays out of a search result,
		     where listing items that are not in the list below would only confuse. -->
		{#if situation === 'default'}
			<div
				class="flex shrink-0 flex-wrap items-center gap-x-2 gap-y-1 border-b border-slate-200 bg-white px-4 py-2 text-sm"
			>
				<Icon icon="heroicons:light-bulb" size="sm" class="text-slate-500" />
				<span class="hidden text-slate-700 sm:inline">
					{recurringSuggestions.map((s) => s.name).join(' · ')}
				</span>
				<span class="hidden text-slate-500 sm:inline">
					— possible recurring items from your history. Nothing is added until you confirm it.
				</span>
				<span class="text-slate-700 sm:hidden">
					{recurringSuggestions.length} possible recurring items from your history
				</span>
			</div>
		{/if}

		<DataTable
			class="hidden lg:flex"
			{rows}
			{loading}
			{error}
			{emptyText}
			sortKey="next"
			sortDirection="asc"
			onRowClick={() => (itemOpen = true)}
			columns={[
				{
					key: 'name',
					header: group === 'income' ? 'From' : 'To',
					sortKey: 'name',
					cell: nameCell
				},
				{ key: 'rhythm', header: 'Rhythm', sortKey: 'rhythm', width: 'w-28', cell: rhythmCell },
				{ key: 'trend', header: 'Amount over time', width: 'w-56', cell: trendCell },
				{
					key: 'amount',
					header: 'Last amount',
					sortKey: 'amount',
					align: 'right',
					width: 'w-32',
					cell: amountCell
				},
				{
					key: 'next',
					header: group === 'ended' ? 'Status' : 'Next expected',
					sortKey: 'next',
					width: 'w-56',
					cell: nextCell
				}
			]}
		/>

		<RecurringList
			class="lg:hidden"
			{rows}
			{loading}
			{error}
			{emptyText}
			onOpen={() => (itemOpen = true)}
		/>
	</PageContentTemplate>
</AppShellTemplate>

<SuggestionsDrawer bind:open={suggestionsOpen} />
<RecurringItemDrawer bind:open={itemOpen} item={openedItem} onEnd={() => (endOpen = true)} />
<EndRecurringDialog bind:open={endOpen} item={openedItem} />
