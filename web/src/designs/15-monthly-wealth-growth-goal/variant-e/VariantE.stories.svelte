<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantEPage from './VariantEPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#15 Monthly wealth goal/Variant E: Card rail',
		component: VariantEPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" args={{ startAt: 'month' }} />

<Story name="Mobile" args={{ startAt: 'month' }} />

<Story name="Cashflow cards" args={{ startAt: 'trend' }} />

<Story name="Empty" args={{ startAt: 'goal', goalSet: false }} />
