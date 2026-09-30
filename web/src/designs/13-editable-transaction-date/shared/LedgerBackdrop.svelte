<script lang="ts">
	// Shared page context for the #13 prototypes: the Cashflow ledger the create/edit surfaces
	// open on top of. Composes the existing PageContentTemplate, LedgerToolbar and DataTable.
	import type { Snippet } from 'svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { CashflowTransaction } from '$lib/api/types';
	import { designTransactions, isManual, sourceLabel } from './design-data';

	type Props = {
		rows?: CashflowTransaction[];
		loading?: boolean;
		/** Distinguishes an empty account from a filter that matches nothing. */
		emptyText?: string;
		/** Persistent page-level message, rendered above the toolbar. */
		banner?: Snippet;
		/** Toolbar content, e.g. the active filters of the "No matches" state. */
		filters?: Snippet;
		/** Extra right-hand column, e.g. a per-row action. */
		actionCell?: Snippet<[CashflowTransaction]>;
		onRowClick?: (row: CashflowTransaction) => void;
	};

	let {
		rows = designTransactions,
		loading = false,
		emptyText = 'No transactions yet. Add one to start your ledger.',
		banner,
		filters,
		actionCell,
		onRowClick
	}: Props = $props();
</script>

{#snippet sourceCell(row: CashflowTransaction)}
	<Badge intent={isManual(row) ? 'info' : 'neutral'} variant="soft" size="sm">
		{sourceLabel(row)}
	</Badge>
{/snippet}

{#snippet amountCell(row: CashflowTransaction)}
	<Money amount={scaledToNumber(row.amountCents)} currency="EUR" size="sm" />
{/snippet}

<div class="flex h-screen flex-col bg-taupe-100 pt-5">
	<PageContentTemplate>
		{#if banner}
			{@render banner()}
		{/if}
		<LedgerToolbar title="Transactions" actionLabel="Add transaction" onAdd={() => {}}>
			{#if filters}
				{@render filters()}
			{/if}
		</LedgerToolbar>
		<DataTable
			{rows}
			{loading}
			{emptyText}
			{onRowClick}
			sortKey="date"
			sortDirection="desc"
			columns={[
				{
					key: 'date',
					header: 'Date',
					sortKey: 'date',
					width: 'w-40',
					value: (r: CashflowTransaction) => formatDisplayDate(r.date.slice(0, 10))
				},
				{
					key: 'description',
					header: 'Description',
					value: (r: CashflowTransaction) => r.description
				},
				{ key: 'source', header: 'Entered', width: 'w-28', cell: sourceCell },
				{ key: 'tag', header: 'Tag', value: (r: CashflowTransaction) => r.tag },
				{ key: 'amount', header: 'Amount', align: 'right', width: 'w-32', cell: amountCell },
				...(actionCell
					? [
							{
								key: 'actions',
								header: '',
								align: 'right' as const,
								width: 'w-32',
								cell: actionCell
							}
						]
					: [])
			]}
		/>
	</PageContentTemplate>
</div>
