<script lang="ts">
	// Final design for #15 — "Card rail", consolidated from variant E.
	//
	// The analytics band of Cashflow becomes one horizontally scrolling row of cards, grouped into
	// "Cashflow overview" (the three charts that are there today) and "Wealth goal" (the running
	// month against your goal, your streak, and the monthly standing). The ledger underneath never
	// changes, so you mark transactions and read the score without leaving the page. Which cards
	// you see, and in what order, becomes a setting later; for now the two groups are fixed.
	import { untrack } from 'svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import PortalFrame from '../shared/PortalFrame.svelte';
	import CardRail from './CardRail.svelte';
	import NetTrendChart from '../shared/NetTrendChart.svelte';
	import TagDonut from '../shared/TagDonut.svelte';
	import TransactionsPanel from '../shared/TransactionsPanel.svelte';
	import RunningMonthPanel from '../shared/RunningMonthPanel.svelte';
	import StreakPanel from '../shared/StreakPanel.svelte';
	import NoGoalPanel from '../shared/NoGoalPanel.svelte';
	import MonthStandingRow from '../shared/MonthStandingRow.svelte';
	import GoalLoadingPanel from '../shared/GoalLoadingPanel.svelte';
	import GoalErrorPanel from '../shared/GoalErrorPanel.svelte';
	import SetGoalDialog from '../shared/SetGoalDialog.svelte';
	import { currentGoalPercent, goalMonths, goalTransactions, type GoalMonth } from '../goal-data';

	type Props = {
		/** `data-card` the rail opens on; the real page simply starts at the left. */
		startAt?: string;
		/** `no-goal` is the first run: nothing to score against yet. */
		state?: 'ready' | 'loading' | 'error' | 'no-goal';
		/** Preselect rows so the bulk marking actions are on screen. */
		selection?: boolean;
		/** Open the goal dialog. */
		dialog?: boolean;
		/** Arrive from "n transactions are not assigned yet" on the standing. */
		focusMonth?: string | null;
		/** The Purpose filter matched nothing. */
		noMatches?: boolean;
	};

	let {
		startAt = 'month',
		state: view = 'ready',
		selection = false,
		dialog = false,
		focusMonth = null,
		noMatches = false
	}: Props = $props();

	// The props seed the prototype's starting state once; `untrack` keeps that out of the reactive
	// graph, which is what the props are for here.
	let goalDialog = $state(untrack(() => dialog));
	let goalPercent = $state(currentGoalPercent);
	let selectedIds = $state<string[]>(untrack(() => (selection ? ['px-04', 'px-05', 'px-07'] : [])));
	let scope = $state<string | null>(untrack(() => focusMonth));
	let purposeFilter = $state<string[]>(untrack(() => (focusMonth || noMatches ? ['none'] : [])));

	const running = goalMonths[0];
	const completed = goalMonths.slice(1);

	const groups = [
		{ id: 'cashflow', label: 'Cashflow overview' },
		{ id: 'goal', label: 'Wealth goal' }
	];

	// The standing sends you to exactly the transactions that keep a month incomplete; the ledger
	// header says which month you are looking at so the jump is undoable by eye.
	function openUnassigned(month: GoalMonth) {
		scope = month.label;
		purposeFilter = ['none'];
		selectedIds = [];
	}

	const unassignedRows = $derived(goalTransactions.filter((row) => row.purpose === null));
	// First run: there is no goal yet and nothing has been pointed at either.
	const blankRows = $derived(goalTransactions.map((row) => ({ ...row, purpose: null })));
	const rows = $derived(
		noMatches
			? []
			: view === 'no-goal'
				? blankRows
				: scope
					? unassignedRows
					: goalTransactions
	);

	// One width vocabulary for the rail: a full-width card on a phone, a fixed column on desktop.
	// The three wealth-goal cards add up to a full band at `lg`, so jumping to the group lands
	// flush against the left edge instead of leaving a sliver of the chart behind it.
	const card = 'w-[84vw] shrink-0 snap-start sm:w-[20rem]';
	const wideCard = 'w-[84vw] shrink-0 snap-start sm:w-[24rem] lg:w-[32rem]';
	// A group of one card has to fill the band itself, or it can never reach the left edge.
	const soloCard = 'w-[84vw] shrink-0 snap-start sm:w-full';
</script>

<PortalFrame>
	<PageContentTemplate>
		{#snippet analytics()}
			<CardRail {groups} {startAt}>
				<div data-card="trend" data-group="cashflow" class={wideCard}>
					<AnalyticsCard title="Net trend">
						<NetTrendChart loading={view === 'loading'} />
					</AnalyticsCard>
				</div>
				<div data-card="incoming" data-group="cashflow" class={card}>
					<AnalyticsCard title="Incoming">
						<TagDonut direction="in" loading={view === 'loading'} />
					</AnalyticsCard>
				</div>
				<div data-card="outgoing" data-group="cashflow" class={card}>
					<AnalyticsCard title="Outgoing">
						<TagDonut direction="out" loading={view === 'loading'} />
					</AnalyticsCard>
				</div>

				{#if view === 'no-goal'}
					<div data-card="month" data-group="goal" class={soloCard}>
						<AnalyticsCard title="Wealth goal">
							<NoGoalPanel onSet={() => (goalDialog = true)} />
						</AnalyticsCard>
					</div>
				{:else if view === 'error'}
					<div data-card="month" data-group="goal" class={soloCard}>
						<AnalyticsCard title="Wealth goal">
							<GoalErrorPanel />
						</AnalyticsCard>
					</div>
				{:else}
					<div data-card="month" data-group="goal" class={wideCard}>
						<AnalyticsCard title="{running.label} against your goal">
							{#if view === 'loading'}
								<GoalLoadingPanel shape="month" />
							{:else}
								<RunningMonthPanel
									month={running}
									stacked
									onOpenUnassigned={openUnassigned}
									onAdjust={() => (goalDialog = true)}
								/>
							{/if}
						</AnalyticsCard>
					</div>
					<div data-card="streak" data-group="goal" class={card}>
						<AnalyticsCard title="Streak">
							{#if view === 'loading'}
								<GoalLoadingPanel shape="streak" />
							{:else}
								<StreakPanel />
							{/if}
						</AnalyticsCard>
					</div>
					<div data-card="standing" data-group="goal" class={wideCard}>
						<AnalyticsCard title="Monthly standing">
							{#if view === 'loading'}
								<GoalLoadingPanel shape="rows" />
							{:else}
								<!-- Four months fit a phone card without making the whole rail taller than the
								     card you are reading; the rest are one tap away. -->
								{#each completed as month, index (month.month)}
									<div class={index >= 4 ? 'hidden sm:block' : ''}>
										<MonthStandingRow
											{month}
											divider={index > 0}
											onOpenUnassigned={openUnassigned}
										/>
									</div>
								{/each}
								{#if completed.length > 4}
									<span class="sm:hidden">
										<Button size="sm" variant="ghost" intent="secondary" class="mt-1 px-0">
											Show {completed.length - 4} earlier months
										</Button>
									</span>
								{/if}
							{/if}
						</AnalyticsCard>
					</div>
				{/if}
			</CardRail>
		{/snippet}

		<TransactionsPanel
			{rows}
			{scope}
			loading={view === 'loading'}
			emptyText={noMatches
				? 'No transactions match your filters'
				: 'No transactions yet. Add one to start your ledger.'}
			bind:selectedIds
			bind:purposeFilter
		/>
	</PageContentTemplate>
</PortalFrame>

<SetGoalDialog bind:open={goalDialog} bind:percent={goalPercent} />
