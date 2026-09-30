<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ImportDrawerDesign from './ImportDrawerDesign.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#10 DeGiro import/Variant B: Zijpaneel met logboek',
		component: ImportDrawerDesign
	});
</script>

<!-- The page stays visible next to the panel. -->
<Story name="Default" args={{ phase: 'form' }} />

<Story
	name="Mobile"
	args={{ phase: 'form' }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>

<!-- Each destination reports as its own line in the log. -->
<Story name="Result" args={{ phase: 'result' }} />

<Story name="Loading" args={{ phase: 'processing' }} />

<Story name="Error" args={{ phase: 'error' }} />
