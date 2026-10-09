<script lang="ts">
	// Design prototype for issue #16, variant A ("Ledger").
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
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import { chartColors } from '$lib/charts/theme';
	import {
		monthlyExpenseTotal,
		monthlyIncomeTotal,
		expenseTrendMonths,
		expenseTrendTotals,
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
	};

	let { situation = 'default' }: Props = $props();

	let group = $state('expenses');

	const counted = $derived(situation !== 'empty');
	const groups = $derived([
		{ value: 'expenses', label: `Expenses (${counted ? recurringExpenses.length : 0})` },
		{ value: 'income', label: `Income (${counted ? recurringIncome.length : 0})` },
		{ value: 'ended', label: `Ended (${counted ? recurringEnded.length : 0})` }
	]);

	const allRows = $derived(
		group === 'income' ? recurringIncome : group === 'ended' ? recurringEnded : recurringExpenses
	);

	const rows = $derived(situation === 'default' ? allRows : []);
	const loading = $derived(situation === 'loading');
	const error = $derived(situation === 'error' ? 'Could not load your recurring items.' : null);
	const emptyText = $derived(
		situation === 'no-matches'
			? 'No recurring items match this search.'
			: 'No recurring items yet. Point at a transaction in Cashflow to start one.'
	);

	const showSuggestions = $derived(situation === 'default' || situation === 'no-matches');

	// An account without recurring items has nothing to total: the analytics must not keep
	// claiming amounts the list below cannot show.
	const hasItems = $derived(situation !== 'empty');
	const trendLabels = $derived(hasItems ? expenseTrendMonths : []);
	const trendTotals = $derived(hasItems ? expenseTrendTotals : []);
	const expenseTotal = $derived(hasItems ? monthlyExpenseTotal : 0);
	const incomeTotal = $derived(hasItems ? monthlyIncomeTotal : 0);
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
	<div class="flex items-center justify-start gap-2">
		<Money
			amount={row.history[0]}
			currency="EUR"
			size="sm"
			weight="normal"
			class="text-slate-500"
		/>
		<Icon icon="heroicons:arrow-long-right" size="sm" class="text-slate-400" />
		<Sparkline
			data={row.history}
			tone="neutral"
			width={72}
			height={24}
			ariaLabel={`Amounts for ${row.name}, oldest to newest`}
		/>
	</div>
{/snippet}

{#snippet nextCell(row: RecurringItem)}
	{#if row.endedFrom}
		<span class="text-sm text-slate-500">Ended from {row.endedFrom}</span>
	{:else}
		<span class="text-sm text-slate-700">{nextExpectedLabel(row.nextExpected)}</span>
	{/if}
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Recurring"
			showSearch
			searchPlaceholder="Search name…"
			accountName="Account"
			accountEmail="you@example.com"
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
				<AnalyticsCard title="Recurring expenses per month" class="lg:col-span-2">
					<TimeSeriesChart
						height="h-44"
						labels={trendLabels}
						xTickFormat={shortMonth}
						{loading}
						ariaLabel="Total of your running recurring expenses per month"
						datasets={[
							{
								label: 'Recurring expenses',
								data: trendTotals,
								color: chartColors.negative,
								fill: true
							}
						]}
					/>
				</AnalyticsCard>
				<StatCard label="Expenses per month" amount={expenseTotal} currency="EUR" />
				<StatCard label="Income per month" amount={incomeTotal} currency="EUR" />
			</div>
		{/snippet}

		<!-- Ruled ledger header: title, the three groups, and the one action this page owns. -->
		<div
			class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
		>
			<div class="flex min-w-0 flex-wrap items-center gap-4">
				<h2 class="text-2xl">Recurring items</h2>
				<Tabs tabs={groups} bind:value={group} size="sm" ariaLabel="Recurring item groups" />
			</div>
			{#if showSuggestions}
				<Button shape="default">
					<Icon icon="heroicons:light-bulb" />
					Review suggestions ({recurringSuggestions.length})
				</Button>
			{/if}
		</div>

		{#if showSuggestions}
			<div class="shrink-0 border-b border-slate-200 bg-white px-4 py-3">
				<Alert intent="info">
					<Text as="span" size="sm">
						{recurringSuggestions.map((s) => s.name).join(' · ')} — possible recurring items from your
						history. Nothing is added until you confirm it.
					</Text>
				</Alert>
			</div>
		{/if}

		<DataTable
			{rows}
			{loading}
			{error}
			{emptyText}
			sortKey="next"
			sortDirection="asc"
			onRowClick={() => {}}
			columns={[
				{
					key: 'name',
					header: group === 'income' ? 'From' : 'To',
					sortKey: 'name',
					cell: nameCell
				},
				{ key: 'rhythm', header: 'Rhythm', sortKey: 'rhythm', width: 'w-32', cell: rhythmCell },
				{
					key: 'trend',
					header: 'Amount over time',
					width: 'w-52',
					cell: trendCell
				},
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
	</PageContentTemplate>
</AppShellTemplate>
