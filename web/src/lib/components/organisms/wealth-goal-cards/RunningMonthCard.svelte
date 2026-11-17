<script lang="ts">
	// The month you are in: what you marked, what the goal asks for, what is still open, and
	// the goal itself with the way back to it — so the running month is one card, not three.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import GoalProgress from '$lib/components/molecules/goal-progress/GoalProgress.svelte';
	import { goalResultBadgeIntents } from '$lib/components/molecules/goal-progress/goal-progress.variants';
	import { scaledToNumber } from '$lib/api/money';
	import {
		goalAmount,
		monthLabel,
		monthResultLabel,
		remainingAmount,
		sharePercent
	} from '$lib/api/wealthgoal';
	import { goalFigureClass } from './wealth-goal-cards.variants';
	import type { MonthStanding } from '$lib/api/types';

	type Props = {
		month: MonthStanding;
		/** Stack the amounts under the bar instead of beside it (narrow cards). */
		stacked?: boolean;
		onOpenUnassigned?: (month: MonthStanding) => void;
		onAdjust?: () => void;
	};

	let { month, stacked = false, onOpenUnassigned, onAdjust }: Props = $props();

	const label = $derived(monthLabel(month.month));
	const share = $derived(sharePercent(month));
	const target = $derived(goalAmount(month));
	const remaining = $derived(remainingAmount(month));
	const income = $derived(scaledToNumber(month.income_cents));
	const contributed = $derived(scaledToNumber(month.contributed_cents));
</script>

<div class="flex flex-col gap-3">
	<div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
		<span class={goalFigureClass}>
			{share}% <span class="font-sans text-base text-slate-500">of income so far</span>
		</span>
		<Badge intent={goalResultBadgeIntents[month.result]} variant="soft" size="md">
			{monthResultLabel[month.result]}
		</Badge>
	</div>

	<div class={stacked ? 'flex flex-col gap-3' : 'flex flex-col gap-4 lg:flex-row lg:items-end'}>
		<div class="min-w-0 flex-1">
			<GoalProgress {month} />
		</div>

		<!-- Right-aligned, equal-width columns so the three amounts read as one set of numbers.
		     The stacked card divides its own width instead of three fixed columns, which
		     overflow a 328px card. -->
		<dl class={stacked ? 'grid grid-cols-3 gap-3' : 'flex shrink-0 justify-end gap-6'}>
			<div class={['flex flex-col items-end gap-0.5', stacked ? 'min-w-0' : 'w-28'].join(' ')}>
				<dt class="text-xs text-slate-500">Income</dt>
				<dd><Money amount={income} currency="EUR" size="md" /></dd>
			</div>
			<div class={['flex flex-col items-end gap-0.5', stacked ? 'min-w-0' : 'w-28'].join(' ')}>
				<dt class="text-xs text-slate-500">To wealth</dt>
				<dd><Money amount={contributed} currency="EUR" size="md" weight="semibold" /></dd>
			</div>
			<div class={['flex flex-col items-end gap-0.5', stacked ? 'min-w-0' : 'w-28'].join(' ')}>
				<dt class="text-xs text-slate-500">Still to go</dt>
				<dd><Money amount={remaining} currency="EUR" size="md" /></dd>
			</div>
		</dl>
	</div>

	{#if month.unassigned_count > 0}
		<button
			type="button"
			class="w-fit text-sm text-sky-700 underline underline-offset-2 hover:text-sky-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
			onclick={() => onOpenUnassigned?.(month)}
		>
			{month.unassigned_count} transactions in {label} are not assigned yet
		</button>
	{:else}
		<Text size="sm" tone="muted">Every transaction in {label} is assigned.</Text>
	{/if}

	<!-- The goal lives on the same card as the month it judges: a ruled footer, not a fourth card. -->
	<div
		class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 border-t border-slate-200 pt-3"
	>
		<Text size="sm" tone="muted">
			Goal {month.goal_percent}% of income — <Money amount={target} currency="EUR" size="sm" /> on
			{label}'s income so far. Earlier months keep the goal they had.
		</Text>
		<Button size="sm" variant="outline" onclick={onAdjust}>Adjust goal</Button>
	</div>
</div>
