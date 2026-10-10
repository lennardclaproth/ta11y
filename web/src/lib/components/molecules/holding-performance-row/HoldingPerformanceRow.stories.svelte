<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';

	import HoldingPerformanceRow from './HoldingPerformanceRow.svelte';

	import { holdingRowModes } from './holding-performance-row.types';
	import { assetHoldings } from '$lib/data/fixtures/assets';

	const bitcoin = assetHoldings['ast-btc'];
	const ethereum = assetHoldings['ast-eth'];

	/** An item linked today: no purchase has a price behind it yet. */
	const unpriced = {
		...bitcoin,
		id: 'ast-sol',
		name: 'Solana',
		symbol: 'SOL/EUR',
		quantity: '12',
		paid: '1200.000000',
		avg_unit_price: '100.000000',
		price: '0.000000',
		price_date: null,
		price_carried_forward: false,
		value: '1200.000000',
		unrealized: '0.000000',
		unrealized_pct: 0,
		series: []
	};

	const { Story } = defineMeta({
		title: 'Molecules/HoldingPerformanceRow',
		component: HoldingPerformanceRow,
		tags: ['autodocs'],
		argTypes: {
			mode: { control: 'select', options: holdingRowModes },
			open: { control: 'boolean' },
			current: { control: 'boolean' },
			showKind: { control: 'boolean' }
		}
	});
</script>

<Story name="Playground" args={{ holding: bitcoin, mode: 'expand', open: true }} />

<Story name="Collapsed" asChild>
	<ul class="max-w-lg rounded-xl border border-slate-200">
		<HoldingPerformanceRow holding={bitcoin} />
		<HoldingPerformanceRow holding={ethereum} />
	</ul>
</Story>

<Story name="Expanded" asChild>
	<ul class="max-w-lg rounded-xl border border-slate-200">
		<HoldingPerformanceRow holding={bitcoin} open onOpenItem={() => {}} />
	</ul>
</Story>

<Story name="Carried forward price" asChild>
	<ul class="max-w-lg rounded-xl border border-slate-200">
		<HoldingPerformanceRow holding={ethereum} open />
	</ul>
</Story>

<Story name="Select mode" asChild>
	<ul class="max-w-lg rounded-xl border border-slate-200">
		<HoldingPerformanceRow holding={bitcoin} mode="select" current showKind={false} />
		<HoldingPerformanceRow holding={ethereum} mode="select" showKind={false} />
	</ul>
</Story>

<Story name="No price yet" asChild>
	<ul class="max-w-lg rounded-xl border border-slate-200">
		<HoldingPerformanceRow holding={unpriced} open />
	</ul>
</Story>
