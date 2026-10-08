<script lang="ts">
	// The Cashflow page as it is today, plus the marking surface every variant needs: a Purpose
	// column, a Purpose header filter, and bulk actions in the selection footer.
	import type { Snippet } from 'svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import CashflowAnalytics from './CashflowAnalytics.svelte';
	import PurposeTransactionsTable from './PurposeTransactionsTable.svelte';
	import { goalTransactions } from '../goal-data';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	type Props = {
		/** An extra analytics card appended to the existing band. */
		goalCard?: Snippet;
		/** Overlays rendered above the page (e.g. a drawer). */
		overlay?: Snippet;
	};

	let { goalCard, overlay }: Props = $props();

	let selectedIds = $state<string[]>(['px-04', 'px-05']);

	const navActions: MenuItem[] = [{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' }];
</script>

<!-- The shell is viewport-tall in the app; in the prototype it grows with its content so a
     screenshot shows the whole screen instead of the first fold. -->
<AppShellTemplate class="h-auto! min-h-dvh">
	{#snippet top()}
		<TopNavbar
			title="Cashflow"
			showSearch
			showDateRange
			actions={navActions}
			accountName="Design preview"
			accountEmail="preview@example.com"
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			<CashflowAnalytics extra={goalCard} />
		{/snippet}

		<LedgerToolbar title="Transactions" actionLabel="Add transaction" onAdd={() => {}}>
			<Button size="sm" variant="ghost" intent="secondary">Mark all 24 matches…</Button>
		</LedgerToolbar>

		<PurposeTransactionsTable rows={goalTransactions} bind:selectedIds />
	</PageContentTemplate>
</AppShellTemplate>

{@render overlay?.()}
