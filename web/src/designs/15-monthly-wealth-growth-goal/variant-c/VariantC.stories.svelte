<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantCPage from './VariantCPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#15 Monthly wealth goal/Variant C: Cashflow card',
		component: VariantCPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" args={{ drawerOpen: false }} />

<Story name="Mobile" args={{ drawerOpen: false }} />

<Story name="Monthly standing drawer" args={{ drawerOpen: true }} />
