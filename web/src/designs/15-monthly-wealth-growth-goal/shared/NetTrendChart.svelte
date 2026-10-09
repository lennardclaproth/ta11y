<script lang="ts">
	// The net-trend chart Cashflow has today, on its own so a card can be placed per variant.
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import { cashflowMonthly } from '$lib/data/fixtures/cashflow';
	import { scaledToNumber } from '$lib/api/money';
	import { chartColors } from '$lib/charts/theme';

	type Props = { height?: string; loading?: boolean };

	let { height = 'h-44', loading = false }: Props = $props();

	const monthShort = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });
</script>

<TimeSeriesChart
	{height}
	{loading}
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
