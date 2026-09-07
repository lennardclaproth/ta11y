<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { ApiError } from '$lib/api/client';
	import ListingForm from './ListingForm.svelte';
	const { Story } = defineMeta({ title: 'Organisms/ListingForm', component: ListingForm });
</script>

<Story name="Create" args={{ onSave: async () => {}, onCancel: () => {} }} />
<Story
	name="Duplicate"
	args={{
		onSave: async () => {
			throw new ApiError(409, 'Duplicate', {});
		},
		onCancel: () => {}
	}}
/>
<Story name="Saving" args={{ saving: true, onSave: async () => {}, onCancel: () => {} }} />
