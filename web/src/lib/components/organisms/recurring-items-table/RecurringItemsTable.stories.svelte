<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RecurringItemsTable from './RecurringItemsTable.svelte';
	import { recurringEnded, recurringExpenses, recurringIncome } from '$lib/data/fixtures/recurring';

	const emptyAccount = 'No recurring items yet. Mark a transaction in Cashflow to start one.';
	const noMatches = 'No recurring items match this search.';

	const { Story } = defineMeta({
		title: 'Organisms/RecurringItemsTable',
		component: RecurringItemsTable,
		tags: ['autodocs'],
		argTypes: {
			group: { control: 'select', options: ['expenses', 'income', 'ended'] }
		}
	});
</script>

<Story
	name="Playground"
	args={{ rows: recurringExpenses, group: 'expenses', emptyText: emptyAccount }}
/>

<Story name="Income" args={{ rows: recurringIncome, group: 'income', emptyText: emptyAccount }} />

<!-- Ended items have no expectation, so the last column reports the month instead. -->
<Story name="Ended" args={{ rows: recurringEnded, group: 'ended', emptyText: emptyAccount }} />

<Story name="Loading" args={{ rows: [], loading: true, emptyText: emptyAccount }} />

<!-- An empty account and a search that found nothing are different situations with
     different ways out, so they never share a sentence. -->
<Story name="Empty account" args={{ rows: [], emptyText: emptyAccount }} />

<Story name="No matches" args={{ rows: [], emptyText: noMatches }} />

<Story
	name="Error"
	args={{
		rows: [],
		error: 'Could not load your recurring items. Try again.',
		emptyText: emptyAccount
	}}
/>
