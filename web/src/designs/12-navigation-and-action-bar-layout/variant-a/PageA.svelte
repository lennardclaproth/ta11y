<script lang="ts">
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import MockShell from '../shared/MockShell.svelte';
	import CashflowCharts from '../shared/CashflowCharts.svelte';
	import PortfolioCharts from '../shared/PortfolioCharts.svelte';
	import TransactionsTable from '../shared/TransactionsTable.svelte';
	import PositionsTable from '../shared/PositionsTable.svelte';
	import NavBarA from './NavBarA.svelte';
	import TableToolbarA from './TableToolbarA.svelte';
	import { icons } from '../shared/icons';
	import { selectedTransactionIds } from '../shared/mock-data';

	type Props = {
		page?: 'cashflow' | 'portfolio';
		overviewOpen?: boolean;
	};

	let { page = 'cashflow', overviewOpen = false }: Props = $props();

	let selectedIds = $state([...selectedTransactionIds]);
	let tab = $state('positions');
</script>

<MockShell>
	{#snippet top()}
		<NavBarA activeHref={page === 'cashflow' ? '/cashflow' : '/portfolio'} {overviewOpen} />
	{/snippet}

	<Heading level="h1" size="2xl" class="leading-none">
		{page === 'cashflow' ? 'Cashflow' : 'Portfolio'}
	</Heading>

	{#if page === 'cashflow'}
		<CashflowCharts />
	{:else}
		<PortfolioCharts>
			{#snippet actions()}
				<Button variant="outline" intent="secondary" size="sm" shape="default">
					<Icon icon={icons.rebuild} />Rebuild portfolio
				</Button>
			{/snippet}
		</PortfolioCharts>
	{/if}

	<Panel
		variant="muted"
		shape="square"
		shadow="none"
		padding="none"
		class="flex min-h-[32rem] flex-col overflow-hidden"
	>
		{#if page === 'cashflow'}
			<TableToolbarA title="Transactions" searchPlaceholder="Search description…">
				{#snippet actions()}
					<Button variant="outline" intent="secondary" shape="default">
						<Icon icon={icons.upload} />Import CSV
					</Button>
					<Button shape="default"><Icon icon={icons.plus} />Add transaction</Button>
				{/snippet}
			</TableToolbarA>
			<TransactionsTable bind:selectedIds>
				{#snippet bulkActions()}
					<Button size="sm" variant="ghost" intent="secondary" shape="default">
						<Icon icon={icons.tag} />Tag
					</Button>
				{/snippet}
			</TransactionsTable>
		{:else}
			<!-- Positions have no search today, so they do not get one; it lives on Transactions. -->
			<TableToolbarA title="Holdings" showSearch={tab === 'transactions'} searchPlaceholder="Search transactions…">
				{#snippet children()}
					<Tabs
						tabs={[
							{ value: 'positions', label: 'Positions' },
							{ value: 'transactions', label: 'Transactions' }
						]}
						bind:value={tab}
						ariaLabel="Portfolio view"
					/>
				{/snippet}
				{#snippet actions()}
					<Button variant="outline" intent="secondary" shape="default">
						<Icon icon={icons.upload} />Import CSV
					</Button>
					<Button shape="default"><Icon icon={icons.plus} />Add transaction</Button>
				{/snippet}
			</TableToolbarA>
			<PositionsTable />
		{/if}
	</Panel>
</MockShell>
