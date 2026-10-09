<script lang="ts">
	// The same month, compressed to one ruled line for a narrow card: label, bar, share, result.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import GoalProgress from './GoalProgress.svelte';
	import { monthStatusLabel, sharePercent, type GoalMonth } from '../goal-data';

	type Props = { month: GoalMonth };

	let { month }: Props = $props();

	const badgeIntent = {
		met: 'success',
		missed: 'warning',
		'in-progress': 'neutral'
	} as const;
</script>

<div class="flex items-center gap-3 border-t border-slate-200 py-2.5 first:border-t-0">
	<span class="w-16 shrink-0 text-sm text-slate-700">{month.label.slice(0, 3)} {month.label.slice(-4)}</span>
	<div class="min-w-0 flex-1">
		<GoalProgress {month} size="sm" caption={false} />
	</div>
	<span class="w-10 shrink-0 text-right text-sm text-slate-700 tabular-nums">
		{sharePercent(month)}%
	</span>
	<span class="w-24 shrink-0 text-right">
		<Money amount={month.contributed} currency="EUR" size="sm" />
	</span>
	<Badge intent={badgeIntent[month.status]} variant="soft" size="sm" class="w-24 justify-center">
		{monthStatusLabel[month.status]}
	</Badge>
</div>
