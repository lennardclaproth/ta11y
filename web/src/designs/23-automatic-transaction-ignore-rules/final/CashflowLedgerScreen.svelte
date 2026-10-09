<script lang="ts">
	// Final design, screen 3 of 3 — Cashflow, the two touchpoints the feature needs there.
	//
	// 1. "Show ignored rows" puts ignored transactions back in the ledger, each naming the rule
	//    that put it there ("By hand" when nobody's rule did), with Restore on the row.
	// 2. "Make an ignore rule" takes the selected rows to the Ignore rules page with a draft
	//    prefilled — the rule is still written and previewed in the one place that owns rules.
	//
	// Everything else on Cashflow stays as it is: same ledger header, same table, same pager.
	import PreviewShell from '../PreviewShell.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import FooterBar from '$lib/components/organisms/footer-bar/FooterBar.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Switch from '$lib/components/atoms/switch/Switch.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';
	import { ledgerRows, ruleById, type IgnoredTransaction } from '../ignore-rules.fixture';

	type Props = {
		loading?: boolean;
		error?: string | null;
		/** Lay the shell out in document flow (narrow-screen previews). */
		flow?: boolean;
		/** Preselect two rows so the bulk bar is visible. */
		selected?: boolean;
	};

	let { loading = false, error = null, flow = false, selected = true }: Props = $props();

	let showIgnored = $state(true);
	let selectedIds = $state<string[]>([]);

	$effect(() => {
		selectedIds = selected ? ['ig-02', 'ig-03'] : [];
	});

	const rows = $derived(showIgnored ? ledgerRows : ledgerRows.filter((row) => !row.ignored));

	const actions: MenuItem[] = [{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' }];

	function fmtDate(value: string): string {
		return formatDisplayDate(value.slice(0, 10));
	}
</script>

{#snippet dateCell(row: IgnoredTransaction)}
	<span class="whitespace-nowrap tabular-nums">{fmtDate(row.date)}</span>
{/snippet}

{#snippet descriptionCell(row: IgnoredTransaction)}
	<span class={row.ignored ? 'text-slate-500' : 'text-slate-800'}>{row.description}</span>
{/snippet}

<!-- The rule is the reason this row left the totals, so it is named rather than implied.
     A row nobody ignored says so in words too — not by being the only one without a badge. -->
{#snippet ignoredByCell(row: IgnoredTransaction)}
	{#if row.ignored && row.ruleId}
		<Badge intent="info" variant="soft" size="sm">{ruleById(row.ruleId).name}</Badge>
	{:else if row.ignored}
		<Badge intent="neutral" variant="soft" size="sm">By hand</Badge>
	{:else}
		<span class="text-sm text-slate-500">Counted</span>
	{/if}
{/snippet}

{#snippet directionCell(row: IgnoredTransaction)}
	<Badge intent={row.direction === 'in' ? 'success' : 'error'} variant="soft" size="sm">
		{row.direction === 'in' ? 'In' : 'Out'}
	</Badge>
{/snippet}

{#snippet amountCell(row: IgnoredTransaction)}
	<Money amount={scaledToNumber(row.amountCents)} currency="EUR" size="sm" />
{/snippet}

{#snippet rowActionCell(row: IgnoredTransaction)}
	{#if row.ignored}
		<Button size="sm" variant="ghost" intent="secondary" shape="default">Restore</Button>
	{/if}
{/snippet}

<PreviewShell {flow}>
	{#snippet top()}
		<TopNavbar
			title="Cashflow"
			showSearch
			showDateRange
			{actions}
			accountName="Lennard Claproth"
			accountEmail="lennard@example.com"
		/>
	{/snippet}

	<PageContentTemplate>
		<LedgerToolbar title="Transactions" actionLabel="Add transaction" onAdd={() => {}}>
			<label class="flex items-center gap-2 text-sm text-slate-700">
				<Switch bind:checked={showIgnored} />
				Show ignored rows
			</label>
			<Button size="sm" variant="ghost" intent="secondary" shape="default">
				<Icon icon="heroicons:funnel" size="sm" />Ignore rules
			</Button>
		</LedgerToolbar>

		<DataTable
			{rows}
			{loading}
			{error}
			selectable
			bind:selectedIds
			sortKey="date"
			sortDirection="desc"
			emptyText="No transactions match your filters"
			class="min-h-0 flex-1"
			columns={[
				{ key: 'date', header: 'Date', sortKey: 'date', width: 'w-28', cell: dateCell },
				{
					key: 'description',
					header: 'Description',
					sortKey: 'description',
					cell: descriptionCell
				},
				{ key: 'rule', header: 'Ignored by', width: 'w-52', cell: ignoredByCell },
				{ key: 'direction', header: 'Direction', width: 'w-28', cell: directionCell },
				{ key: 'amount', header: 'Amount', sortKey: 'amount', align: 'right', cell: amountCell },
				{ key: 'restore', header: '', align: 'right', width: 'w-28', cell: rowActionCell }
			]}
		>
			{#snippet footer()}
				<FooterBar total={1284} limit={25} offset={0} selectedCount={selectedIds.length}>
					{#snippet actions()}
						<Button size="sm" variant="outline" intent="secondary" shape="default">Restore</Button>
						<Button size="sm" variant="ghost" intent="secondary" shape="default">
							Make an ignore rule
						</Button>
					{/snippet}
				</FooterBar>
			{/snippet}
		</DataTable>
	</PageContentTemplate>
</PreviewShell>
