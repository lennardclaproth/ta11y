<script lang="ts">
	// The tag-distribution donut Cashflow has today, one direction per card.
	import DonutChart from '$lib/components/organisms/charts/DonutChart.svelte';
	import { cashflowTagDistribution } from '$lib/data/fixtures/cashflow';
	import { scaledToNumber } from '$lib/api/money';
	import { donutRamps } from '$lib/charts/theme';

	type Props = { direction: 'in' | 'out' };

	let { direction }: Props = $props();

	const euro = (n: number) => `€${n.toLocaleString('en', { maximumFractionDigits: 0 })}`;

	const entries = $derived(
		direction === 'in' ? cashflowTagDistribution.incoming : cashflowTagDistribution.outgoing
	);
</script>

<DonutChart
	data={entries.map((e) => ({ label: e.tag, value: scaledToNumber(e.totalCents) }))}
	ramp={direction === 'in' ? donutRamps.incoming : donutRamps.outgoing}
	formatValue={euro}
	centerLabel={direction === 'in' ? 'In' : 'Out'}
/>
