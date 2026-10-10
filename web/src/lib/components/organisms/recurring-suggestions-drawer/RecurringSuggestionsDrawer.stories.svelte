<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RecurringSuggestionsDrawer from './RecurringSuggestionsDrawer.svelte';
	import { recurringSuggestions } from '$lib/data/fixtures/recurring';

	const { Story } = defineMeta({
		title: 'Organisms/RecurringSuggestionsDrawer',
		component: RecurringSuggestionsDrawer,
		tags: ['autodocs']
	});
</script>

<Story name="Playground" args={{ open: true, suggestions: recurringSuggestions }} />

<Story name="Loading" args={{ open: true, suggestions: [], loading: true }} />

<!-- Nothing found is not a failure: the way out is to mark one yourself. -->
<Story name="Nothing found" args={{ open: true, suggestions: [] }} />

<Story
	name="Error"
	args={{
		open: true,
		suggestions: [],
		error: 'Could not read your history right now. Try again.'
	}}
/>

<!-- While one suggestion is being confirmed, its card's actions are held. -->
<Story
	name="Confirming"
	args={{
		open: true,
		suggestions: recurringSuggestions,
		pending: recurringSuggestions[0].match_key
	}}
/>
