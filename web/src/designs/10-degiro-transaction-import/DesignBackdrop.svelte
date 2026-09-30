<script lang="ts">
	// Prototype decor only: a stand-in for the Portfolio page so the overlay variants
	// are judged in context instead of floating on white. Not a proposed component.
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';

	let { label = 'Portfolio' }: { label?: string } = $props();
</script>

<div class="min-h-[40rem] bg-taupe-100 px-4 py-4 lg:px-8">
	<header class="mb-4 border-b-2 border-slate-900 pb-2">
		<Heading level="h1" size="xl" class="text-slate-900">{label}</Heading>
	</header>

	<KpiRow
		items={[
			{ label: 'Market value', amount: 24000, currency: 'EUR', change: 6.5 },
			{ label: 'Total P&L', amount: 4000, currency: 'EUR', change: 20 },
			{ label: 'Cost basis', amount: 20000, currency: 'EUR' }
		]}
		columns={3}
		class="mb-4"
	/>

	<div class="border-t border-slate-400 bg-taupe-50 px-3 py-6">
		<Text as="p" size="sm" tone="muted">Positions</Text>
		<div class="mt-3 space-y-2">
			{#each [0, 1, 2, 3, 4] as row (row)}
				<div class="h-6 border-b border-slate-200"></div>
			{/each}
		</div>
	</div>
</div>
