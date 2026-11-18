<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RankedBarChart from './RankedBarChart.svelte';
	import { donutRamps } from '$lib/charts/theme';

	const { Story } = defineMeta({
		title: 'Organisms/Charts/RankedBarChart',
		component: RankedBarChart,
		tags: ['autodocs']
	});

	const euro = (value: number) =>
		new Intl.NumberFormat('nl-NL', { style: 'currency', currency: 'EUR' }).format(value);
	const outgoing = [
		{ key: 'rent', label: 'rent', value: 5800 },
		{ key: 'groceries', label: 'groceries', value: 389.5 },
		{ key: 'entertainment', label: 'entertainment', value: 311 },
		{ key: 'utilities', label: 'utilities', value: 291.42 },
		{ key: 'dining', label: 'dining', value: 105 },
		{ key: 'untagged', label: 'Untagged', value: 89.49 }
	];
</script>

<script lang="ts">
	let selected = $state<string | null>(null);
</script>

<Story name="Playground" args={{ data: outgoing, ramp: donutRamps.outgoing, formatValue: euro }} />

<!-- Rows are toggle buttons when the chart drives a filter. -->
<Story name="Selectable" asChild>
	<div class="w-[32rem] max-w-full">
		<RankedBarChart
			data={outgoing}
			ramp={donutRamps.outgoing}
			formatValue={euro}
			{selected}
			onSelect={(key) => (selected = key)}
		/>
		<p class="mt-2 text-xs text-slate-500">Selected: {selected ?? '—'}</p>
	</div>
</Story>

<Story name="Extra cells" asChild>
	<div class="w-[40rem] max-w-full">
		<RankedBarChart data={outgoing} formatValue={euro}>
			{#snippet extra(row)}
				<span class="w-14 shrink-0 text-right text-xs text-emerald-700 tabular-nums">
					+{(row.value / 100).toFixed(1)}%
				</span>
			{/snippet}
		</RankedBarChart>
	</div>
</Story>

<Story name="Loading" args={{ data: [], loading: true }} />

<Story name="Empty" args={{ data: [] }} />
