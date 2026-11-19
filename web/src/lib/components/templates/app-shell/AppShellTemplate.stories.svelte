<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import AppShellTemplate from './AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import CashflowTransactionsTable from '$lib/components/organisms/cashflow-transactions-table/CashflowTransactionsTable.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { cashflowTransactions } from '$lib/data/fixtures/cashflow';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';

	const kpis: KpiItem[] = [
		{ label: 'Incoming', amount: 5400, currency: 'EUR', change: 4.9 },
		{ label: 'Outgoing', amount: 3815.5, currency: 'EUR', change: -2.1 },
		{ label: 'Net', amount: 1584.5, currency: 'EUR', change: 12.3 }
	];
	const tagOptions = [
		{ value: 'salary', label: 'salary' },
		{ value: 'rent', label: 'rent' },
		{ value: 'groceries', label: 'groceries' }
	];

	const { Story } = defineMeta({
		title: 'Templates/AppShellTemplate',
		component: AppShellTemplate,
		tags: ['autodocs']
	});
</script>

<!--
  The whole shell in one place: navigation and the account overview on top, the page's own
  heading below it, and everything that acts on the rows in the ledger header.
-->
<Story name="Cashflow page" asChild>
	<AppShellTemplate>
		{#snippet top()}
			<TopNavbar />
		{/snippet}

		<PageContentTemplate title="Cashflow">
			{#snippet analytics()}
				<KpiRow items={kpis} columns={3} />
			{/snippet}

			<LedgerToolbar
				title="Transactions"
				meta="{cashflowTransactions.length} rows"
				showSearch
				searchPlaceholder="Search description…"
			>
				{#snippet actions()}
					<Button shape="default"><Icon icon="heroicons:plus" />Add transaction</Button>
				{/snippet}
			</LedgerToolbar>
			<CashflowTransactionsTable
				rows={cashflowTransactions.slice(0, 10)}
				total={cashflowTransactions.length}
				{tagOptions}
			/>
		</PageContentTemplate>
	</AppShellTemplate>
</Story>
