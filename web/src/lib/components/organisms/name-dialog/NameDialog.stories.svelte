<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';

	import NameDialog from './NameDialog.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/NameDialog',
		component: NameDialog,
		tags: ['autodocs'],
		argTypes: {
			open: { control: 'boolean' },
			title: { control: 'text' },
			label: { control: 'text' },
			value: { control: 'text' }
		}
	});
</script>

<Story
	name="Playground"
	args={{ open: true, title: 'Rename class', value: 'Savings', onSave: async () => {} }}
/>

<Story
	name="Refused"
	args={{
		open: true,
		title: 'Rename class',
		value: 'Savings',
		onSave: async () => {
			throw Object.assign(new Error('conflict'), {
				body: { name: 'a class with this name already exists' }
			});
		}
	}}
/>
