<script lang="ts">
	// Variant C — "Cashflow card": no new destination and no new tab. The goal lives in the
	// analytics band where you already work, and the month history opens in a drawer.
	import CashflowPage from '../shared/CashflowPage.svelte';
	import WealthGoalCard from './WealthGoalCard.svelte';
	import MonthStandingDrawer from './MonthStandingDrawer.svelte';

	type Props = { drawerOpen?: boolean };

	let { drawerOpen = $bindable(false) }: Props = $props();
</script>

<CashflowPage>
	{#snippet goalCard()}
		<WealthGoalCard onOpen={() => (drawerOpen = true)} />
	{/snippet}

	{#snippet overlay()}
		<MonthStandingDrawer bind:open={drawerOpen} />
	{/snippet}
</CashflowPage>
