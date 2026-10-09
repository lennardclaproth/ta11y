<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantB from './VariantB.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#22 Alerts and toasts/Variant B: Ledger entry',
		component: VariantB,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<VariantB />
</Story>

<Story name="Mobile" asChild>
	<VariantB />
</Story>
