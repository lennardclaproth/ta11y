<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import PageFinal from './PageFinal.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#12 Navigation and action bar/Final',
		component: PageFinal,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" args={{ page: 'cashflow' }} />
<Story name="Mobile" args={{ page: 'cashflow', overviewOpen: true }} />
<Story name="Loading" args={{ page: 'cashflow', view: 'loading' }} />
<Story name="Empty" args={{ page: 'cashflow', view: 'empty' }} />
<Story name="Error" args={{ page: 'cashflow', view: 'error' }} />
<Story name="No matches" args={{ page: 'cashflow', view: 'no-matches' }} />
<Story name="Period picker" args={{ page: 'cashflow', periodOpen: true }} />
<Story name="Portfolio" args={{ page: 'portfolio' }} />
<Story name="Admin" args={{ page: 'admin' }} />
