<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import PageA from './PageA.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#12 Navigation and action bar/Variant A: Ledger masthead',
		component: PageA,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" args={{ page: 'cashflow' }} />
<Story name="Mobile" args={{ page: 'cashflow', overviewOpen: true }} />
<Story name="Portfolio" args={{ page: 'portfolio' }} />
