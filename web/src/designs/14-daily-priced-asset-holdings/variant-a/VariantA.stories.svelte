<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantA from './VariantA.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#14 Daily-priced asset holdings/Variant A: Expandable drawer',
		component: VariantA,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<VariantA />
</Story>

<Story name="Mobile" asChild>
	<VariantA />
</Story>

<Story name="Add item" asChild>
	<VariantA addOpen />
</Story>

<Story name="Empty class" asChild>
	<VariantA empty expandedId={null} />
</Story>
