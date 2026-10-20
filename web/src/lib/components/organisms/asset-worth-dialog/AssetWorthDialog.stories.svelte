<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';

	import AssetWorthDialog from './AssetWorthDialog.svelte';

	import { assetClassDetails } from '$lib/data/fixtures/assets';

	const savings = { id: 'ast-ing-savings', name: 'ING savings', current_worth: '12450.000000' };

	const { Story } = defineMeta({
		title: 'Organisms/AssetWorthDialog',
		component: AssetWorthDialog,
		tags: ['autodocs'],
		argTypes: {
			open: { control: 'boolean' },
			opening: { control: 'boolean' }
		}
	});
</script>

<!-- Saving goes through the mock service branch, so the dialog closes as it would live. -->
<Story name="Playground" args={{ open: true, asset: savings }} />

<Story
	name="Edit a recorded change"
	args={{ open: true, asset: savings, editing: assetClassDetails['cls-cash'].mutations[0] }}
/>
