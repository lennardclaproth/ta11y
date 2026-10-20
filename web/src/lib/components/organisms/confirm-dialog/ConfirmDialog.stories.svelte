<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';

	import ConfirmDialog from './ConfirmDialog.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/ConfirmDialog',
		component: ConfirmDialog,
		tags: ['autodocs'],
		argTypes: {
			open: { control: 'boolean' },
			title: { control: 'text' },
			message: { control: 'text' },
			confirmLabel: { control: 'text' }
		}
	});
</script>

<Story
	name="Playground"
	args={{
		open: true,
		title: 'Delete ING savings?',
		message: 'ING savings and its history are removed. This cannot be undone.',
		onConfirm: async () => {}
	}}
/>

<Story
	name="Refused"
	args={{
		open: true,
		title: 'Delete this change?',
		message: 'The change is removed and the history after it recalculated.',
		onConfirm: async () => {
			throw Object.assign(new Error('conflict'), {
				body: { error: 'an item keeps at least one worth change; delete the item instead' }
			});
		}
	}}
/>
