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
	import NavBarC from './NavBarC.svelte';
	import TableHeaderC from './TableHeaderC.svelte';
	import RuledAction from './RuledAction.svelte';
	import { icons } from '../shared/icons';
	import { period, positions, selectedTransactionIds, transactionTotal } from '../shared/mock-data';

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
		<NavBarC activeHref={page === 'cashflow' ? '/cashflow' : '/portfolio'} {overviewOpen} />
	{/snippet}

	<div class="flex flex-wrap items-baseline justify-between gap-2 pt-1">
		<Heading level="h1" size="2xl" class="leading-none">
			{page === 'cashflow' ? 'Cashflow' : 'Portfolio'}
		</Heading>
		<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">
			Period · {period.label}
		</Text>
	</div>

	{#if page === 'cashflow'}
		<CashflowCharts />
	{:else}
		<PortfolioCharts>
			{#snippet actions()}
				<RuledAction label="Rebuild portfolio" icon={icons.rebuild} />
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
			<TableHeaderC
				title="Transactions"
				meta="{transactionTotal} rows · {selectedIds.length} selected"
				searchPlaceholder="Search description…"
			>
				{#snippet actions()}
					<RuledAction label="Tag selected" icon={icons.tag} />
					<RuledAction label="Import CSV" icon={icons.upload} />
					<Button shape="default"><Icon icon={icons.plus} />Add transaction</Button>
				{/snippet}
			</TableHeaderC>
			<TransactionsTable bind:selectedIds />
		{:else}
			<!-- Positions have no search today, so they do not get one; it lives on Transactions. -->
			<TableHeaderC
				title="Holdings"
				meta="{positions.length} positions"
				showSearch={tab === 'transactions'}
				searchPlaceholder="Search transactions…"
			>
				{#snippet before()}
					<Tabs
						tabs={[
							{ value: 'positions', label: 'Positions' },
							{ value: 'transactions', label: 'Transactions' }
						]}
						bind:value={tab}
						size="sm"
						ariaLabel="Portfolio view"
					/>
				{/snippet}
				{#snippet actions()}
					<RuledAction label="Import CSV" icon={icons.upload} />
					<Button shape="default"><Icon icon={icons.plus} />Add transaction</Button>
				{/snippet}
			</TableHeaderC>
			<PositionsTable />
		{/if}
	</Panel>
</MockShell>
