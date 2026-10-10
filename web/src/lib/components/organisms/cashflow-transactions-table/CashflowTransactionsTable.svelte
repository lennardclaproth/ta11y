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
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import type { Column, SortDirection } from '$lib/components/organisms/data-table/data-table.types';
	import type { CashflowDirection, CashflowTransaction, IgnoreRule } from '$lib/api/types';

	type Props = {
		rows: CashflowTransaction[];
		loading?: boolean;
		error?: string | null;
		total?: number;
		/**
		 * Ignored rows are in the list. The "Ignored by" column and the row-level Restore
		 * only make sense then, so they appear with them rather than sitting empty.
		 */
		showIgnored?: boolean;
		/** The account's ignore rules, used to name the rule that ignored a row. */
		ignoreRules?: IgnoreRule[];
		/** Put one ignored transaction back in the ledger. */
		onRestore?: (row: CashflowTransaction) => void;
		limit?: number;
		offset?: number;
		selectedIds?: string[];
		sortKey?: string;
		sortDirection?: SortDirection;
		descriptionFilter?: string;
		tagFilter?: string[];
		directionFilter?: CashflowDirection | null;
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
		showIgnored = false,
		ignoreRules = [],
		onRestore,
		limit = $bindable(25),
		offset = $bindable(0),
		selectedIds = $bindable([]),
		sortKey = 'date',
		sortDirection = 'desc',
		descriptionFilter = $bindable(''),
		tagFilter = $bindable([]),
		directionFilter = $bindable(null),
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

	/** The rule that ignored a row, resolved from the account's rules. */
	function ruleFor(row: CashflowTransaction): IgnoreRule | undefined {
		if (!row.ignored_by_rule_id) return undefined;
		return ignoreRules.find((rule) => rule.id === row.ignored_by_rule_id);
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

<!-- Where the row came from, as a word rather than only a badge colour: it is what decides
     whether its date can still be changed. -->
{#snippet enteredCell(row: CashflowTransaction)}
	<Badge intent={isManualCashflowTransaction(row) ? 'info' : 'neutral'} variant="soft" size="sm">
		{cashflowOriginLabel(row)}
	</Badge>
{/snippet}

<!-- The rule is the reason a row left the totals, so it is named rather than implied. A row
     nobody ignored says so in words too, not by being the only one without a badge. -->
{#snippet ignoredByCell(row: CashflowTransaction)}
	{@const rule = ruleFor(row)}
	{#if row.ignored && rule}
		<Badge intent="info" variant="soft" size="sm">{rule.name}</Badge>
	{:else if row.ignored}
		<Badge intent="neutral" variant="soft" size="sm">By hand</Badge>
	{:else}
		<span class="text-sm text-slate-500">Counted</span>
	{/if}
{/snippet}

{#snippet restoreCell(row: CashflowTransaction)}
	{#if row.ignored}
		<Button
			size="sm"
			variant="ghost"
			intent="secondary"
			shape="default"
			onclick={() => onRestore?.(row)}
		>
			Restore
		</Button>
	{/if}
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
		...(showIgnored
			? ([
					{ key: 'ignoredBy', header: 'Ignored by', width: 'w-52', cell: ignoredByCell }
				] as Column<CashflowTransaction>[])
			: []),
		{ key: 'direction', header: 'Direction', cell: directionCell, filter: directionFilterControl },
		{ key: 'amount', header: 'Amount', sortKey: 'amount', align: 'right', cell: amountCell },
		...(showIgnored
			? ([
					{ key: 'restore', header: '', align: 'right', width: 'w-28', cell: restoreCell }
				] as Column<CashflowTransaction>[])
			: [])
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
