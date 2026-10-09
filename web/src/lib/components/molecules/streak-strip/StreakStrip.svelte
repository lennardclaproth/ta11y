<script lang="ts">
	// The run of met months, oldest to newest — the one piece of encouragement the pitch
	// allows. No points, badges or levels.
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { monthLabel, monthResultLabel } from '$lib/api/wealthgoal';
	import { streakMarkClasses, streakMarkIcons } from './streak-strip.variants';
	import type { MonthStanding } from '$lib/api/types';

	type Props = {
		/** Newest first, as the standing returns them. */
		months: MonthStanding[];
		class?: string;
	};

	let { months, class: className = '' }: Props = $props();

	const marks = $derived([...months].reverse());
</script>

<!-- Marks stay small enough that a year of months still fits one line in a quarter-width card. -->
<ol class={['flex flex-wrap items-start gap-2', className].filter(Boolean).join(' ')}>
	{#each marks as mark (mark.month)}
		<li class="flex w-9 flex-col items-center gap-1">
			<span
				class={[
					'flex size-8 items-center justify-center rounded-full border',
					streakMarkClasses[mark.result]
				].join(' ')}
			>
				<Icon icon={streakMarkIcons[mark.result]} size="lg" />
			</span>
			<span class="text-xs text-slate-500">{monthLabel(mark.month).slice(0, 3)}</span>
			<span class="sr-only">{monthLabel(mark.month)}: {monthResultLabel[mark.result]}</span>
		</li>
	{/each}
</ol>
