<script lang="ts">
	// Variant A — "Ledger lane". Everything the feature needs stays on Cashflow: the ledger
	// header gains view tabs (All / Ignored / Rules), the last import reports itself as a ruled
	// strip above the records, and an ignored row names the rule that put it there.
	import PreviewShell from '../PreviewShell.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import FooterBar from '$lib/components/organisms/footer-bar/FooterBar.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';
	import {
		ignoredTransactions,
		lastImport,
		ruleById,
		type IgnoredTransaction
	} from '../ignore-rules.fixture';

	type Props = {
		/** Start with no ignored rows, to show the state before any rule has fired. */
		empty?: boolean;
		loading?: boolean;
		error?: string | null;
		/** Lay the shell out in document flow (narrow-screen previews). */
		flow?: boolean;
		/**
		 * Preselect two rows so the bulk bar is visible. Off on narrow previews: FooterBar does
		 * not wrap, so its bulk actions collide with the pager below `sm`.
		 */
		selected?: boolean;
	};

	let {
		empty = false,
		loading = false,
		error = null,
		flow = false,
		selected = true
	}: Props = $props();

	let view = $state('ignored');
	let selectedIds = $state<string[]>([]);

	$effect(() => {
		selectedIds = selected ? ['ig-02', 'ig-06'] : [];
	});

	const rows = $derived(empty ? [] : ignoredTransactions);

	const tabs = [
		{ value: 'all', label: 'All · 1,284' },
		{ value: 'ignored', label: 'Ignored · 312' },
		{ value: 'rules', label: 'Rules · 5' }
	];

	const actions: MenuItem[] = [{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' }];

	const counts = [
		{ label: 'New', value: lastImport.imported },
		{ label: 'Duplicates', value: lastImport.duplicates },
		{ label: 'Auto-ignored', value: lastImport.autoIgnored }
	];

	function fmtDate(value: string): string {
		return formatDisplayDate(value.slice(0, 10));
	}
</script>

{#snippet dateCell(row: IgnoredTransaction)}
	<span class="whitespace-nowrap tabular-nums">{fmtDate(row.date)}</span>
{/snippet}

{#snippet descriptionCell(row: IgnoredTransaction)}
	<span class="text-slate-800">{row.description}</span>
{/snippet}

<!-- The rule is the reason this row left the totals, so it is named rather than implied. -->
{#snippet ignoredByCell(row: IgnoredTransaction)}
	{#if row.ruleId}
		<Badge intent="info" variant="soft" size="sm">{ruleById(row.ruleId).name}</Badge>
	{:else}
		<Badge intent="neutral" variant="soft" size="sm">By hand</Badge>
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

{#snippet restoreCell()}
	<Button size="sm" variant="ghost" intent="secondary" shape="default">Restore</Button>
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
			<Tabs {tabs} bind:value={view} size="sm" ariaLabel="Ledger view" />
		</LedgerToolbar>

		<!-- The import reports itself where the records are, not in a dialog that is already gone. -->
		<section
			class="flex shrink-0 flex-wrap items-center justify-between gap-x-6 gap-y-2 border-b border-slate-200 bg-white px-4 py-2"
			aria-label="Last import"
		>
			<div class="min-w-0">
				<Text as="p" size="sm" class="text-slate-900">
					Last import · {lastImport.file} · {lastImport.finishedAgo}
				</Text>
				<dl class="flex flex-wrap items-baseline gap-x-4 gap-y-0.5">
					{#each counts as count (count.label)}
						<div class="flex items-baseline gap-1.5">
							<dt class="text-sm text-slate-500">{count.label}</dt>
							<dd class="text-sm text-slate-900 tabular-nums">{count.value}</dd>
						</div>
					{/each}
				</dl>
			</div>

			<Button size="sm" variant="outline" intent="secondary" shape="default">
				Review auto-ignored
			</Button>
		</section>

		<DataTable
			rows={rows}
			{loading}
			{error}
			selectable
			bind:selectedIds
			sortKey="date"
			sortDirection="desc"
			emptyText="No ignored transactions. Rules put them here as imports come in."
			class="min-h-0 flex-1"
			columns={[
				{ key: 'date', header: 'Date', sortKey: 'date', width: 'w-28', cell: dateCell },
				{
					key: 'description',
					header: 'Description',
					sortKey: 'description',
					cell: descriptionCell
				},
				{ key: 'rule', header: 'Ignored by', width: 'w-56', cell: ignoredByCell },
				{ key: 'direction', header: 'Direction', width: 'w-28', cell: directionCell },
				{ key: 'amount', header: 'Amount', sortKey: 'amount', align: 'right', cell: amountCell },
				{ key: 'restore', header: '', align: 'right', width: 'w-28', cell: restoreCell }
			]}
		>
			{#snippet footer()}
				<FooterBar total={312} limit={25} offset={0} selectedCount={selectedIds.length}>
					{#snippet actions()}
						<Button size="sm" variant="outline" intent="secondary" shape="default">
							Restore selected
						</Button>
						<Button size="sm" variant="ghost" intent="secondary" shape="default">
							Make a rule from these
						</Button>
					{/snippet}
				</FooterBar>
			{/snippet}
		</DataTable>
	</PageContentTemplate>
</PreviewShell>
