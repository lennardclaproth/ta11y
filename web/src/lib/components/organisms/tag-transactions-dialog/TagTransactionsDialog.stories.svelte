<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import TagTransactionsDialog from './TagTransactionsDialog.svelte';
	import { cashflowTransactions } from '$lib/data/fixtures/cashflow';
	import type { CashflowTag } from '$lib/api/types';

	// A mix of tagged and untagged rows, so the preview shows rows that change and rows that do not.
	const selection = cashflowTransactions.slice(0, 4);
	const tags: CashflowTag[] = [
		{ tag: 'groceries', count: 42 },
		{ tag: 'transport', count: 18 },
		{ tag: 'dining', count: 15 },
		{ tag: 'rent', count: 12 },
		{ tag: 'entertainment', count: 11 },
		{ tag: 'utilities', count: 9 },
		{ tag: 'subscriptions', count: 7 },
		{ tag: 'salary', count: 6 },
		{ tag: 'health', count: 4 }
	];

	const { Story } = defineMeta({
		title: 'Organisms/TagTransactionsDialog',
		component: TagTransactionsDialog,
		tags: ['autodocs']
	});
</script>

<Story name="Playground" args={{ open: true, selection, tags }} />

<!-- One row, opened from a selection of one. -->
<Story name="One transaction" args={{ open: true, selection: selection.slice(0, 1), tags }} />

<!-- An account that has tagged nothing yet: the list only offers to start a tag. -->
<Story name="No tags yet" args={{ open: true, selection, tags: [] }} />

<Story name="Loading tags" args={{ open: true, selection, tags: [], tagsLoading: true }} />

<!-- Select-all across pages: the preview stops at 50 rows and says how many follow. -->
<Story name="Many transactions" args={{ open: true, selection: cashflowTransactions, tags }} />

<Story name="Saving" args={{ open: true, selection, tags, saving: true }} />

<Story
	name="Error"
	args={{ open: true, selection, tags, error: 'Could not tag these transactions. Try again.' }}
/>
