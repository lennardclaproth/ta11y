<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import GoalProgress from './GoalProgress.svelte';
	import { progressBarSizes } from '$lib/components/atoms/progress-bar/progress-bar.types';
	import { numberToScaled as s } from '$lib/api/money';
	import type { MonthStanding } from '$lib/api/types';

	const { Story } = defineMeta({
		title: 'Molecules/GoalProgress',
		component: GoalProgress,
		tags: ['autodocs'],
		argTypes: {
			size: { control: 'select', options: progressBarSizes },
			caption: { control: 'boolean' }
		}
	});

	const month = (contributed: number, result: MonthStanding['result'] = 'met'): MonthStanding => ({
		month: '2026-06-01',
		income_cents: s(5000),
		contributed_cents: s(contributed),
		goal_percent: 30,
		result,
		unassigned_count: 0
	});
</script>

<Story name="Playground" args={{ month: month(1600), size: 'lg', caption: true }} />

<Story name="Met">
	<GoalProgress month={month(1600)} />
</Story>

<Story name="Missed">
	<GoalProgress month={month(600, 'missed')} />
</Story>

<Story name="Running month">
	<GoalProgress month={month(900, 'in_progress')} />
</Story>

<!-- The track runs to half of income, so a month that overshot still reads as overshot. -->
<Story name="Overshot">
	<GoalProgress month={month(2400)} />
</Story>

<Story name="Compact, no caption">
	<GoalProgress month={month(1600)} size="sm" caption={false} />
</Story>
