<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantB from './VariantB.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#14 Daily-priced asset holdings/Variant B: Class page',
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

<Story name="Add item — pick what you own" asChild>
	<VariantB addOpen addStep="quote" />
</Story>

<Story name="Add item — first purchase" asChild>
	<VariantB addOpen addStep="purchase" />
</Story>

<Story name="Loading" asChild>
	<VariantB loading />
</Story>

<Story name="Stale prices" asChild>
	<VariantB stale />
</Story>
