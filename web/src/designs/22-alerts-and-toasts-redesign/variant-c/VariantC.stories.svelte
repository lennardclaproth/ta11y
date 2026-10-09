<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantC from './VariantC.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#22 Alerts and toasts/Variant C: Masthead band',
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
