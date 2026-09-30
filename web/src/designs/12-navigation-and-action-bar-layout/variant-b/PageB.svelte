<script lang="ts">
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import MockShell from '../shared/MockShell.svelte';
	import CashflowCharts from '../shared/CashflowCharts.svelte';
	import PortfolioCharts from '../shared/PortfolioCharts.svelte';
	import TransactionsTable from '../shared/TransactionsTable.svelte';
	import PositionsTable from '../shared/PositionsTable.svelte';
	import NavBarB from './NavBarB.svelte';
	import TableToolbarB from './TableToolbarB.svelte';
	import { icons } from '../shared/icons';
	import { period, selectedTransactionIds } from '../shared/mock-data';

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
		<NavBarB activeHref={page === 'cashflow' ? '/cashflow' : '/portfolio'} {overviewOpen} />
	{/snippet}

	<!-- The period control is hidden in the overview, so the page states it in words. -->
	<div class="flex flex-wrap items-baseline justify-between gap-2">
		<Heading level="h1" size="2xl" class="leading-none">
			{page === 'cashflow' ? 'Cashflow' : 'Portfolio'}
		</Heading>
		<Text as="span" size="sm" tone="muted">Showing {period.label}</Text>
	</div>

	{#if page === 'cashflow'}
		<CashflowCharts />
	{:else}
		<PortfolioCharts>
			{#snippet actions()}
				<Button variant="ghost" intent="secondary" size="sm" shape="default">
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
			<TableToolbarB searchPlaceholder="Search description…" searchLabel="Search transactions">
				{#snippet actions()}
					<Button variant="outline" intent="secondary" shape="default">
						<Icon icon={icons.upload} />Import CSV
					</Button>
					<Button shape="default"><Icon icon={icons.plus} />Add transaction</Button>
				{/snippet}
			</TableToolbarB>
			<TransactionsTable bind:selectedIds>
				{#snippet bulkActions()}
					<Button size="sm" variant="ghost" intent="secondary" shape="default">
						<Icon icon={icons.tag} />Tag
					</Button>
				{/snippet}
			</TransactionsTable>
		{:else}
			<!-- Positions have no search today, so they do not get one; it lives on Transactions. -->
			<TableToolbarB
				searchPlaceholder="Search transactions…"
				searchLabel="Search transactions"
				showSearch={tab === 'transactions'}
			>
				{#snippet before()}
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
			</TableToolbarB>
			<PositionsTable />
		{/if}
	</Panel>
</MockShell>
