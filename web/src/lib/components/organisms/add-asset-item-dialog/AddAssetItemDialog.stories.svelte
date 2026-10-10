<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';

	import AddAssetItemDialog from './AddAssetItemDialog.svelte';

	import { assetHoldings } from '$lib/data/fixtures/assets';

	const { Story } = defineMeta({
		title: 'Organisms/AddAssetItemDialog',
		component: AddAssetItemDialog,
		tags: ['autodocs'],
		argTypes: {
			open: { control: 'boolean' },
			classId: { control: 'text' },
			className: { control: 'text' }
		}
	});
</script>

<!-- The quote list comes from the mock service branch, so the picker fills itself. -->
<Story name="Playground" args={{ open: true, classId: 'cls-crypto', className: 'Crypto' }} />

<Story name="Add purchase to an existing item" asChild>
	<AddAssetItemDialog
		open
		classId="cls-crypto"
		className="Crypto"
		holding={assetHoldings['ast-btc']}
	/>
</Story>
