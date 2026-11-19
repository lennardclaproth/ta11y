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

<!-- Confirming leaves this drawer open so the rest of the list can be worked through, and the
     drawer covers the notice region, so the outcome is stated here. -->
<Story
	name="Confirmed"
	args={{
		open: true,
		suggestions: recurringSuggestions.slice(1),
		notice: {
			intent: 'success',
			message: `${recurringSuggestions[0].name} added, with ${recurringSuggestions[0].matches} transactions`
		}
	}}
/>
