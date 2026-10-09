<script lang="ts">
	// Proposal for a new molecule: one calendar month of the standing, read as an entry in a ruled
	// column rather than a table row. Composes Heading / Text / Money / Badge / ProgressBar.
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import GoalProgress from './GoalProgress.svelte';
	import { monthStatusLabel, type GoalMonth } from '../goal-data';

	type Props = {
		month: GoalMonth;
		onOpenUnassigned?: (month: GoalMonth) => void;
	};

	let { month, onOpenUnassigned }: Props = $props();

	const badgeIntent = {
		met: 'success',
		missed: 'warning',
		'in-progress': 'neutral'
	} as const;
</script>

<article class="flex flex-col gap-3 border-t border-slate-400 py-4">
	<div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
		<Heading level="h3" size="md">{month.label}</Heading>
		<Badge intent={badgeIntent[month.status]} variant="soft" size="md">
			{monthStatusLabel[month.status]}
		</Badge>
	</div>

	<div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:gap-8">
		<div class="min-w-0 flex-1">
			<GoalProgress {month} />
		</div>

		<!-- Fixed column widths so the amounts line up down the whole column, not just inside one entry. -->
		<dl class="flex shrink-0 justify-end gap-6">
			<div class="flex w-28 flex-col items-end gap-0.5">
				<dt class="text-xs text-slate-500">Income</dt>
				<dd><Money amount={month.income} currency="EUR" size="md" /></dd>
			</div>
			<div class="flex w-28 flex-col items-end gap-0.5">
				<dt class="text-xs text-slate-500">To wealth</dt>
				<dd><Money amount={month.contributed} currency="EUR" size="md" weight="semibold" /></dd>
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
		<Text size="xs" tone="muted">All transactions in this month are assigned.</Text>
	{/if}
</article>
