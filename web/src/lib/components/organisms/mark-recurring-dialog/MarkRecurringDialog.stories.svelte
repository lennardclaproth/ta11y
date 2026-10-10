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

<script lang="ts">
	let lateItems = $state<typeof items>([]);
	let lateOpen = $state(true);
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

<!--
	The route is chosen once, when the dialog opens: items that arrive afterwards show up as
	a second radio but no longer move the choice, and the form keeps what was typed. That is
	why Cashflow reads the items before it opens the dialog, not after.
-->
<Story name="Items arriving after opening">
	{#snippet template()}
		<div class="flex flex-col gap-4">
			<button
				type="button"
				class="w-fit rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
				onclick={() => {
					lateOpen = false;
					lateItems = [];
					lateOpen = true;
				}}
			>
				Reopen with no items
			</button>
			<button
				type="button"
				class="w-fit rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
				onclick={() => (lateItems = items)}
			>
				Let the items arrive
			</button>
			<MarkRecurringDialog bind:open={lateOpen} {selection} items={lateItems} />
		</div>
	{/snippet}
</Story>
