<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import MonthStandingRow from './MonthStandingRow.svelte';
	import { numberToScaled as s } from '$lib/api/money';
	import type { MonthStanding } from '$lib/api/types';

	const { Story } = defineMeta({
		title: 'Molecules/MonthStandingRow',
		component: MonthStandingRow,
		tags: ['autodocs'],
		argTypes: { divider: { control: 'boolean' } }
	});

	const month = (
		key: string,
		result: MonthStanding['result'],
		contributed: number,
		unassigned = 0
	): MonthStanding => ({
		month: key,
		income_cents: s(5000),
		contributed_cents: s(contributed),
		goal_percent: 30,
		result,
		unassigned_count: unassigned
	});

	const run = [
		month('2026-06-01', 'met', 1600),
		month('2026-05-01', 'met', 1890, 2),
		month('2026-04-01', 'missed', 900)
	];
</script>

<Story name="Playground" args={{ month: month('2026-06-01', 'met', 1600), divider: false }} />

<!-- The list the monthly standing card renders: one ruled line per finished month. -->
<Story name="Standing list">
	{#each run as entry, index (entry.month)}
		<MonthStandingRow month={entry} divider={index > 0} />
	{/each}
</Story>

<Story name="Incomplete month">
	<MonthStandingRow month={month('2026-05-01', 'met', 1890, 2)} divider={false} />
</Story>

<Story name="Missed">
	<MonthStandingRow month={month('2026-02-01', 'missed', 960)} divider={false} />
</Story>
