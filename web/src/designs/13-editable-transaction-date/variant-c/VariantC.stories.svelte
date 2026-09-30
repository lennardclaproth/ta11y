<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantC from './VariantC.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#13 Transaction date/Variant C: In the ledger',
		component: VariantC,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<VariantC scene="create" />
</Story>

<Story name="Mobile" asChild>
	<VariantC scene="create" />
</Story>

<Story name="Change date" asChild>
	<VariantC scene="edit" />
</Story>
