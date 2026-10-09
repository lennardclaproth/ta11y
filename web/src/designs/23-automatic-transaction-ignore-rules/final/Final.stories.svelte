<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import IgnoreRulesScreen from './IgnoreRulesScreen.svelte';
	import ImportReviewScreen from './ImportReviewScreen.svelte';
	import CashflowLedgerScreen from './CashflowLedgerScreen.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#23 Ignore rules/Final',
		component: IgnoreRulesScreen
	});
</script>

<Story name="Default" asChild>
	<IgnoreRulesScreen />
</Story>

<Story name="Mobile" asChild>
	<IgnoreRulesScreen mobileView="detail" flow />
</Story>

<Story name="Loading" asChild>
	<IgnoreRulesScreen loading />
</Story>

<Story name="Empty" asChild>
	<IgnoreRulesScreen empty />
</Story>

<Story name="Error" asChild>
	<IgnoreRulesScreen error="The server did not answer." />
</Story>

<Story name="No matches" asChild>
	<IgnoreRulesScreen noMatches />
</Story>

<Story name="Apply to existing" asChild>
	<IgnoreRulesScreen applyOpen />
</Story>

<Story name="Import review" asChild>
	<ImportReviewScreen />
</Story>

<Story name="Import review Mobile" asChild>
	<ImportReviewScreen flow />
</Story>

<Story name="Import still running" asChild>
	<ImportReviewScreen loading />
</Story>

<Story name="Import review, nothing ignored" asChild>
	<ImportReviewScreen empty />
</Story>

<Story name="Cashflow with ignored rows" asChild>
	<CashflowLedgerScreen />
</Story>
