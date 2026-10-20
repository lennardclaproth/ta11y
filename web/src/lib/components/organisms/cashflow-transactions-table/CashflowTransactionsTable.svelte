<script lang="ts">
	import type { Snippet } from 'svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import FooterBar from '$lib/components/organisms/footer-bar/FooterBar.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import TextFilter from '$lib/components/molecules/text-filter/TextFilter.svelte';
	import SelectFilter from '$lib/components/molecules/select-filter/SelectFilter.svelte';
	import DirectionFilter from '$lib/components/molecules/direction-filter/DirectionFilter.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { cashflowOriginLabel, isManualCashflowTransaction } from '$lib/api/transactions';
	import { purposeLabel } from '$lib/api/wealthgoal';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { purposeFilterOptions } from './cashflow-transactions-table.types';
	import type { SortDirection } from '$lib/components/organisms/data-table/data-table.types';
	import type { CashflowDirection, CashflowTransaction } from '$lib/api/types';

	type Props = {
		rows: CashflowTransaction[];
		loading?: boolean;
		error?: string | null;
		total?: number;
		limit?: number;
		offset?: number;
		selectedIds?: string[];
		sortKey?: string;
		sortDirection?: SortDirection;
		descriptionFilter?: string;
		tagFilter?: string[];
		directionFilter?: CashflowDirection | null;
		/** Goal purposes to show: `income`, `wealth`, `none`. Empty is no filter. */
		purposeFilter?: string[];
		tagOptions?: { value: string; label: string }[];
		/** Shown when there are no rows; say "nothing yet" and "nothing matched" differently. */
		emptyText?: string;
		onSort?: (key: string, direction: SortDirection) => void;
		onPageChange?: (offset: number) => void;
		onLimitChange?: (limit: number) => void;
		onFilterChange?: () => void;
		/** Open one transaction (e.g. in a detail drawer). */
		onRowClick?: (row: CashflowTransaction) => void;
		/** Bulk actions for the footer when rows are selected. */
		bulkActions?: Snippet;
		class?: string;
	};

	let {
		rows,
		loading = false,
		error = null,
		total = 0,
		limit = $bindable(25),
		offset = $bindable(0),
		selectedIds = $bindable([]),
		sortKey = 'date',
		sortDirection = 'desc',
		descriptionFilter = $bindable(''),
		tagFilter = $bindable([]),
		directionFilter = $bindable(null),
		purposeFilter = $bindable([]),
		tagOptions = [],
		emptyText = 'No transactions match your filters',
		onSort,
		onPageChange,
		onLimitChange,
		onFilterChange,
		onRowClick,
		bulkActions,
		class: className = ''
	}: Props = $props();

	function fmtDate(value: string): string {
		return formatDisplayDate(value.slice(0, 10));
	}
</script>

{#snippet descriptionFilterControl()}
	<TextFilter
		bind:value={descriptionFilter}
		label="Description"
		onApply={() => onFilterChange?.()}
	/>
{/snippet}

{#snippet tagFilterControl()}
	<SelectFilter
		options={tagOptions}
		bind:selected={tagFilter}
		label="Tags"
		onApply={() => onFilterChange?.()}
	/>
{/snippet}

{#snippet directionFilterControl()}
	<DirectionFilter bind:value={directionFilter} onApply={() => onFilterChange?.()} />
{/snippet}

{#snippet purposeFilterControl()}
	<SelectFilter
		options={purposeFilterOptions}
		bind:selected={purposeFilter}
		label="Purpose"
		onApply={() => onFilterChange?.()}
	/>
{/snippet}

{#snippet tagCell(row: CashflowTransaction)}
	{#if row.tag}
		<Badge intent="neutral" variant="soft" size="sm">{row.tag}</Badge>
	{:else}
		<span class="text-slate-500">Untagged</span>
	{/if}
{/snippet}

{#snippet directionCell(row: CashflowTransaction)}
	<Badge intent={row.direction === 'in' ? 'success' : 'error'} variant="soft" size="sm">
		{row.direction === 'in' ? 'In' : 'Out'}
	</Badge>
{/snippet}

{#snippet amountCell(row: CashflowTransaction)}
	<Money amount={scaledToNumber(row.amountCents)} currency="EUR" size="sm" />
{/snippet}

<!-- What the row counts as towards the monthly goal. Outline rather than soft, so the
     purpose never reads as a second direction badge. -->
{#snippet purposeCell(row: CashflowTransaction)}
	{#if row.purpose === 'income'}
		<Badge intent="success" variant="outline" size="sm">{purposeLabel.income}</Badge>
	{:else if row.purpose === 'wealth'}
		<Badge intent="info" variant="outline" size="sm">{purposeLabel.wealth}</Badge>
	{:else}
		<span class="text-slate-500">{purposeLabel['']}</span>
	{/if}
{/snippet}

<!-- Where the row came from, as a word rather than only a badge colour: it is what decides
     whether its date can still be changed. -->
{#snippet enteredCell(row: CashflowTransaction)}
	<Badge intent={isManualCashflowTransaction(row) ? 'info' : 'neutral'} variant="soft" size="sm">
		{cashflowOriginLabel(row)}
	</Badge>
{/snippet}

<DataTable
	{rows}
	{loading}
	{error}
	bind:selectedIds
	selectable
	{sortKey}
	{sortDirection}
	{onSort}
	{onRowClick}
	{emptyText}
	class={className}
	columns={[
		{
			key: 'date',
			header: 'Date',
			sortKey: 'date',
			value: (r: CashflowTransaction) => fmtDate(r.date)
		},
		{
			key: 'description',
			header: 'Description',
			sortKey: 'description',
			value: (r: CashflowTransaction) => r.description,
			filter: descriptionFilterControl
		},
		{ key: 'tag', header: 'Tag', sortKey: 'tag', cell: tagCell, filter: tagFilterControl },
		{ key: 'source', header: 'Entered', sortKey: 'source', width: 'w-32', cell: enteredCell },
		{ key: 'direction', header: 'Direction', cell: directionCell, filter: directionFilterControl },
		{
			key: 'purpose',
			header: 'Purpose',
			width: 'w-40',
			cell: purposeCell,
			filter: purposeFilterControl
		},
		{ key: 'amount', header: 'Amount', sortKey: 'amount', align: 'right', cell: amountCell }
	]}
>
	{#snippet footer()}
		<FooterBar
			{total}
			bind:limit
			bind:offset
			selectedCount={selectedIds.length}
			{onPageChange}
			{onLimitChange}
			actions={bulkActions}
		/>
	{/snippet}
</DataTable>
