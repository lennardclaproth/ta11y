<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ImportDialogDesign from './ImportDialogDesign.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#10 DeGiro import/Final',
		component: ImportDialogDesign
	});

	const mobile = { viewport: { defaultViewport: 'mobile1' } };
</script>

<!-- Step 1: account, export, and the trail showing what the one upload will do. -->
<Story name="Default" args={{ phase: 'form', direction: 'horizontal' }} />

<Story name="Mobile" args={{ phase: 'form', direction: 'horizontal' }} parameters={mobile} />

<!-- The same dialog with the trail running vertical. -->
<Story name="Vertical trail" args={{ phase: 'form', direction: 'vertical' }} />

<Story
	name="Vertical trail Mobile"
	args={{ phase: 'form', direction: 'vertical' }}
	parameters={mobile}
/>

<!-- Processing: the trail replaces the progress bar; Cashflow is done, Portfolio is running. -->
<Story name="Loading" args={{ phase: 'processing', direction: 'horizontal' }} />

<Story name="Loading vertical trail" args={{ phase: 'processing', direction: 'vertical' }} />

<!-- Step 2: what landed where, per destination. -->
<Story name="Result" args={{ phase: 'result', direction: 'horizontal' }} />

<Story name="Result vertical trail" args={{ phase: 'result', direction: 'vertical' }} />

<Story name="Result Mobile" args={{ phase: 'result', direction: 'horizontal' }} parameters={mobile} />

<!-- Empty for this flow: a monthly export that overlapped completely. -->
<Story name="Empty" args={{ phase: 'empty', direction: 'horizontal' }} />

<!-- The refusal that matters: another DEGIRO export than the account statement. -->
<Story name="Error" args={{ phase: 'error', direction: 'horizontal' }} />
