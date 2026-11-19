<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import PortfolioTransactionDrawer from './PortfolioTransactionDrawer.svelte';
	import type { PortfolioTransaction } from '$lib/api/types';

	const { Story } = defineMeta({
		title: 'Organisms/PortfolioTransactionDrawer',
		component: PortfolioTransactionDrawer,
		tags: ['autodocs']
	});

	// Fixed days and invented amounts so the stories read the same whenever they are opened.
	const today = '2026-09-30';

	const manualBuy: PortfolioTransaction = {
		id: 'ptx-manual',
		account_id: 'acc-demo',
		origin: 'MANUAL',
		source: 'DEGIRO',
		occurred_at: '2026-07-14T00:00:00Z',
		type: 'BUY',
		listing_id: 'lst-vwrl',
		isin: 'IE00B3RBWM25',
		symbol: 'VWRL',
		description: 'Monthly top-up',
		amount: '592.20',
		quantity: '5',
		unit_price: '118.44',
		created_at: '2026-07-14T09:00:00Z',
		updated_at: '2026-07-14T09:00:00Z'
	};

	const manualCash: PortfolioTransaction = {
		...manualBuy,
		id: 'ptx-cash',
		type: 'CASH',
		listing_id: null,
		isin: null,
		symbol: null,
		description: 'Deposit',
		amount: '1000',
		quantity: '1',
		unit_price: '0'
	};

	const importedDividend: PortfolioTransaction = {
		...manualBuy,
		id: 'ptx-imported',
		origin: 'IMPORT',
		type: 'DIVIDEND',
		description: 'VANGUARD FTSE ALL-WORLD',
		amount: '14.32',
		quantity: '0',
		unit_price: '0'
	};
</script>

<Story name="Manual buy" asChild>
	<PortfolioTransactionDrawer open transaction={manualBuy} editable originLabel="Manual" {today} />
</Story>

<Story name="Manual cash, no listing" asChild>
	<PortfolioTransactionDrawer open transaction={manualCash} editable originLabel="Manual" {today} />
</Story>

<Story name="Saving" asChild>
	<PortfolioTransactionDrawer
		open
		transaction={manualBuy}
		editable
		originLabel="Manual"
		{today}
		saving
	/>
</Story>

<Story name="Deleting" asChild>
	<PortfolioTransactionDrawer
		open
		transaction={manualBuy}
		editable
		originLabel="Manual"
		{today}
		deleting
	/>
</Story>

<Story name="Refused as a duplicate" asChild>
	<PortfolioTransactionDrawer
		open
		transaction={manualBuy}
		editable
		originLabel="Manual"
		{today}
		error="A transaction with the same listing and amount already exists on 14 Jul 2026. Change the date or the amount."
	/>
</Story>

<!-- Save and Delete report in the same band, so the heading names the button that was pressed
     rather than describing a change that was never attempted. -->
<Story name="Refused delete" asChild>
	<PortfolioTransactionDrawer
		open
		transaction={manualBuy}
		editable
		originLabel="Manual"
		{today}
		errorTitle="This transaction was not deleted"
		error="Could not delete the transaction. Try again."
	/>
</Story>

<Story name="Imported, read-only" asChild>
	<PortfolioTransactionDrawer open transaction={importedDividend} originLabel="DEGIRO" {today} />
</Story>
