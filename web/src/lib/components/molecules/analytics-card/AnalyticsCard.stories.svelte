<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import AnalyticsCard from './AnalyticsCard.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/AnalyticsCard',
		component: AnalyticsCard,
		tags: ['autodocs'],
		argTypes: {
			title: { control: 'text' }
		}
	});
</script>

<Story name="Playground" asChild>
	<div class="max-w-sm">
		<AnalyticsCard title="Net trend">
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">
				Chart goes here
			</div>
		</AnalyticsCard>
	</div>
</Story>

<Story name="Without title" asChild>
	<div class="max-w-sm">
		<AnalyticsCard>
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">Untitled card</div>
		</AnalyticsCard>
	</div>
</Story>

<!-- Actions that recalculate the chart hang on its rule, not in the navigation bar. -->
<Story name="With a chart action" asChild>
	<div class="max-w-xl">
		<AnalyticsCard title="Value vs cost basis">
			{#snippet actions()}
				<span class="text-xs text-slate-500">1 Jan 2026 – 30 Jun 2026</span>
				<Button variant="ruled">
					<Icon icon="heroicons:arrow-path" />
					Rebuild portfolio
				</Button>
			{/snippet}
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">Chart</div>
		</AnalyticsCard>
	</div>
</Story>

<Story name="In a grid" asChild>
	<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
		<AnalyticsCard title="Net trend" class="lg:col-span-2">
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">Wide chart</div>
		</AnalyticsCard>
		<AnalyticsCard title="Incoming">
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">Donut</div>
		</AnalyticsCard>
		<AnalyticsCard title="Outgoing">
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">Donut</div>
		</AnalyticsCard>
	</div>
</Story>
