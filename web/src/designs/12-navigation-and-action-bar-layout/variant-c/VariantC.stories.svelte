<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import PageC from './PageC.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#12 Navigation and action bar/Variant C: Rules carry the actions',
		component: PageC,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" args={{ page: 'cashflow' }} />
<Story name="Mobile" args={{ page: 'cashflow', overviewOpen: true }} />
<Story name="Overview open" args={{ page: 'cashflow', overviewOpen: true }} />
<Story name="Portfolio" args={{ page: 'portfolio' }} />
