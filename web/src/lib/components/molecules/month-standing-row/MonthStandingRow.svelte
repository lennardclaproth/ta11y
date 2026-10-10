<script lang="ts">
	// One completed month: label, bar against the goal, share, what went to wealth, the result
	// and whether the month is still incomplete. One ruled line from `sm` up; stacked below
	// that, because six aligned columns do not survive a 328px card without dropping a label.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import GoalProgress from '$lib/components/molecules/goal-progress/GoalProgress.svelte';
	import { goalResultBadgeIntents } from '$lib/components/molecules/goal-progress/goal-progress.variants';
	import { monthLabelShort, monthResultLabel, sharePercent } from '$lib/api/wealthgoal';
	import { scaledToNumber } from '$lib/api/money';
	import type { MonthStanding } from '$lib/api/types';

	type Props = {
		month: MonthStanding;
		/** The track shared by every row of the list, so bars stay comparable when goals differ. */
		track?: number;
		/** The rule above the row; the first row in a list does not get one. */
		divider?: boolean;
		/** Jump to the transactions of this month that carry no purpose yet. */
		onOpenUnassigned?: (month: MonthStanding) => void;
		class?: string;
	};

	let { month, track, divider = true, onOpenUnassigned, class: className = '' }: Props = $props();

	const label = $derived(monthLabelShort(month.month));
	const share = $derived(sharePercent(month));
	const contributed = $derived(scaledToNumber(month.contributed_cents));
</script>

{#snippet openLink()}
	{#if month.unassigned_count > 0}
		<button
			type="button"
			class="text-xs text-amber-700 underline underline-offset-2 hover:text-amber-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
			onclick={() => onOpenUnassigned?.(month)}
		>
			{month.unassigned_count} open
			<span class="sr-only">transactions in {label} are not assigned yet</span>
		</button>
	{/if}
{/snippet}

<div
	class={['py-2', divider ? 'border-t border-slate-200' : '', className].filter(Boolean).join(' ')}
>
	<!-- From `sm`: one line, comparable amounts in aligned right-hand columns. -->
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
		<!-- Incomplete is a separate fact from the result: the month counts, but it may still move. -->
		<span class="w-14 shrink-0 text-right">{@render openLink()}</span>
	</div>

	<!-- Below `sm`: two lines, nothing dropped. -->
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
			{@render openLink()}
		</div>
	</div>
</div>
