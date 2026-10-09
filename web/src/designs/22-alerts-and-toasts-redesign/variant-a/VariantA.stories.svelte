<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantA from './VariantA.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#22 Alerts and toasts/Variant A: Ruled notice',
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
