<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantB from './VariantB.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#13 Transaction date/Variant B: Field and drawer',
		component: VariantB,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<VariantB scene="create" />
</Story>

<Story name="Mobile" asChild>
	<VariantB scene="create" />
</Story>

<Story name="Change date" asChild>
	<VariantB scene="edit" />
</Story>
