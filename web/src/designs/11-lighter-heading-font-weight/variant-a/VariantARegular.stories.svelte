<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';

	import VariantARegular from './VariantARegular.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#11 Lighter heading weight/Variant A: Regular',
		component: VariantARegular,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" />

<Story name="Mobile" />
