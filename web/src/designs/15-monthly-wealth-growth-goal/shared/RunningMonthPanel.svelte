<script lang="ts">
	// The month you are in: what you marked, what the goal asks for, and what is still open.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import GoalProgress from './GoalProgress.svelte';
	import { figureClass } from './figure';
	import { goalAmount, monthStatusLabel, sharePercent, type GoalMonth } from '../goal-data';

	type Props = {
		month: GoalMonth;
		/** Stack the amounts under the bar instead of beside it (narrow cards). */
		stacked?: boolean;
		onOpenUnassigned?: (month: GoalMonth) => void;
	};

	let { month, stacked = false, onOpenUnassigned }: Props = $props();

	const share = $derived(sharePercent(month));
	const remaining = $derived(Math.max(0, goalAmount(month) - month.contributed));
</script>

<div class="flex flex-col gap-3">
	<div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
		<span class={figureClass}>
			{share}% <span class="font-sans text-base text-slate-500">of income so far</span>
		</span>
		<Badge intent="neutral" variant="soft" size="md">{monthStatusLabel[month.status]}</Badge>
	</div>

	<div class={stacked ? 'flex flex-col gap-3' : 'flex flex-col gap-4 lg:flex-row lg:items-end'}>
		<div class="min-w-0 flex-1">
			<GoalProgress {month} />
		</div>

		<!-- Right-aligned, equal-width columns so the three amounts read as one set of numbers. -->
		<dl class={stacked ? 'flex justify-between gap-4' : 'flex shrink-0 justify-end gap-6'}>
			<div class="flex w-28 flex-col items-end gap-0.5">
				<dt class="text-xs text-slate-500">Income</dt>
				<dd><Money amount={month.income} currency="EUR" size="md" /></dd>
			</div>
			<div class="flex w-28 flex-col items-end gap-0.5">
				<dt class="text-xs text-slate-500">To wealth</dt>
				<dd><Money amount={month.contributed} currency="EUR" size="md" weight="semibold" /></dd>
			</div>
			<div class="flex w-28 flex-col items-end gap-0.5">
				<dt class="text-xs text-slate-500">Still to go</dt>
				<dd><Money amount={remaining} currency="EUR" size="md" /></dd>
			</div>
		</dl>
	</div>

	{#if month.unassigned > 0}
		<button
			type="button"
			class="w-fit text-sm text-sky-700 underline underline-offset-2 hover:text-sky-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
			onclick={() => onOpenUnassigned?.(month)}
		>
			{month.unassigned} transactions in {month.label} are not assigned yet
		</button>
	{:else}
		<Text size="sm" tone="muted">Every transaction in {month.label} is assigned.</Text>
	{/if}
</div>
