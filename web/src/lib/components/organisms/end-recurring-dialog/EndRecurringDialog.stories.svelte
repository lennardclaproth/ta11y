<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import EndRecurringDialog from './EndRecurringDialog.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/EndRecurringDialog',
		component: EndRecurringDialog,
		tags: ['autodocs']
	});
</script>

<Story name="Playground" args={{ open: true, name: 'Pixel Stream' }} />

<Story name="Saving" args={{ open: true, name: 'Pixel Stream', saving: true }} />

<!-- A refused month is reported on the field, with the chosen month still selected. -->
<Story
	name="Error"
	args={{ open: true, name: 'Pixel Stream', error: 'Could not end this item. Try again.' }}
/>
