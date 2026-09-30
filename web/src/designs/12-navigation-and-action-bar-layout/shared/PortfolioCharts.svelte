<script lang="ts">
	import type { Snippet } from 'svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import AnalyticsSection from './AnalyticsSection.svelte';
	import { chartColors } from '$lib/charts/theme';
	import {
		monthShort,
		portfolioCostBasis,
		portfolioKpis,
		portfolioLabels,
		portfolioMarketValue
	} from './mock-data';

	/** `actions` holds the chart-level actions, e.g. "Rebuild portfolio". */
	let { actions }: { actions?: Snippet } = $props();
</script>

<div class="flex flex-col gap-3">
	<KpiRow items={portfolioKpis} columns={3} />
	<AnalyticsSection title="Value vs cost basis" {actions}>
		<TimeSeriesChart
			height="h-52"
			labels={portfolioLabels}
			xTickFormat={monthShort}
			datasets={[
				{
					label: 'Market value',
					data: portfolioMarketValue,
					color: chartColors.positive,
					fill: true
				},
				{
					label: 'Cost basis',
					data: portfolioCostBasis,
					color: chartColors.net,
					dashed: true
				}
			]}
		/>
	</AnalyticsSection>
</div>
