<script lang="ts">
	// Every month's share against the goal that applied in it, as one chart: a bar per month
	// coloured by its result, the goal as a dashed line. The run of months met — what the
	// streak card used to say — is the line under it.
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import { chartHeights } from '$lib/components/organisms/charts/charts.variants';
	import { chartColors } from '$lib/charts/theme';
	import { monthLabelShort, monthResultLabel, sharePercent } from '$lib/api/wealthgoal';
	import type { MonthResult, MonthStanding } from '$lib/api/types';

	type Props = {
		/** Newest first, as the standing returns them. */
		months: MonthStanding[];
		currentStreak: number;
		bestStreak: number;
		height?: string;
	};

	let {
		months,
		currentStreak,
		bestStreak,
		height = chartHeights.tallBetweenLegendAndCaption
	}: Props = $props();

	const resultColors = {
		met: chartColors.positive,
		missed: chartColors.warning,
		in_progress: chartColors.info,
		not_scored: chartColors.muted
	} satisfies Record<MonthResult, string>;
	const legend: MonthResult[] = ['met', 'missed', 'in_progress'];

	const ordered = $derived([...months].sort((a, b) => a.month.localeCompare(b.month)));
	const monthsText = (n: number) => (n === 1 ? '1 month' : `${n} months`);
</script>

<div class="flex flex-col gap-3">
	<ul class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-slate-600" aria-label="Legend">
		{#each legend as result (result)}
			<li class="inline-flex items-center gap-1.5">
				<span class="size-2.5 rounded-sm" style:background-color={resultColors[result]}></span>
				{monthResultLabel[result]}
			</li>
		{/each}
		<li class="inline-flex items-center gap-1.5">
			<span class="w-3 border-t-2 border-dashed border-slate-400"></span>Goal
		</li>
	</ul>
	<TimeSeriesChart
		{height}
		labels={ordered.map((m) => m.month)}
		xTickFormat={(month) => monthLabelShort(month).split(' ')[0]}
		ariaLabel="Share of income that went to wealth per month, against the goal"
		datasets={[
			{
				label: 'Goal %',
				data: ordered.map((m) => m.goal_percent),
				color: chartColors.zeroLine,
				dashed: true
			},
			{
				label: 'To wealth %',
				data: ordered.map(sharePercent),
				color: chartColors.info,
				barColors: ordered.map((m) => resultColors[m.result]),
				type: 'bar'
			}
		]}
	/>
	<Text size="sm" tone="muted">
		{currentStreak > 0
			? `Goal met ${monthsText(currentStreak)} in a row`
			: 'No run of met months right now'}; best run {monthsText(bestStreak)}.
	</Text>
</div>
