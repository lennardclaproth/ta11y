<script lang="ts">
	// One month against its goal: a bar with a goal marker, and the two numbers spelled out so
	// the result never depends on colour alone.
	import ProgressBar from '$lib/components/atoms/progress-bar/ProgressBar.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { sharePercent } from '$lib/api/wealthgoal';
	import { goalProgressIntents, goalProgressScale } from './goal-progress.variants';
	import type { ProgressBarSize } from '$lib/components/atoms/progress-bar/progress-bar.types';
	import type { MonthStanding } from '$lib/api/types';

	type Props = {
		month: MonthStanding;
		size?: ProgressBarSize;
		/** Hide the caption where the surrounding card already spells the numbers out. */
		caption?: boolean;
		class?: string;
	};

	let { month, size = 'lg', caption = true, class: className = '' }: Props = $props();

	const share = $derived(sharePercent(month));
	// The track runs to half of your income, so the goal marker sits inside it and a month
	// that overshot still reads as overshot.
	const markerLeft = $derived(
		`left: ${Math.min(100, (month.goal_percent / goalProgressScale) * 100)}%`
	);
</script>

<div class={['min-w-0', className].filter(Boolean).join(' ')}>
	<div class="relative">
		<ProgressBar
			value={Math.min(share, goalProgressScale)}
			max={goalProgressScale}
			{size}
			intent={goalProgressIntents[month.result]}
			ariaLabel="{share}% of marked income put towards wealth, goal {month.goal_percent}%"
		/>
		<span class="absolute inset-y-0 w-0.5 bg-slate-800" style={markerLeft} aria-hidden="true"
		></span>
	</div>
	{#if caption}
		<div class="mt-1 flex flex-wrap items-baseline justify-between gap-x-4">
			<Text size="xs" tone="muted">{share}% of income</Text>
			<Text size="xs" tone="muted">goal {month.goal_percent}%</Text>
		</div>
	{/if}
</div>
