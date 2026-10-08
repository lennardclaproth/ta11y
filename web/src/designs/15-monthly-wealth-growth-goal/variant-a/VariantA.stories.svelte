<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantAPage from './VariantAPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#15 Monthly wealth goal/Variant A: Ledger tab',
		component: VariantAPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" args={{ tab: 'goal' }} />

<Story name="Mobile" args={{ tab: 'goal' }} />

<Story name="Marking transactions" args={{ tab: 'transactions' }} />
