<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ImportDialogDesign from './ImportDialogDesign.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#10 DeGiro import/Variant A: Dialoog in twee stappen',
		component: ImportDialogDesign
	});
</script>

<!-- Step 1: pick the account and the export. -->
<Story name="Default" args={{ phase: 'form', entry: 'Portfolio' }} />

<Story
	name="Mobile"
	args={{ phase: 'form', entry: 'Portfolio' }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>

<!-- Step 2: what landed where, per destination. -->
<Story name="Result" args={{ phase: 'result' }} />

<Story name="Loading" args={{ phase: 'processing' }} />

<!-- The refusal that matters: another DEGIRO export than the account statement. -->
<Story name="Error" args={{ phase: 'error' }} />
