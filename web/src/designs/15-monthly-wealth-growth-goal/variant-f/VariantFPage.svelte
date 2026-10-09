<script lang="ts">
	// Variant F — "Section stack": the analytics band keeps reading downwards. Cashflow overview is
	// the first section, Wealth goal the second, and the monthly standing simply continues inside
	// it. You scroll from the one to the other; the ledger stays pinned underneath.
	import { untrack } from 'svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { zClasses } from '$lib/styles/z-index';
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
		/** Which section the stack is scrolled to when the page opens. */
		startAt?: 'overview' | 'goal';
		goalSet?: boolean;
	};

	let { startAt = 'goal', goalSet = true }: Props = $props();

	let stack = $state<HTMLDivElement | null>(null);
	// The prop only seeds where the prototype opens; it is not meant to stay in sync.
	let active = $state(untrack(() => startAt));
	let goalDialog = $state(false);
	let goalPercent = $state(currentGoalPercent);
	let selectedIds = $state<string[]>([]);

	const running = goalMonths[0];
	const past = goalMonths.slice(1);

	// The prototype opens on a section so a screenshot shows the goal half; in the app the stack
	// simply starts at the top.
	$effect(() => {
		const region = stack;
		if (!region) return;
		const section = region.querySelector(`[data-section="${startAt}"]`);
		if (section instanceof HTMLElement) region.scrollTop = section.offsetTop - region.offsetTop;
	});

	function show(section: 'overview' | 'goal') {
		active = section;
		const target = stack?.querySelector(`[data-section="${section}"]`);
		if (target instanceof HTMLElement && stack) {
			stack.scrollTo({ top: target.offsetTop - stack.offsetTop, behavior: 'smooth' });
		}
	}

	function onScroll() {
		const region = stack;
		const goal = region?.querySelector('[data-section="goal"]');
		if (!region || !(goal instanceof HTMLElement)) return;
		active = region.scrollTop >= goal.offsetTop - region.offsetTop - 24 ? 'goal' : 'overview';
	}

	const linkBase =
		'rounded-md px-2 py-1 text-sm transition-colors duration-150 ease-out focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none';
	const linkActive = 'bg-amber-100 text-slate-900';
	const linkIdle = 'text-slate-600 hover:text-slate-900';
	const sectionLabel = 'sticky top-0 bg-taupe-100 pb-2 text-slate-500';
</script>

<PortalFrame>
	<PageContentTemplate>
		{#snippet analytics()}
			<div class="flex flex-col gap-2">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<nav class="flex items-center gap-1" aria-label="Analytics sections">
						<button
							type="button"
							class={[linkBase, active === 'overview' ? linkActive : linkIdle].join(' ')}
							aria-current={active === 'overview' ? 'true' : undefined}
							onclick={() => show('overview')}>Cashflow overview</button
						>
						<span class="text-slate-300" aria-hidden="true">·</span>
						<button
							type="button"
							class={[linkBase, active === 'goal' ? linkActive : linkIdle].join(' ')}
							aria-current={active === 'goal' ? 'true' : undefined}
							onclick={() => show('goal')}>Wealth goal</button
						>
					</nav>
					<Text size="xs" tone="muted">
						Two sections for now; which ones you see becomes a setting later.
					</Text>
				</div>

				<div
					bind:this={stack}
					onscroll={onScroll}
					class="flex flex-col gap-5 lg:max-h-[30rem] lg:overflow-y-auto lg:pr-2"
				>
					<section data-section="overview">
						<Heading level="h2" size="sm" uppercase class={[sectionLabel, zClasses.stickyHeader].join(' ')}>
							Cashflow overview
						</Heading>
						<CashflowCharts />
					</section>

					<section data-section="goal" class="flex flex-col">
						<Heading level="h2" size="sm" uppercase class={[sectionLabel, zClasses.stickyHeader].join(' ')}>
							Wealth goal
						</Heading>

						{#if goalSet}
							<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
								<AnalyticsCard title="Monthly goal">
									<GoalSettingPanel
										percent={goalPercent}
										amount={goalAmount(running)}
										onAdjust={() => (goalDialog = true)}
									/>
								</AnalyticsCard>
								<AnalyticsCard title={running.label} class="lg:col-span-2">
									<RunningMonthPanel month={running} />
								</AnalyticsCard>
								<AnalyticsCard title="Streak">
									<StreakPanel />
								</AnalyticsCard>
							</div>

							<div class="mt-5">
								<Heading level="h3" size="sm" class="mb-1 text-slate-700">Earlier months</Heading>
								{#each past as month (month.month)}
									<GoalMonthEntry {month} />
								{/each}
							</div>
						{:else}
							<AnalyticsCard title="No goal yet">
								<NoGoalPanel onSet={() => (goalDialog = true)} />
							</AnalyticsCard>
						{/if}
					</section>
				</div>
			</div>
		{/snippet}

		<TransactionsPanel bind:selectedIds />
	</PageContentTemplate>
</PortalFrame>

<SetGoalDialog bind:open={goalDialog} bind:percent={goalPercent} />
