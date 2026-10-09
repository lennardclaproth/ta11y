<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import RunningMonthCard from './RunningMonthCard.svelte';
	import StreakCard from './StreakCard.svelte';
	import MonthlyStandingCard from './MonthlyStandingCard.svelte';
	import NoGoalCard from './NoGoalCard.svelte';
	import GoalErrorCard from './GoalErrorCard.svelte';
	import GoalLoadingCard from './GoalLoadingCard.svelte';
	import { numberToScaled as s } from '$lib/api/money';
	import { monthLabel } from '$lib/api/wealthgoal';
	import type { MonthStanding } from '$lib/api/types';

	const { Story } = defineMeta({
		title: 'Organisms/WealthGoalCards',
		component: RunningMonthCard,
		tags: ['autodocs']
	});

	const month = (
		key: string,
		result: MonthStanding['result'],
		income: number,
		contributed: number,
		unassigned = 0
	): MonthStanding => ({
		month: key,
		income_cents: s(income),
		contributed_cents: s(contributed),
		goal_percent: 30,
		result,
		unassigned_count: unassigned
	});

	const running = month('2026-07-01', 'in_progress', 5000, 900, 5);
	const finished: MonthStanding[] = [
		month('2026-06-01', 'met', 5000, 1600),
		month('2026-05-01', 'met', 5400, 1890, 2),
		month('2026-04-01', 'met', 5000, 1500),
		month('2026-03-01', 'met', 5000, 1750),
		month('2026-02-01', 'missed', 4800, 960),
		month('2026-01-01', 'met', 4800, 1200)
	];
</script>

<Story name="Running month">
	<AnalyticsCard title="{monthLabel(running.month)} against your goal">
		<RunningMonthCard month={running} stacked />
	</AnalyticsCard>
</Story>

<Story name="Streak">
	<AnalyticsCard title="Streak">
		<StreakCard months={[running, ...finished]} currentStreak={4} bestStreak={4} />
	</AnalyticsCard>
</Story>

<Story name="Monthly standing">
	<AnalyticsCard title="Monthly standing">
		<MonthlyStandingCard months={finished} />
	</AnalyticsCard>
</Story>

<!-- The first run: nothing to score against, so the card sends you to the goal first. -->
<Story name="No goal yet">
	<AnalyticsCard title="Wealth goal">
		<NoGoalCard />
	</AnalyticsCard>
</Story>

<Story name="Loading">
	<AnalyticsCard title="July 2026 against your goal">
		<GoalLoadingCard shape="month" />
	</AnalyticsCard>
</Story>

<Story name="Loading, standing rows">
	<AnalyticsCard title="Monthly standing">
		<GoalLoadingCard shape="rows" />
	</AnalyticsCard>
</Story>

<Story name="Error">
	<AnalyticsCard title="Wealth goal">
		<GoalErrorCard />
	</AnalyticsCard>
</Story>

<!-- Before any month has finished there is nothing to list, and nothing is invented. -->
<Story name="No finished month yet">
	<AnalyticsCard title="Monthly standing">
		<MonthlyStandingCard months={[]} />
	</AnalyticsCard>
</Story>
