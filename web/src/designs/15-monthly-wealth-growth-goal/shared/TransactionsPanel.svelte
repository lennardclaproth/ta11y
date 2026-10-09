<script lang="ts">
	// The ledger half of Cashflow, with the marking surface the design needs: a Purpose column, a
	// Purpose header filter, bulk actions in the selection footer and "mark everything that matches
	// the filter" in the ledger header.
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import PurposeTransactionsTable from './PurposeTransactionsTable.svelte';
	import { goalTransactions, type GoalTransaction } from '../goal-data';

	type Props = {
		rows?: GoalTransaction[];
		selectedIds?: string[];
		purposeFilter?: string[];
		title?: string;
		/** Caption on the ledger header, e.g. the month the standing sent you to. */
		scope?: string | null;
		total?: number;
		loading?: boolean;
		error?: string | null;
		emptyText?: string;
	};

	let {
		rows = goalTransactions,
		selectedIds = $bindable([]),
		purposeFilter = $bindable([]),
		title = 'Transactions',
		scope = null,
		total,
		loading = false,
		error = null,
		emptyText = 'No transactions match your filters'
	}: Props = $props();

	// "Mark all matches" acts on the filter, so it names the number the filter actually returns.
	const matchCount = $derived(total ?? rows.length);
</script>

<LedgerToolbar {title} actionLabel="Add transaction" onAdd={() => {}}>
	{#if scope}
		<Badge intent="info" variant="soft" size="sm">{scope}</Badge>
	{/if}
	{#if !loading && !error && matchCount > 0}
		<Button size="sm" variant="ghost" intent="secondary">
			Mark all {matchCount} matches…
		</Button>
	{/if}
</LedgerToolbar>

<PurposeTransactionsTable
	{rows}
	{total}
	{loading}
	{error}
	{emptyText}
	bind:selectedIds
	bind:purposeFilter
/>
