<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantC from './VariantC.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#14 Daily-priced asset holdings/Variant C: Holdings ledger',
		component: VariantC,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<VariantC />
</Story>

<Story name="Mobile" asChild>
	<VariantC />
</Story>

<Story name="Add holding" asChild>
	<VariantC addOpen />
</Story>

<Story name="Nothing selected" asChild>
	<VariantC selectedId={null} />
</Story>

<Story name="Empty" asChild>
	<VariantC empty selectedId={null} />
</Story>
