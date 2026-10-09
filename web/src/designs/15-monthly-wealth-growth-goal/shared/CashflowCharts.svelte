<script lang="ts">
	// The analytics band Cashflow has today, unchanged, so the goal boards can be judged next to
	// what is already on the page.
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import NetTrendChart from './NetTrendChart.svelte';
	import TagDonut from './TagDonut.svelte';
</script>

<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
	<AnalyticsCard title="Net trend" class="lg:col-span-2">
		<NetTrendChart />
	</AnalyticsCard>
	<AnalyticsCard title="Incoming">
		<TagDonut direction="in" />
	</AnalyticsCard>
	<AnalyticsCard title="Outgoing">
		<TagDonut direction="out" />
	</AnalyticsCard>
</div>
