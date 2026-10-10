<script lang="ts">
	// The recurring ledger: an aligned five-column table on desktop, the same items
	// stacked below `lg`. Both halves share one set of rows and one set of states, so
	// the two never disagree about what the account holds.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import RecurringItemRow from '$lib/components/molecules/recurring-item-row/RecurringItemRow.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { endedFromLabel, nextExpectedLabel, rhythmLabels } from '$lib/api/recurring';
	import type { RecurringItem } from '$lib/api/types';

	type Props = {
		rows: RecurringItem[];
		/** 'expenses' | 'income' | 'ended' — only the two headers that differ per group. */
		group?: 'expenses' | 'income' | 'ended';
		loading?: boolean;
		error?: string | null;
		emptyText: string;
		onOpen?: (item: RecurringItem) => void;
		class?: string;
	};

	let {
		rows,
		group = 'expenses',
		loading = false,
		error = null,
		emptyText,
		onOpen,
		class: className = ''
	}: Props = $props();

	function amounts(item: RecurringItem): number[] {
		return item.history.map((point) => scaledToNumber(point.amountCents));
	}
</script>

{#snippet nameCell(item: RecurringItem)}
	<div class="flex min-w-0 flex-col">
		<span class="truncate font-medium text-slate-900">{item.name}</span>
		<span class="text-xs text-slate-500">{item.linked_count} transactions</span>
	</div>
{/snippet}

{#snippet rhythmCell(item: RecurringItem)}
	<Badge intent="neutral" variant="soft" size="sm">{rhythmLabels[item.rhythm]}</Badge>
{/snippet}

<!-- Oldest amount first, then the shape of the amounts since: a reading of what happened,
     in a tone that carries no verdict. A rising expense is not an error, a falling one not
     a success, and this page passes judgement on neither. -->
{#snippet trendCell(item: RecurringItem)}
	{@const series = amounts(item)}
	{#if series.length > 1}
		<div class="flex items-center gap-2">
			<Money
				amount={series[0]}
				currency="EUR"
				size="sm"
				weight="normal"
				class="w-16 shrink-0 text-right text-slate-500"
			/>
			<Icon icon="heroicons:arrow-long-right" size="sm" class="shrink-0 text-slate-400" />
			<Sparkline
				data={series}
				tone="neutral"
				width={72}
				height={24}
				ariaLabel={`Amounts for ${item.name}, oldest to newest`}
			/>
		</div>
	{:else}
		<span class="text-sm text-slate-500">One amount so far</span>
	{/if}
{/snippet}

{#snippet amountCell(item: RecurringItem)}
	<Money amount={scaledToNumber(item.lastAmountCents)} currency="EUR" size="sm" />
{/snippet}

{#snippet nextCell(item: RecurringItem)}
	{#if item.ended_from}
		<span class="text-sm text-slate-500">Ended from {endedFromLabel(item.ended_from)}</span>
	{:else}
		<span class="text-sm whitespace-nowrap text-slate-700">
			{nextExpectedLabel(item.next_expected)}
		</span>
	{/if}
{/snippet}

<DataTable
	class={['hidden lg:flex', className].filter(Boolean).join(' ')}
	{rows}
	{loading}
	{error}
	{emptyText}
	sortKey="next"
	sortDirection="asc"
	onRowClick={onOpen}
	columns={[
		{ key: 'name', header: group === 'income' ? 'From' : 'To', cell: nameCell },
		{ key: 'rhythm', header: 'Rhythm', width: 'w-28', cell: rhythmCell },
		{ key: 'trend', header: 'Amount over time', width: 'w-56', cell: trendCell },
		{ key: 'amount', header: 'Last amount', align: 'right', width: 'w-32', cell: amountCell },
		{
			key: 'next',
			header: group === 'ended' ? 'Status' : 'Next expected',
			width: 'w-56',
			cell: nextCell
		}
	]}
/>

<div class={['min-h-0 flex-1 overflow-y-auto lg:hidden', className].filter(Boolean).join(' ')}>
	{#if loading}
		{#each [0, 1, 2, 3, 4] as index (index)}
			<div class="flex flex-col gap-2 border-b border-slate-100 px-4 py-3">
				<Skeleton width="60%" />
				<Skeleton width="40%" />
			</div>
		{/each}
	{:else if error}
		<div class="px-4 py-8">
			<p
				class="mx-auto max-w-md rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-center text-sm text-red-700"
				role="alert"
			>
				{error}
			</p>
		</div>
	{:else if rows.length === 0}
		<p class="px-6 py-10 text-center text-sm text-slate-500">{emptyText}</p>
	{:else}
		<ul>
			{#each rows as item (item.id)}
				<li class="border-b border-slate-100">
					<RecurringItemRow {item} {onOpen} />
				</li>
			{/each}
		</ul>
	{/if}
</div>
