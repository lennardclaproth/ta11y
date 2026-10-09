<script lang="ts">
	// One completed month: label, bar against the goal, share, what went to wealth, the result and
	// whether the month is still incomplete. One ruled line from `sm` up; stacked below that,
	// because six aligned columns do not survive a 328px card without dropping a label.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import GoalProgress from './GoalProgress.svelte';
	import { monthStatusLabel, sharePercent, type GoalMonth } from '../goal-data';

	type Props = {
		month: GoalMonth;
		/** The rule above the row; the first row in a list does not get one. */
		divider?: boolean;
		onOpenUnassigned?: (month: GoalMonth) => void;
	};

	let { month, divider = true, onOpenUnassigned }: Props = $props();

	const badgeIntent = {
		met: 'success',
		missed: 'warning',
		'in-progress': 'neutral'
	} as const;

	const short = $derived(`${month.label.slice(0, 3)} ${month.label.slice(-4)}`);
	const share = $derived(sharePercent(month));
</script>

{#snippet openLink()}
	{#if month.unassigned > 0}
		<button
			type="button"
			class="text-xs text-amber-700 underline underline-offset-2 hover:text-amber-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
			onclick={() => onOpenUnassigned?.(month)}
		>
			{month.unassigned} open
			<span class="sr-only">transactions in {short} are not assigned yet</span>
		</button>
	{/if}
{/snippet}

<div class={['py-2', divider ? 'border-t border-slate-200' : ''].filter(Boolean).join(' ')}>
	<!-- From `sm`: one line, comparable amounts in aligned right-hand columns. -->
	<div class="hidden items-center gap-3 sm:flex">
		<span class="w-[4.5rem] shrink-0 text-sm text-slate-700">{short}</span>
		<div class="min-w-0 flex-1">
			<GoalProgress {month} size="sm" caption={false} />
		</div>
		<span class="w-10 shrink-0 text-right text-sm text-slate-700 tabular-nums">{share}%</span>
		<span class="w-24 shrink-0 text-right">
			<Money amount={month.contributed} currency="EUR" size="sm" />
		</span>
		<Badge
			intent={badgeIntent[month.status]}
			variant="soft"
			size="sm"
			class="w-[4.5rem] justify-center"
		>
			{monthStatusLabel[month.status]}
		</Badge>
		<!-- Incomplete is a separate fact from the result: the month counts, but it may still move. -->
		<span class="w-14 shrink-0 text-right">{@render openLink()}</span>
	</div>

	<!-- Below `sm`: two lines, nothing dropped — six aligned columns do not survive a 328px card. -->
	<div class="flex flex-col gap-1 sm:hidden">
		<div class="flex items-center justify-between gap-2">
			<span class="text-sm text-slate-700">{short}</span>
			<Badge intent={badgeIntent[month.status]} variant="soft" size="sm">
				{monthStatusLabel[month.status]}
			</Badge>
		</div>
		<div class="flex items-center gap-2">
			<div class="min-w-0 flex-1">
				<GoalProgress {month} size="sm" caption={false} />
			</div>
			<span class="shrink-0 text-xs text-slate-500 tabular-nums">{share}%</span>
			<span class="shrink-0"><Money amount={month.contributed} currency="EUR" size="sm" /></span>
			{@render openLink()}
		</div>
	</div>
</div>
