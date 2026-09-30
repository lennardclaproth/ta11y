<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantA from './VariantA.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#13 Transaction date/Variant A: Calendar in view',
		component: VariantA,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<VariantA scene="create" />
</Story>

<Story name="Mobile" asChild>
	<VariantA scene="create" />
</Story>

<Story name="Change date" asChild>
	<VariantA scene="edit" />
</Story>
