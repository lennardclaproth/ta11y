<script lang="ts">
	// The only encouragement the pitch allows: the run of months you met, nothing else.
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import StreakStrip from '$lib/components/molecules/streak-strip/StreakStrip.svelte';
	import { monthLabel } from '$lib/api/wealthgoal';
	import { goalFigureClass } from './wealth-goal-cards.variants';
	import type { MonthStanding } from '$lib/api/types';

	type Props = {
		/** Newest first, as the standing returns them. */
		months: MonthStanding[];
		currentStreak: number;
		bestStreak: number;
	};

	let { months, currentStreak, bestStreak }: Props = $props();

	const running = $derived(months.find((month) => month.result === 'in_progress'));
	const figure = $derived(
		currentStreak === 1 ? '1 month in a row' : `${currentStreak} months in a row`
	);
</script>

<div class="flex flex-col gap-3">
	<span class={goalFigureClass}>{figure}</span>
	<StreakStrip {months} />
	<Text size="sm" tone="muted">
		Best run so far: {bestStreak}
		{bestStreak === 1 ? 'month' : 'months'}.
		{#if running}
			{monthLabel(running.month).split(' ')[0]} is still running, so it does not count yet.
		{/if}
	</Text>
</div>
