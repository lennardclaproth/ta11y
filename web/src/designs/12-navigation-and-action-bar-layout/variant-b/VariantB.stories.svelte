<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import PageB from './PageB.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#12 Navigation and action bar/Variant B: One entry, search first',
		component: PageB,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" args={{ page: 'cashflow' }} />
<Story name="Mobile" args={{ page: 'cashflow', overviewOpen: true }} />
<Story name="Overview open" args={{ page: 'cashflow', overviewOpen: true }} />
<Story name="Portfolio" args={{ page: 'portfolio' }} />
