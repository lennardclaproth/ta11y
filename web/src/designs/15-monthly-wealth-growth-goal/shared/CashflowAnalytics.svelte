<script lang="ts">
	// The analytics band Cashflow already has, unchanged, so the goal proposals can be judged
	// next to what is on the page today.
	import type { Snippet } from 'svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import DonutChart from '$lib/components/organisms/charts/DonutChart.svelte';
	import { cashflowMonthly, cashflowTagDistribution } from '$lib/data/fixtures/cashflow';
	import { scaledToNumber } from '$lib/api/money';
	import { chartColors, donutRamps } from '$lib/charts/theme';

	type Props = {
		/** An extra card appended to the band (the wealth-goal proposal). */
		extra?: Snippet;
	};

	let { extra }: Props = $props();

	const euro = (n: number) => `€${n.toLocaleString('en', { maximumFractionDigits: 0 })}`;
	const monthShort = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	// The band stays four columns wide: a fifth card would squeeze the donut legends. The goal
	// card takes the column the net trend gives up.
	const trendClass = $derived(extra ? '' : 'lg:col-span-2');
</script>

<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
	<AnalyticsCard title="Net trend" class={trendClass}>
		<TimeSeriesChart
			height="h-44"
			labels={cashflowMonthly.map((m) => m.month)}
			xTickFormat={monthShort}
			datasets={[
				{
					label: 'Net',
					data: cashflowMonthly.map((m) => scaledToNumber(m.net_cents)),
					color: chartColors.net,
					signed: true
				}
			]}
		/>
	</AnalyticsCard>
	<AnalyticsCard title="Incoming">
		<DonutChart
			data={cashflowTagDistribution.incoming.map((e) => ({
				label: e.tag,
				value: scaledToNumber(e.totalCents)
			}))}
			ramp={donutRamps.incoming}
			formatValue={euro}
			centerLabel="In"
		/>
	</AnalyticsCard>
	<AnalyticsCard title="Outgoing">
		<DonutChart
			data={cashflowTagDistribution.outgoing.map((e) => ({
				label: e.tag,
				value: scaledToNumber(e.totalCents)
			}))}
			ramp={donutRamps.outgoing}
			formatValue={euro}
			centerLabel="Out"
		/>
	</AnalyticsCard>
	{@render extra?.()}
</div>
