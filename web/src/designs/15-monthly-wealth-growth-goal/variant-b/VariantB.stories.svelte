<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantBPage from './VariantBPage.svelte';
	import CashflowPage from '../shared/CashflowPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#15 Monthly wealth goal/Variant B: Goal page',
		component: VariantBPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" />

<Story name="Mobile" />

<Story name="Marking transactions" asChild>
	<CashflowPage />
</Story>
