<script lang="ts">
	// Prototype only: the real table is organisms/cashflow-transactions-table. This copy adds the
	// "Purpose" column, its header filter and the bulk marking actions so the proposal can be read.
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import FooterBar from '$lib/components/organisms/footer-bar/FooterBar.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import ActionMenu from '$lib/components/molecules/action-menu/ActionMenu.svelte';
	import SelectFilter from '$lib/components/molecules/select-filter/SelectFilter.svelte';
	import { purposeLabel, type GoalTransaction } from '../goal-data';

	type Props = {
		rows: GoalTransaction[];
		selectedIds?: string[];
		/** Bindable Purpose filter, so a story can open on a filtered ledger. */
		purposeFilter?: string[];
		total?: number;
		loading?: boolean;
		error?: string | null;
		emptyText?: string;
		class?: string;
	};

	let {
		rows,
		selectedIds = $bindable([]),
		purposeFilter = $bindable([]),
		total,
		loading = false,
		error = null,
		emptyText = 'No transactions match your filters',
		class: className = ''
	}: Props = $props();

	const purposeOptions = [
		{ value: 'income', label: 'Income' },
		{ value: 'wealth', label: 'To wealth' },
		{ value: 'none', label: 'Not assigned' }
	];

	// The bulk actions do not fit beside "n selected" on a 390px footer, so below `sm` the same
	// actions live in one menu instead of being hidden or wrapped off-screen.
	const bulkItems = [
		{ label: 'Mark as income' },
		{ label: 'Mark as wealth' },
		{ label: 'Clear purpose' },
		{ label: 'Tag', divider: true }
	];
</script>

{#snippet purposeFilterControl()}
	<SelectFilter options={purposeOptions} bind:selected={purposeFilter} label="Purpose" />
{/snippet}

{#snippet tagCell(row: GoalTransaction)}
	<Badge intent="neutral" variant="soft" size="sm">{row.tag}</Badge>
{/snippet}

{#snippet directionCell(row: GoalTransaction)}
	<Badge intent={row.direction === 'in' ? 'success' : 'error'} variant="soft" size="sm">
		{row.direction === 'in' ? 'In' : 'Out'}
	</Badge>
{/snippet}

<!-- Outline rather than soft, so the purpose never reads as a second direction badge. -->
{#snippet purposeCell(row: GoalTransaction)}
	{#if row.purpose === 'income'}
		<Badge intent="success" variant="outline" size="sm">{purposeLabel.income}</Badge>
	{:else if row.purpose === 'wealth'}
		<Badge intent="info" variant="outline" size="sm">{purposeLabel.wealth}</Badge>
	{:else}
		<span class="text-slate-500">Not assigned</span>
	{/if}
{/snippet}

{#snippet amountCell(row: GoalTransaction)}
	<Money amount={row.amount} currency="EUR" size="sm" />
{/snippet}

<DataTable
	{rows}
	{loading}
	{error}
	{emptyText}
	selectable
	bind:selectedIds
	sortKey="date"
	sortDirection="desc"
	class={className}
	columns={[
		{ key: 'date', header: 'Date', sortKey: 'date', value: (r: GoalTransaction) => r.date },
		{
			key: 'description',
			header: 'Description',
			sortKey: 'description',
			value: (r: GoalTransaction) => r.description
		},
		{ key: 'tag', header: 'Tag', sortKey: 'tag', cell: tagCell },
		{ key: 'direction', header: 'Direction', cell: directionCell },
		{
			key: 'purpose',
			header: 'Purpose',
			width: 'w-40',
			cell: purposeCell,
			filter: purposeFilterControl
		},
		{ key: 'amount', header: 'Amount', sortKey: 'amount', align: 'right', cell: amountCell }
	]}
>
	{#snippet footer()}
		<FooterBar
			total={total ?? rows.length}
			limit={25}
			offset={0}
			selectedCount={selectedIds.length}
		>
			{#snippet actions()}
				<!-- The wrapper carries the breakpoint: `hidden` on the Button itself loses to the
				     atom's own `inline-flex`, which Tailwind emits later in the sheet. -->
				<span class="hidden items-center gap-2 sm:flex">
					<Button size="sm" variant="ghost" intent="secondary">Tag</Button>
					<Button size="sm" variant="outline" intent="success">Mark as income</Button>
					<Button size="sm" variant="outline" intent="info">Mark as wealth</Button>
				</span>
				<span class="sm:hidden">
					<ActionMenu items={bulkItems} label="Mark…" placement="top-start" />
				</span>
			{/snippet}
		</FooterBar>
	{/snippet}
</DataTable>
