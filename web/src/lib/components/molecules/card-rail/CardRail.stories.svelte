<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import CardRail from './CardRail.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import { cardRailWidths } from './card-rail.types';

	const { Story } = defineMeta({
		title: 'Molecules/CardRail',
		component: CardRail,
		tags: ['autodocs']
	});

	const groups = [
		{ id: 'overview', label: 'Overview' },
		{ id: 'goal', label: 'Wealth goal' }
	];
</script>

<!-- Children carry `data-card` and `data-group`; the rail reads those rather than owning
     the card list. -->
<Story name="Playground">
	<CardRail
		{groups}
		ariaLabel="Overview and wealth-goal cards"
		hint="Scroll sideways for the rest."
	>
		<div data-card="trend" data-group="overview" class={cardRailWidths.wide}>
			<AnalyticsCard title="Net trend">
				<div class="h-32 rounded bg-taupe-200/60"></div>
			</AnalyticsCard>
		</div>
		<div data-card="incoming" data-group="overview" class={cardRailWidths.default}>
			<AnalyticsCard title="Incoming">
				<div class="h-32 rounded bg-taupe-200/60"></div>
			</AnalyticsCard>
		</div>
		<div data-card="outgoing" data-group="overview" class={cardRailWidths.default}>
			<AnalyticsCard title="Outgoing">
				<div class="h-32 rounded bg-taupe-200/60"></div>
			</AnalyticsCard>
		</div>
		<div data-card="month" data-group="goal" class={cardRailWidths.wide}>
			<AnalyticsCard title="July 2026 against your goal">
				<div class="h-32 rounded bg-taupe-200/60"></div>
			</AnalyticsCard>
		</div>
		<div data-card="streak" data-group="goal" class={cardRailWidths.default}>
			<AnalyticsCard title="Streak">
				<div class="h-32 rounded bg-taupe-200/60"></div>
			</AnalyticsCard>
		</div>
	</CardRail>
</Story>

<!-- A group of one card fills the band itself, so jumping to it reaches the left edge. -->
<Story name="Single-card group">
	<CardRail {groups} ariaLabel="Overview and wealth-goal cards">
		<div data-card="trend" data-group="overview" class={cardRailWidths.wide}>
			<AnalyticsCard title="Net trend">
				<div class="h-32 rounded bg-taupe-200/60"></div>
			</AnalyticsCard>
		</div>
		<div data-card="month" data-group="goal" class={cardRailWidths.solo}>
			<AnalyticsCard title="Wealth goal">
				<div class="h-32 rounded bg-taupe-200/60"></div>
			</AnalyticsCard>
		</div>
	</CardRail>
</Story>
