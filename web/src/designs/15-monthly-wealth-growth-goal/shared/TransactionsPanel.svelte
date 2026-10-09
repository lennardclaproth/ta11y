<script lang="ts">
	// The ledger half of Cashflow, with the marking surface every variant needs: a Purpose column,
	// a Purpose header filter, bulk actions in the selection footer and "mark everything that
	// matches the filter" in the ledger header.
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import PurposeTransactionsTable from './PurposeTransactionsTable.svelte';
	import { goalTransactions } from '../goal-data';

	type Props = {
		selectedIds?: string[];
		title?: string;
	};

	let { selectedIds = $bindable([]), title = 'Transactions' }: Props = $props();
</script>

<LedgerToolbar {title} actionLabel="Add transaction" onAdd={() => {}}>
	<Button size="sm" variant="ghost" intent="secondary">Mark all 24 matches…</Button>
</LedgerToolbar>

<PurposeTransactionsTable rows={goalTransactions} bind:selectedIds />
