<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';

	import VariantBMedium from './VariantBMedium.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#11 Lighter heading weight/Variant B: Medium',
		component: VariantBMedium,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" />

<Story name="Mobile" />
