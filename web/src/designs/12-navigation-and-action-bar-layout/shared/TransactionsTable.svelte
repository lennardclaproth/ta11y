<script lang="ts">
	import type { Snippet } from 'svelte';
	import CashflowTransactionsTable from '$lib/components/organisms/cashflow-transactions-table/CashflowTransactionsTable.svelte';
	import { tagOptions, transactionTotal, transactions } from './mock-data';

	type Props = {
		selectedIds?: string[];
		/** Selection actions rendered in the existing footer bar. */
		bulkActions?: Snippet;
		loading?: boolean;
		error?: string | null;
		rows?: typeof transactions;
	};

	let {
		selectedIds = $bindable([]),
		bulkActions,
		loading = false,
		error = null,
		rows = transactions
	}: Props = $props();
</script>

<CashflowTransactionsTable
	{rows}
	{loading}
	{error}
	total={rows.length === 0 ? 0 : transactionTotal}
	bind:selectedIds
	{tagOptions}
	{bulkActions}
	class="min-h-0 flex-1"
/>
