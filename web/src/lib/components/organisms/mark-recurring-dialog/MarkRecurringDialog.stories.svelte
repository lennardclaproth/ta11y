<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import MarkRecurringDialog from './MarkRecurringDialog.svelte';
	import { cashflowTransactions } from '$lib/data/fixtures/cashflow';
	import { recurringExpenses, recurringIncome } from '$lib/data/fixtures/recurring';

	const selection = cashflowTransactions.slice(0, 3);
	const items = [...recurringExpenses, ...recurringIncome];

	const { Story } = defineMeta({
		title: 'Organisms/MarkRecurringDialog',
		component: MarkRecurringDialog,
		tags: ['autodocs']
	});
</script>

<Story name="Playground" args={{ open: true, selection, items }} />

<!-- One row, picked from the table: the common case. -->
<Story name="One transaction" args={{ open: true, selection: selection.slice(0, 1), items }} />

<!-- An account with no items yet has only one route, so the dialog opens on it. -->
<Story name="First item" args={{ open: true, selection, items: [] }} />

<Story name="Saving" args={{ open: true, selection, items, saving: true }} />

<Story
	name="Error"
	args={{
		open: true,
		selection,
		items,
		error: 'A recurring item with that name already exists.'
	}}
/>
