<script lang="ts">
	// The month you are in: what you marked, what the goal asks for, what is still open, and
	// the goal itself with the way back to it — so the running month is one card, not three.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import GoalProgress from '$lib/components/molecules/goal-progress/GoalProgress.svelte';
	import PurposeMarkers from '$lib/components/molecules/purpose-markers/PurposeMarkers.svelte';
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
		/** Jump to this month's unmarked money that could carry the missing purpose. */
		onOpenMissing?: (month: MonthStanding, purpose: 'income' | 'wealth') => void;
		onAdjust?: () => void;
	};

	let { month, stacked = false, onOpenMissing, onAdjust }: Props = $props();

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

	<!-- Whether income and a contribution are marked at all; the score means nothing until both are. -->
	<PurposeMarkers {month} {onOpenMissing} class="text-sm" />

	<!-- The goal lives on the same card as the month it judges: a ruled footer, not a fourth card.
	     One line, so this card does not set the height of the whole rail; the goal dialog says
	     that earlier months keep the goal they had. -->
	<div class="flex items-center justify-between gap-4 border-t border-slate-200 pt-3">
		<Text size="sm" tone="muted" class="min-w-0">
			Goal {month.goal_percent}% of income — <Money amount={target} currency="EUR" size="sm" /> so far
		</Text>
		<Button size="sm" variant="outline" class="shrink-0" onclick={onAdjust}>Adjust goal</Button>
	</div>
</div>
