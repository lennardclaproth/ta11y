<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RecurringItemDrawer from './RecurringItemDrawer.svelte';
	import {
		recurringEnded,
		recurringExpenses,
		recurringLinkedTransactions
	} from '$lib/data/fixtures/recurring';

	// The amount of this item grew from €9 to €12 without any announcement — reading that
	// is exactly what the chart is for.
	const opened = {
		...recurringExpenses[5],
		transactions: recurringLinkedTransactions['rc-006']
	};
	const ended = { ...recurringEnded[0], transactions: [] };

	const { Story } = defineMeta({
		title: 'Organisms/RecurringItemDrawer',
		component: RecurringItemDrawer,
		tags: ['autodocs']
	});
</script>

<!-- The two rows dated 18 July are the same payment twice: an overlapping import can do
     that, and Unlink is how it is corrected. -->
<Story name="Playground" args={{ open: true, item: opened }} />

<Story name="Loading" args={{ open: true, item: null, loading: true }} />

<Story
	name="Error"
	args={{ open: true, item: null, error: 'Could not load this recurring item. Try again.' }}
/>

<!-- An ended item keeps its history and loses the way to end it again. -->
<Story name="Ended" args={{ open: true, item: ended }} />

<Story
	name="Unlinking"
	args={{ open: true, item: opened, unlinking: recurringLinkedTransactions['rc-006'][3].id }}
/>
