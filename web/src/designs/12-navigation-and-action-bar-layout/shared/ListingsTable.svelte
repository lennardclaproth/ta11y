<script lang="ts">
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import { adminListings } from './mock-data';
	import type { Listing } from '$lib/api/types';

	/** Stand-in for the admin Listings table: an admin page has a table and no charts. */
	type Props = { rows?: Listing[] };

	let { rows = adminListings }: Props = $props();
</script>

{#snippet typeCell(row: Listing)}
	<Badge intent="neutral" variant="soft" size="sm">{row.type ?? '—'}</Badge>
{/snippet}

<DataTable
	{rows}
	emptyText="No listings"
	class="min-h-0 flex-1"
	columns={[
		{ key: 'symbol', header: 'Symbol', value: (r: Listing) => r.symbol },
		{ key: 'name', header: 'Name', value: (r: Listing) => r.name },
		{ key: 'exchange', header: 'Exchange', value: (r: Listing) => r.exchange ?? '—' },
		{ key: 'currency', header: 'Currency', value: (r: Listing) => r.currency ?? '—' },
		{ key: 'type', header: 'Type', cell: typeCell }
	]}
/>
