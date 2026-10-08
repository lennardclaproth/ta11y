<script lang="ts">
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import { monthStatusLabel, sharePercent, type GoalMonth } from '../goal-data';

	type Props = {
		months: GoalMonth[];
		loading?: boolean;
		error?: string | null;
		emptyText?: string;
		onOpenUnassigned?: (month: GoalMonth) => void;
		class?: string;
	};

	let {
		months,
		loading = false,
		error = null,
		emptyText = 'No months yet. Mark your income and contributions to build the standing.',
		onOpenUnassigned,
		class: className = ''
	}: Props = $props();

	// A missed month needs attention; it is not a validation error, so it is amber, not red —
	// and it always carries the word "Missed" next to the colour.
	const badgeIntent = {
		met: 'success',
		missed: 'warning',
		'in-progress': 'neutral'
	} as const;
</script>

{#snippet monthCell(row: GoalMonth)}
	<div class="flex min-w-0 flex-col gap-0.5">
		<span class="font-medium text-slate-900">{row.label}</span>
		{#if row.unassigned > 0}
			<button
				type="button"
				class="w-fit whitespace-nowrap text-xs text-sky-700 underline underline-offset-2 hover:text-sky-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
				onclick={() => onOpenUnassigned?.(row)}
			>
				{row.unassigned} not assigned
			</button>
		{/if}
	</div>
{/snippet}

{#snippet incomeCell(row: GoalMonth)}
	<Money amount={row.income} currency="EUR" size="sm" />
{/snippet}

{#snippet contributedCell(row: GoalMonth)}
	<Money amount={row.contributed} currency="EUR" size="sm" />
{/snippet}

{#snippet shareCell(row: GoalMonth)}
	<span class="font-medium text-slate-900 tabular-nums">{sharePercent(row)}%</span>
{/snippet}

{#snippet goalCell(row: GoalMonth)}
	<span class="text-slate-500 tabular-nums">{row.goalPercent}%</span>
{/snippet}

{#snippet resultCell(row: GoalMonth)}
	<Badge intent={badgeIntent[row.status]} variant="soft" size="sm">
		{monthStatusLabel[row.status]}
	</Badge>
{/snippet}

<DataTable
	rows={months}
	getRowId={(row: GoalMonth) => row.month}
	{loading}
	{error}
	{emptyText}
	class={className}
	columns={[
		{ key: 'month', header: 'Month', cell: monthCell },
		{ key: 'income', header: 'Income', align: 'right', width: 'w-36', cell: incomeCell },
		{ key: 'contributed', header: 'To wealth', align: 'right', width: 'w-36', cell: contributedCell },
		{ key: 'share', header: 'Share', align: 'right', width: 'w-24', cell: shareCell },
		{ key: 'goal', header: 'Goal', align: 'right', width: 'w-24', cell: goalCell },
		{ key: 'result', header: 'Result', width: 'w-32', cell: resultCell }
	]}
/>
