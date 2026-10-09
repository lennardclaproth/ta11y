<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantDPage from './VariantDPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#15 Monthly wealth goal/Variant D: Board tabs',
		component: VariantDPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" args={{ board: 'goal' }} />

<Story name="Mobile" args={{ board: 'goal' }} />

<Story name="Cashflow board" args={{ board: 'cashflow' }} />

<Story name="Empty" args={{ board: 'goal', goalSet: false }} />

<Story name="Setting the goal" args={{ board: 'goal', dialogOpen: true }} />
