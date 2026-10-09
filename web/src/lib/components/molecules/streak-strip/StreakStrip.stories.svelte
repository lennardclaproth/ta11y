<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import StreakStrip from './StreakStrip.svelte';
	import { numberToScaled as s } from '$lib/api/money';
	import type { MonthStanding } from '$lib/api/types';

	const { Story } = defineMeta({
		title: 'Molecules/StreakStrip',
		component: StreakStrip,
		tags: ['autodocs']
	});

	const month = (
		key: string,
		result: MonthStanding['result'],
		contributed: number
	): MonthStanding => ({
		month: key,
		income_cents: s(5000),
		contributed_cents: s(contributed),
		goal_percent: 30,
		result,
		unassigned_count: 0
	});

	// Newest first, as the standing returns them.
	const run: MonthStanding[] = [
		month('2026-07-01', 'in_progress', 900),
		month('2026-06-01', 'met', 1600),
		month('2026-05-01', 'met', 1890),
		month('2026-04-01', 'met', 1500),
		month('2026-03-01', 'met', 1750),
		month('2026-02-01', 'missed', 960),
		month('2026-01-01', 'met', 1200)
	];
</script>

<Story name="Playground" args={{ months: run }} />

<Story name="Unbroken run">
	<StreakStrip months={run.slice(0, 5)} />
</Story>

<Story name="Before the first goal">
	<StreakStrip
		months={[month('2026-02-01', 'not_scored', 0), month('2026-01-01', 'not_scored', 0)]}
	/>
</Story>
