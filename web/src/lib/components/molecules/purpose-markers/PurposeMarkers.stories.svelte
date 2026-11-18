<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import PurposeMarkers from './PurposeMarkers.svelte';
	import { numberToScaled as s } from '$lib/api/money';
	import type { MonthStanding } from '$lib/api/types';

	const { Story } = defineMeta({
		title: 'Molecules/PurposeMarkers',
		component: PurposeMarkers,
		tags: ['autodocs']
	});

	const month = (income: number, contributed: number): MonthStanding => ({
		month: '2026-06-01',
		income_cents: s(income),
		contributed_cents: s(contributed),
		goal_percent: 30,
		result: 'met',
		unassigned_count: 0
	});
</script>

<Story name="Playground" args={{ month: month(5000, 1600) }} />

<!-- Both marked: the month's score means something. -->
<Story name="Both marked" asChild>
	<PurposeMarkers month={month(5000, 1600)} />
</Story>

<!-- Income marked, nothing to wealth yet: the missing mark links to the month's outgoing money. -->
<Story name="Wealth missing" asChild>
	<PurposeMarkers month={month(5000, 0)} />
</Story>

<Story name="Neither marked" asChild>
	<PurposeMarkers month={month(0, 0)} />
</Story>
