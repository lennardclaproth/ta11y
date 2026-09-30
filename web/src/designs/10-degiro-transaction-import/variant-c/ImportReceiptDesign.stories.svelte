<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ImportReceiptDesign from './ImportReceiptDesign.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#10 DeGiro import/Variant C: Importstrook',
		component: ImportReceiptDesign
	});
</script>

<!-- No overlay: the import opens as a ruled section above the ledger. -->
<Story name="Default" args={{ phase: 'form' }} />

<Story
	name="Mobile"
	args={{ phase: 'form' }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>

<!-- Full page width for the receipt and the products without a listing. -->
<Story name="Result" args={{ phase: 'result' }} />

<Story name="Loading" args={{ phase: 'processing' }} />

<Story name="Error" args={{ phase: 'error' }} />
