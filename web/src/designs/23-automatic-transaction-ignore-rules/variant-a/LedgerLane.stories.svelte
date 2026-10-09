<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import LedgerLaneScreen from './LedgerLaneScreen.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#23 Ignore rules/Variant A: Ledger lane',
		component: LedgerLaneScreen
	});
</script>

<Story name="Default" asChild>
	<LedgerLaneScreen />
</Story>

<Story name="Mobile" asChild>
	<LedgerLaneScreen flow selected={false} />
</Story>
