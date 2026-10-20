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
			assetClassName: { control: 'text' }
		}
	});
</script>

<!-- The quote list comes from the mock service branch, so the picker fills itself. -->
<Story name="Playground" args={{ open: true, classId: 'cls-crypto', assetClassName: 'Crypto' }} />

<Story name="Add purchase to an existing item" asChild>
	<AddAssetItemDialog
		open
		classId="cls-crypto"
		assetClassName="Crypto"
		holding={assetHoldings['ast-btc']}
	/>
</Story>

<Story name="Edit a recorded purchase" asChild>
	<AddAssetItemDialog
		open
		classId="cls-crypto"
		assetClassName="Crypto"
		holding={assetHoldings['ast-btc']}
		purchase={assetHoldings['ast-btc'].purchases?.[0] ?? null}
	/>
</Story>
