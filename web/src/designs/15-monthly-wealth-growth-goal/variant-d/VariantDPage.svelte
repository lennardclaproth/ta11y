<script lang="ts">
	// Variant D — "Board tabs": the analytics band gets a tab strip, and the board you pick also
	// decides what the panel underneath shows. Cashflow keeps its charts and its ledger; Wealth
	// goal gets the goal cards and the monthly standing. One switch, one page.
	import { untrack } from 'svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import PortalFrame from '../shared/PortalFrame.svelte';
	import CashflowCharts from '../shared/CashflowCharts.svelte';
	import TransactionsPanel from '../shared/TransactionsPanel.svelte';
	import GoalSettingPanel from '../shared/GoalSettingPanel.svelte';
	import RunningMonthPanel from '../shared/RunningMonthPanel.svelte';
	import StreakPanel from '../shared/StreakPanel.svelte';
	import NoGoalPanel from '../shared/NoGoalPanel.svelte';
	import GoalMonthEntry from '../shared/GoalMonthEntry.svelte';
	import SetGoalDialog from '../shared/SetGoalDialog.svelte';
	import { currentGoalPercent, goalAmount, goalMonths } from '../goal-data';

	type Props = {
		/** Which board the page opens on. */
		board?: 'cashflow' | 'goal';
		/** No goal set yet: the goal board asks for one first. */
		goalSet?: boolean;
		/** Open the goal dialog straight away (for the story that shows it). */
		dialogOpen?: boolean;
	};

	let { board = 'goal', goalSet = true, dialogOpen = false }: Props = $props();

	// The props only seed the prototype's starting state; they are not meant to stay in sync.
	let current = $state(untrack(() => board));
	let goalDialog = $state(untrack(() => dialogOpen));
	let goalPercent = $state(currentGoalPercent);
	let selectedIds = $state<string[]>(untrack(() => (board === 'cashflow' ? ['px-04', 'px-05'] : [])));

	const running = goalMonths[0];
	const showGoalBoard = $derived(current === 'goal' && goalSet);

	const tabs = [
		{ value: 'cashflow', label: 'Cashflow' },
		{ value: 'goal', label: 'Wealth goal' }
	];
</script>

<PortalFrame>
	<PageContentTemplate>
		{#snippet analytics()}
			<div class="flex flex-col gap-3">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<Tabs {tabs} bind:value={current} ariaLabel="Analytics board" />
					<Text size="xs" tone="muted">Two boards for now; which ones you see becomes a setting later.</Text>
				</div>

				{#if current === 'cashflow'}
					<CashflowCharts />
				{:else if goalSet}
					<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
						<AnalyticsCard title="Monthly goal">
							<GoalSettingPanel
								percent={goalPercent}
								amount={goalAmount(running)}
								onAdjust={() => (goalDialog = true)}
							/>
						</AnalyticsCard>
						<AnalyticsCard title={running.label} class="lg:col-span-2">
							<RunningMonthPanel month={running} onOpenUnassigned={() => (current = 'cashflow')} />
						</AnalyticsCard>
						<AnalyticsCard title="Streak">
							<StreakPanel />
						</AnalyticsCard>
					</div>
				{:else}
					<AnalyticsCard title="Wealth goal">
						<NoGoalPanel onSet={() => (goalDialog = true)} />
					</AnalyticsCard>
				{/if}
			</div>
		{/snippet}

		{#if showGoalBoard}
			<div
				class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
			>
				<h2 class="text-2xl">Monthly standing</h2>
				<Button size="sm" variant="ghost" intent="secondary" onclick={() => (current = 'cashflow')}>
					Open transactions
				</Button>
			</div>
			<div class="min-h-0 flex-1 overflow-auto px-4 pb-4">
				{#each goalMonths as month (month.month)}
					<GoalMonthEntry {month} onOpenUnassigned={() => (current = 'cashflow')} />
				{/each}
			</div>
		{:else}
			<TransactionsPanel bind:selectedIds />
		{/if}
	</PageContentTemplate>
</PortalFrame>

<SetGoalDialog bind:open={goalDialog} bind:percent={goalPercent} />
