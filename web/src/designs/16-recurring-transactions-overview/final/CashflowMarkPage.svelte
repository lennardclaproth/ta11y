<script lang="ts">
	// Where a recurring item starts: Cashflow, with rows selected. A trimmed stand-in for the real
	// page — it exists to place the "Mark as recurring" action next to the existing bulk action,
	// not to redesign Cashflow. Analytics, filters and the tag flow stay exactly as they are.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import FooterBar from '$lib/components/organisms/footer-bar/FooterBar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import MarkRecurringDialog from './MarkRecurringDialog.svelte';
	import { cashflowRows, type CashflowRow } from '../recurring.fixture';

	let { dialogOpen = $bindable(true) }: { dialogOpen?: boolean } = $props();

	let selectedIds = $state(['cf-1']);
	const selection = $derived(cashflowRows.filter((row) => selectedIds.includes(row.id)));
</script>

{#snippet dateCell(row: CashflowRow)}
	<span class="whitespace-nowrap">{formatDisplayDate(row.date)}</span>
{/snippet}

{#snippet tagCell(row: CashflowRow)}
	<Badge intent="neutral" variant="soft" size="sm">{row.tag}</Badge>
{/snippet}

{#snippet directionCell(row: CashflowRow)}
	<Badge intent={row.direction === 'in' ? 'success' : 'error'} variant="soft" size="sm">
		{row.direction === 'in' ? 'In' : 'Out'}
	</Badge>
{/snippet}

{#snippet amountCell(row: CashflowRow)}
	<Money amount={row.amount} currency="EUR" size="sm" />
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Cashflow"
			showSearch
			searchPlaceholder="Search description…"
			accountName="Account"
			accountEmail="you@example.com"
		/>
	{/snippet}

	<PageContentTemplate>
		<LedgerToolbar title="Transactions" actionLabel="Add transaction" onAdd={() => {}} />
		<DataTable
			rows={cashflowRows}
			bind:selectedIds
			selectable
			sortKey="date"
			sortDirection="desc"
			columns={[
				{ key: 'date', header: 'Date', sortKey: 'date', width: 'w-32', cell: dateCell },
				{ key: 'description', header: 'Description', sortKey: 'description' , value: (r: CashflowRow) => r.description },
				{ key: 'tag', header: 'Tag', sortKey: 'tag', width: 'w-40', cell: tagCell },
				{ key: 'direction', header: 'Direction', width: 'w-28', cell: directionCell },
				{ key: 'amount', header: 'Amount', sortKey: 'amount', align: 'right', width: 'w-32', cell: amountCell }
			]}
		>
			{#snippet footer()}
				<FooterBar total={cashflowRows.length} selectedCount={selectedIds.length}>
					{#snippet actions()}
						<Button size="sm" variant="ghost" intent="secondary" shape="default">Tag</Button>
						<Button size="sm" variant="outline" shape="default" onclick={() => (dialogOpen = true)}>
							Mark as recurring
						</Button>
					{/snippet}
				</FooterBar>
			{/snippet}
		</DataTable>
	</PageContentTemplate>
</AppShellTemplate>

<MarkRecurringDialog bind:open={dialogOpen} {selection} />
