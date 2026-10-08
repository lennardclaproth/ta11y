<script lang="ts">
	// Proposal for a new molecule: the run of met months, oldest to newest, as the one piece of
	// encouragement the pitch allows. No points, badges or levels.
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { monthStatusLabel, type GoalMonth } from '../goal-data';

	type Props = { months: GoalMonth[] };

	let { months }: Props = $props();

	const marks = $derived([...months].reverse());

	const markClasses = {
		met: 'border-emerald-700 bg-emerald-100 text-emerald-800',
		missed: 'border-amber-500 bg-amber-100 text-amber-800',
		'in-progress': 'border-slate-300 bg-slate-100 text-slate-500'
	} as const;

	const markIcons = {
		met: 'heroicons:check',
		missed: 'heroicons:x-mark',
		'in-progress': 'heroicons:ellipsis-horizontal'
	} as const;
</script>

<ol class="flex flex-wrap items-start gap-3">
	{#each marks as mark (mark.month)}
		<li class="flex w-10 flex-col items-center gap-1">
			<span
				class={[
					'flex size-9 items-center justify-center rounded-full border',
					markClasses[mark.status]
				].join(' ')}
			>
				<Icon icon={markIcons[mark.status]} size="lg" />
			</span>
			<span class="text-xs text-slate-500">{mark.label.slice(0, 3)}</span>
			<span class="sr-only">{mark.label}: {monthStatusLabel[mark.status]}</span>
		</li>
	{/each}
</ol>
