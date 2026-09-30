<script lang="ts">
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import DonutChart from '$lib/components/organisms/charts/DonutChart.svelte';
	import { chartColors, donutRamps } from '$lib/charts/theme';
	import {
		cashflowLabels,
		cashflowNet,
		euro,
		incomingByTag,
		monthShort,
		outgoingByTag
	} from './mock-data';

	let { loading = false }: { loading?: boolean } = $props();
</script>

<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
	<AnalyticsCard title="Net trend" class="lg:col-span-2">
		<TimeSeriesChart
			height="h-44"
			labels={cashflowLabels}
			xTickFormat={monthShort}
			{loading}
			datasets={[{ label: 'Net', data: cashflowNet, color: chartColors.net, signed: true }]}
		/>
	</AnalyticsCard>
	<AnalyticsCard title="Incoming">
		<DonutChart
			data={incomingByTag}
			ramp={donutRamps.incoming}
			{loading}
			formatValue={euro}
			centerLabel="In"
		/>
	</AnalyticsCard>
	<AnalyticsCard title="Outgoing">
		<DonutChart
			data={outgoingByTag}
			ramp={donutRamps.outgoing}
			{loading}
			formatValue={euro}
			centerLabel="Out"
		/>
	</AnalyticsCard>
</div>
