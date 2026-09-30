<script lang="ts">
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import MockShell from '../shared/MockShell.svelte';
	import CashflowCharts from '../shared/CashflowCharts.svelte';
	import PortfolioCharts from '../shared/PortfolioCharts.svelte';
	import TransactionsTable from '../shared/TransactionsTable.svelte';
	import PositionsTable from '../shared/PositionsTable.svelte';
	import ListingsTable from '../shared/ListingsTable.svelte';
	import NavBarFinal from './NavBarFinal.svelte';
	import TableHeader from './TableHeader.svelte';
	import RuledAction from './RuledAction.svelte';
	import PrimaryAction from './PrimaryAction.svelte';
	import { icons } from '../shared/icons';
	import {
		adminListings,
		period,
		positions,
		selectedTransactionIds,
		transactionTotal,
		transactions
	} from '../shared/mock-data';

	type Page = 'cashflow' | 'portfolio' | 'admin';
	type State = 'default' | 'loading' | 'empty' | 'error' | 'no-matches';

	type Props = {
		page?: Page;
		/** Named `view` because `state` would shadow the `$state` rune. */
		view?: State;
		overviewOpen?: boolean;
		/** Opens the period picker so the prototype can show the presets inside it. */
		periodOpen?: boolean;
	};

	let {
		page = 'cashflow',
		view = 'default',
		overviewOpen = false,
		periodOpen = false
	}: Props = $props();

	let selectedIds = $state(view === 'default' ? [...selectedTransactionIds] : []);
	let tab = $state('positions');
	let query = $state(view === 'no-matches' ? 'grocerys' : '');
	let closedFilter = $state('open');

	const href = { cashflow: '/cashflow', portfolio: '/portfolio', admin: '/admin/listings' };
	const heading = { cashflow: 'Cashflow', portfolio: 'Portfolio', admin: 'Listings' };

	const loading = $derived(view === 'loading');
	const failed = $derived(view === 'error');
	const rows = $derived(view === 'default' ? transactions : []);
	const positionRows = $derived(view === 'default' ? positions : []);

	const tableMeta = $derived.by(() => {
		if (loading) return 'Loading…';
		if (failed) return 'Could not load';
		if (view === 'empty') return 'No records yet';
		if (view === 'no-matches') return `0 of ${transactionTotal} rows`;
		return `${transactionTotal} rows · ${selectedIds.length} selected`;
	});

	const emptyText = $derived(
		view === 'no-matches'
			? `No rows match “${query}”. Clear the search to see all rows.`
			: 'Nothing here yet. Add a record or import a file to get started.'
	);
</script>

<MockShell>
	{#snippet top()}
		<NavBarFinal
			activeHref={href[page]}
			showPeriod={page !== 'admin'}
			{loading}
			change={page === 'admin' ? 'no-period' : view === 'empty' ? 'none' : 'value'}
			{overviewOpen}
			{periodOpen}
		/>
	{/snippet}

	<div class="flex flex-wrap items-baseline justify-between gap-2 pt-1">
		<Heading level="h1" size="2xl" class="leading-none">{heading[page]}</Heading>
		{#if page === 'admin'}
			<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">
				Reference data
			</Text>
		{/if}
	</div>

	{#if failed}
		<Alert intent="error" title="Could not load this page">
			<div class="flex flex-wrap items-center gap-3">
				<span>The request failed. Your period and search are still set.</span>
				<Button size="sm" variant="outline" shape="default">Try again</Button>
			</div>
		</Alert>
	{/if}

	<!-- Charts belong to the period; actions that recalculate them sit on their rule. -->
	{#if page !== 'admin' && view === 'empty'}
		<section class="border-t border-slate-400 pt-3 pb-2">
			<Heading level="h2" size="md" weight="medium">Net trend</Heading>
			<p class="py-10 text-center text-sm text-slate-500">
				No records in {period.label} yet, so there is nothing to chart.
			</p>
		</section>
	{:else if page === 'cashflow' && !failed}
		<CashflowCharts {loading} />
	{:else if page === 'portfolio' && !failed}
		<PortfolioCharts {loading}>
			{#snippet actions()}
				<Text as="span" size="xs" tone="subtle">{period.label}</Text>
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
			<TableHeader
				title="Transactions"
				meta={tableMeta}
				searchPlaceholder="Search description…"
				bind:searchValue={query}
			>
				{#snippet actions()}
					{#if selectedIds.length > 0}
						<RuledAction label="Tag {selectedIds.length} selected" icon={icons.tag} />
					{/if}
					<RuledAction label="Import CSV" icon={icons.upload} />
					<PrimaryAction label="Add transaction" icon={icons.plus} />
				{/snippet}
			</TableHeader>
			<TransactionsTable
				bind:selectedIds
				{rows}
				{loading}
				error={failed ? 'Could not load transactions.' : null}
				{emptyText}
			>
				{#snippet emptyAction()}
					{#if view === 'no-matches'}
						<Button size="md" variant="outline" shape="default" onclick={() => (query = '')}>
							Clear search
						</Button>
					{:else}
						<PrimaryAction label="Add transaction" icon={icons.plus} />
					{/if}
				{/snippet}
			</TransactionsTable>
		{:else if page === 'portfolio'}
			<!-- The row filter belongs to the positions, so it disappears on the transactions tab. -->
			{#snippet positionStatus()}
				<Tabs
					tabs={[
						{ value: 'open', label: 'Open' },
						{ value: 'closed', label: 'Closed' },
						{ value: 'all', label: 'All' }
					]}
					bind:value={closedFilter}
					size="sm"
					ariaLabel="Position status"
				/>
			{/snippet}
			<TableHeader
				title={tab === 'positions' ? 'Positions' : 'Transactions'}
				meta={loading
					? 'Loading…'
					: tab === 'positions'
						? `${positionRows.length} positions`
						: `${transactionTotal} rows`}
				searchPlaceholder={tab === 'positions'
					? 'Search symbol or name…'
					: 'Search transactions…'}
				bind:searchValue={query}
				filters={tab === 'positions' ? positionStatus : undefined}
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
					<PrimaryAction label="Add transaction" icon={icons.plus} />
				{/snippet}
			</TableHeader>
			<PositionsTable
				rows={positionRows}
				{loading}
				error={failed ? 'Could not load positions.' : null}
				{emptyText}
			/>
		{:else}
			<!-- Admin: a table, its actions, and nothing that belongs to charts. -->
			<TableHeader
				title="Listings"
				meta="{adminListings.length} listings"
				searchPlaceholder="Search listings…"
			>
				{#snippet actions()}
					<RuledAction label="Upload EOD prices" icon={icons.upload} />
					<PrimaryAction label="Add listing" icon={icons.plus} />
				{/snippet}
			</TableHeader>
			<ListingsTable />
		{/if}
	</Panel>
</MockShell>
