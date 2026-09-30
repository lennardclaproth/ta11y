<script lang="ts">
	import type { Snippet } from 'svelte';
	import CashflowTransactionsTable from '$lib/components/organisms/cashflow-transactions-table/CashflowTransactionsTable.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { tagOptions, transactionTotal, transactions } from './mock-data';

	type Props = {
		selectedIds?: string[];
		/** Selection actions rendered in the existing footer bar. */
		bulkActions?: Snippet;
		loading?: boolean;
		error?: string | null;
		rows?: typeof transactions;
		/**
		 * CashflowTransactionsTable hardcodes "No transactions match your filters" today, which
		 * cannot tell an empty account from a search with no hits. The prototype shows the copy
		 * each state needs; the build turns `emptyText` into a prop on that organism.
		 */
		emptyText?: string;
		emptyAction?: Snippet;
	};

	let {
		selectedIds = $bindable([]),
		bulkActions,
		loading = false,
		error = null,
		rows = transactions,
		emptyText,
		emptyAction
	}: Props = $props();

	const isEmpty = $derived(!loading && !error && rows.length === 0);
</script>

{#if isEmpty && emptyText}
	<div class="flex min-h-0 flex-1 flex-col items-center justify-center gap-3 px-6 py-16 text-center">
		<Text as="p" size="sm" tone="muted" class="max-w-sm">{emptyText}</Text>
		{@render emptyAction?.()}
	</div>
{:else}
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
{/if}
