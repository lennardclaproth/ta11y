<script lang="ts">
	// Final design for issue #13 — "Transaction date".
	// Create: variant C's ruled date header, opening on today, with the Today shortcut.
	// Change afterwards: variant B's detail drawer, only the date editable, manual rows only.
	// The root cause of the original bug is repaired in shared/InFormDatePicker.svelte.
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import SelectFilter from '$lib/components/molecules/select-filter/SelectFilter.svelte';
	import LedgerBackdrop from '../shared/LedgerBackdrop.svelte';
	import NewTransactionDialog from './NewTransactionDialog.svelte';
	import TransactionDrawer from './TransactionDrawer.svelte';
	import PortfolioLedger from './PortfolioLedger.svelte';
	import { backdatedDate, designTransactions } from '../shared/design-data';

	type Scene =
		| 'create'
		| 'create-backdated'
		| 'create-error'
		| 'edit'
		| 'edit-imported'
		| 'edit-saving'
		| 'edit-error'
		| 'loading'
		| 'empty'
		| 'no-matches'
		| 'portfolio-busy';

	type Props = { scene?: Scene };

	let { scene = 'create' }: Props = $props();

	const manualRow = designTransactions[0];
	const importedRow = designTransactions[1];
</script>

{#snippet activeFilters()}
	<div class="flex items-center gap-2">
		<SearchInput value="bicycle" size="sm" ariaLabel="Search transactions" class="w-56" />
		<SelectFilter
			label="Entered"
			selected={['manual']}
			options={[
				{ value: 'manual', label: 'Manual' },
				{ value: 'import', label: 'Import' }
			]}
		/>
	</div>
{/snippet}

{#if scene === 'portfolio-busy'}
	<PortfolioLedger />
{:else if scene === 'loading'}
	<LedgerBackdrop rows={[]} loading />
{:else if scene === 'empty'}
	<LedgerBackdrop rows={[]} />
{:else if scene === 'no-matches'}
	<LedgerBackdrop
		rows={[]}
		filters={activeFilters}
		emptyText="No transactions match your search and filters."
	/>
{:else}
	<LedgerBackdrop onRowClick={() => {}} />

	{#if scene === 'create'}
		<NewTransactionDialog />
	{:else if scene === 'create-backdated'}
		<NewTransactionDialog initialDate={backdatedDate} showCalendar />
	{:else if scene === 'create-error'}
		<NewTransactionDialog
			initialDate={backdatedDate}
			error="A transaction with the same amount and description already exists on 14 Jul 2026. Pick another day or cancel."
		/>
	{:else if scene === 'edit'}
		<TransactionDrawer row={manualRow} date={backdatedDate} />
	{:else if scene === 'edit-imported'}
		<TransactionDrawer row={importedRow} date={importedRow.date.slice(0, 10)} />
	{:else if scene === 'edit-saving'}
		<TransactionDrawer row={manualRow} date={backdatedDate} saving />
	{:else if scene === 'edit-error'}
		<TransactionDrawer
			row={manualRow}
			date={backdatedDate}
			error="A transaction with the same amount and description already exists on 14 Jul 2026. Pick another day or keep the current date."
		/>
	{/if}
{/if}
