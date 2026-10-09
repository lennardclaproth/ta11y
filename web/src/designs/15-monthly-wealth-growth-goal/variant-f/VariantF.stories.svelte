<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VariantFPage from './VariantFPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#15 Monthly wealth goal/Variant F: Section stack',
		component: VariantFPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" args={{ startAt: 'goal' }} />

<Story name="Mobile" args={{ startAt: 'goal' }} />

<Story name="Cashflow section" args={{ startAt: 'overview' }} />

<Story name="Empty" args={{ startAt: 'goal', goalSet: false }} />
