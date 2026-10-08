<script lang="ts">
	// Variant B — "Goal page": the monthly standing gets its own destination next to Cashflow,
	// Assets and Portfolio, read as a column of month entries instead of a table.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import GoalMonthEntry from './GoalMonthEntry.svelte';
	import StreakStrip from './StreakStrip.svelte';
	import { bestStreak, currentStreak, goalMonths } from '../goal-data';

	const running = goalMonths[0];
	const figureClass = 'font-heading text-3xl leading-tight tracking-tight text-slate-900';
</script>

<!-- The shell is viewport-tall in the app; in the prototype it grows with its content so a
     screenshot shows the whole screen instead of the first fold. -->
<AppShellTemplate class="h-auto! min-h-dvh">
	{#snippet top()}
		<TopNavbar
			title="Wealth goal"
			accountName="Design preview"
			accountEmail="preview@example.com"
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			<div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
				<AnalyticsCard title="Streak" class="lg:col-span-2">
					<div class="flex flex-col gap-3">
						<span class={figureClass}>{currentStreak} months in a row</span>
						<StreakStrip months={goalMonths} />
						<Text size="sm" tone="muted">
							Best run so far: {bestStreak} months. July is still running, so it does not count yet.
						</Text>
					</div>
				</AnalyticsCard>

				<AnalyticsCard title="Monthly goal">
					<div class="flex flex-col gap-3">
						<span class={figureClass}>{running.goalPercent}% of income</span>
						<Text size="sm" tone="muted">
							Applies from July 2026. Earlier months keep the goal they had.
						</Text>
						<Button size="sm" variant="outline" class="w-fit">Adjust goal</Button>
					</div>
				</AnalyticsCard>
			</div>
		{/snippet}

		<div
			class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
		>
			<h2 class="text-2xl">Monthly standing</h2>
			<Button size="sm" variant="ghost" intent="secondary">Open Cashflow</Button>
		</div>

		<div class="min-h-0 flex-1 overflow-auto px-4 pb-4">
			{#each goalMonths as month (month.month)}
				<GoalMonthEntry {month} />
			{/each}
		</div>
	</PageContentTemplate>
</AppShellTemplate>
