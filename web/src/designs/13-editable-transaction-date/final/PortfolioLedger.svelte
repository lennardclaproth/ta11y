<script lang="ts">
	// Final — the Portfolio transaction list, used for one state only: the transaction was saved
	// but a rebuild was already running, so the positions on screen are still the old ones.
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { designPortfolioRows, type DesignPortfolioRow } from '../shared/design-data';
</script>

{#snippet originCell(row: DesignPortfolioRow)}
	<Badge intent={row.origin === 'MANUAL' ? 'info' : 'neutral'} variant="soft" size="sm">
		{row.origin === 'MANUAL' ? 'Manual' : 'Import'}
	</Badge>
{/snippet}

{#snippet amountCell(row: DesignPortfolioRow)}
	<Money amount={row.amount} currency="EUR" size="sm" />
{/snippet}

<div class="flex h-screen flex-col bg-taupe-100 pt-5">
	<PageContentTemplate>
		<Alert intent="warning" title="Portfolio not updated yet">
			<div class="flex flex-wrap items-center justify-between gap-3">
				<span>
					The transaction was saved on 14 Jul 2026, but a rebuild was already running. Positions,
					performance and net worth still show the previous result.
				</span>
				<Button size="sm" variant="outline" intent="secondary" shape="default">
					Rebuild portfolio
				</Button>
			</div>
		</Alert>
		<LedgerToolbar title="Transactions" actionLabel="Add transaction" onAdd={() => {}} />
		<DataTable
			rows={designPortfolioRows}
			sortKey="occurredAt"
			sortDirection="desc"
			emptyText="No transactions yet"
			columns={[
				{
					key: 'occurredAt',
					header: 'Date',
					sortKey: 'occurredAt',
					width: 'w-40',
					value: (r: DesignPortfolioRow) => formatDisplayDate(r.occurredAt)
				},
				{ key: 'listing', header: 'Listing', value: (r: DesignPortfolioRow) => r.listing },
				{ key: 'side', header: 'Side', width: 'w-24', value: (r: DesignPortfolioRow) => r.side },
				{ key: 'origin', header: 'Entered', width: 'w-28', cell: originCell },
				{
					key: 'quantity',
					header: 'Quantity',
					align: 'right',
					width: 'w-28',
					value: (r: DesignPortfolioRow) => r.quantity.toFixed(2)
				},
				{ key: 'amount', header: 'Amount', align: 'right', width: 'w-32', cell: amountCell }
			]}
		/>
	</PageContentTemplate>
</div>
