<script lang="ts">
	// One completed month: label, bar against the goal, share, what went to wealth, the result
	// and whether income and a contribution were marked at all. One ruled line from `sm` up; stacked below
	// that, because six aligned columns do not survive a 328px card without dropping a label.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import GoalProgress from '$lib/components/molecules/goal-progress/GoalProgress.svelte';
	import { goalResultBadgeIntents } from '$lib/components/molecules/goal-progress/goal-progress.variants';
	import { monthLabelShort, monthResultLabel, sharePercent } from '$lib/api/wealthgoal';
	import { scaledToNumber } from '$lib/api/money';
	import PurposeMarkers from '$lib/components/molecules/purpose-markers/PurposeMarkers.svelte';
	import type { MonthStanding } from '$lib/api/types';

	type Props = {
		month: MonthStanding;
		/** The track shared by every row of the list, so bars stay comparable when goals differ. */
		track?: number;
		/** The rule above the row; the first row in a list does not get one. */
		divider?: boolean;
		/** Jump to this month's unmarked money that could carry the missing purpose. */
		onOpenMissing?: (month: MonthStanding, purpose: 'income' | 'wealth') => void;
		class?: string;
	};

	let { month, track, divider = true, onOpenMissing, class: className = '' }: Props = $props();

	const label = $derived(monthLabelShort(month.month));
	const share = $derived(sharePercent(month));
	const contributed = $derived(scaledToNumber(month.contributed_cents));
</script>

<div
	class={['py-2', divider ? 'border-t border-slate-200' : '', className].filter(Boolean).join(' ')}
>
	<!-- From `sm`: one line, comparable amounts in aligned right-hand columns, and the marks the
	     score rests on underneath, under the bar — beside it they would squeeze the bar away. -->
	<div class="hidden items-center gap-3 sm:flex">
		<span class="w-[4.5rem] shrink-0 text-sm text-slate-700">{label}</span>
		<div class="min-w-0 flex-1">
			<GoalProgress {month} {track} size="sm" caption={false} />
		</div>
		<span class="w-10 shrink-0 text-right text-sm text-slate-700 tabular-nums">{share}%</span>
		<span class="w-24 shrink-0 text-right">
			<Money amount={contributed} currency="EUR" size="sm" />
		</span>
		<Badge
			intent={goalResultBadgeIntents[month.result]}
			variant="soft"
			size="sm"
			class="w-[4.5rem] justify-center"
		>
			{monthResultLabel[month.result]}
		</Badge>
	</div>
	<div class="mt-1 hidden pl-[5.25rem] sm:block">
		<PurposeMarkers {month} {onOpenMissing} />
	</div>

	<!-- Below `sm`: three lines, nothing dropped. -->
	<div class="flex flex-col gap-1 sm:hidden">
		<div class="flex items-center justify-between gap-2">
			<span class="text-sm text-slate-700">{label}</span>
			<Badge intent={goalResultBadgeIntents[month.result]} variant="soft" size="sm">
				{monthResultLabel[month.result]}
			</Badge>
		</div>
		<div class="flex items-center gap-2">
			<div class="min-w-0 flex-1">
				<GoalProgress {month} {track} size="sm" caption={false} />
			</div>
			<span class="shrink-0 text-xs text-slate-500 tabular-nums">{share}%</span>
			<span class="shrink-0"><Money amount={contributed} currency="EUR" size="sm" /></span>
		</div>
		<PurposeMarkers {month} {onOpenMissing} />
	</div>
</div>
