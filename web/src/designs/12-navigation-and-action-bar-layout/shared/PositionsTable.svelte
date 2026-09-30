<script lang="ts">
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { positions } from './mock-data';
	import type { PortfolioPosition } from '$lib/api/types';
</script>

{#snippet marketValueCell(row: PortfolioPosition)}
	{#if row.market_value !== null && row.market_value !== undefined}
		<Money amount={scaledToNumber(row.market_value)} currency="EUR" size="sm" />
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

{#snippet pnlCell(row: PortfolioPosition)}
	{#if row.unrealized_pnl_pct !== null && row.unrealized_pnl_pct !== undefined}
		<Badge intent={row.unrealized_pnl_pct >= 0 ? 'success' : 'error'} variant="soft" size="sm">
			{row.unrealized_pnl_pct >= 0 ? '+' : ''}{row.unrealized_pnl_pct.toFixed(2)}%
		</Badge>
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

<DataTable
	rows={positions}
	emptyText="No positions"
	class="min-h-0 flex-1"
	columns={[
		{ key: 'symbol', header: 'Symbol', value: (r: PortfolioPosition) => r.symbol ?? '—' },
		{ key: 'name', header: 'Name', value: (r: PortfolioPosition) => r.name ?? '—' },
		{ key: 'quantity', header: 'Qty', align: 'right', value: (r: PortfolioPosition) => r.quantity },
		{ key: 'market_value', header: 'Market value', align: 'right', cell: marketValueCell },
		{ key: 'pnl', header: 'Unrealized', align: 'right', cell: pnlCell }
	]}
/>
