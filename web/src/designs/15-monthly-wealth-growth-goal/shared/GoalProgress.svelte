<script lang="ts">
	// One month against its goal: a bar with a goal marker and the two numbers spelled out, so the
	// result never depends on colour alone.
	import ProgressBar from '$lib/components/atoms/progress-bar/ProgressBar.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { sharePercent, type GoalMonth } from '../goal-data';
	import type { ProgressBarSize } from '$lib/components/atoms/progress-bar/progress-bar.types';

	type Props = {
		month: GoalMonth;
		size?: ProgressBarSize;
		/** Hide the caption where the surrounding card already spells the numbers out. */
		caption?: boolean;
	};

	let { month, size = 'lg', caption = true }: Props = $props();

	// The track runs to half of your income, so the goal marker sits inside it and a month that
	// overshot still reads as overshot.
	const scale = 50;

	const share = $derived(sharePercent(month));
	const markerLeft = $derived(`left: ${Math.min(100, (month.goalPercent / scale) * 100)}%`);

	// Missed needs attention; it is not a validation error, so amber rather than red.
	const barIntent = {
		met: 'success',
		missed: 'warning',
		'in-progress': 'primary'
	} as const;
</script>

<div class="min-w-0">
	<div class="relative">
		<ProgressBar
			value={Math.min(share, scale)}
			max={scale}
			{size}
			intent={barIntent[month.status]}
			ariaLabel="{share}% of marked income put towards wealth, goal {month.goalPercent}%"
		/>
		<span class="absolute inset-y-0 w-0.5 bg-slate-800" style={markerLeft} aria-hidden="true"></span>
	</div>
	{#if caption}
		<div class="mt-1 flex flex-wrap items-baseline justify-between gap-x-4">
			<Text size="xs" tone="muted">{share}% of income</Text>
			<Text size="xs" tone="muted">goal {month.goalPercent}%</Text>
		</div>
	{/if}
</div>
