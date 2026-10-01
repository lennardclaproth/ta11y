<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import TransactionDetailDrawer from './TransactionDetailDrawer.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/TransactionDetailDrawer',
		component: TransactionDetailDrawer,
		tags: ['autodocs']
	});

	// Fixed days and invented amounts so the stories read the same whenever they are opened.
	const today = '2026-09-30';
	const originalDate = '2026-09-30';
	const movedDate = '2026-07-14';
</script>

<script lang="ts">
	let date = $state(movedDate);
</script>

{#snippet amountAndNote()}
	<div class="flex items-center justify-between gap-3 py-3">
		<dt class="text-sm text-slate-500">Amount</dt>
		<dd><Money amount={120} currency="EUR" size="sm" /></dd>
	</div>
	<div class="flex items-center justify-between gap-3 py-3">
		<dt class="text-sm text-slate-500">Note</dt>
		<dd class="text-sm text-slate-800">Entered late</dd>
	</div>
{/snippet}

<Story name="Manual, unchanged" asChild>
	<TransactionDetailDrawer
		open
		title="Bike repair"
		originLabel="Manual"
		subtitle="household"
		editable
		date={originalDate}
		{originalDate}
		{today}
		details={amountAndNote}
	/>
</Story>

<Story name="Manual, moved months back" asChild>
	<TransactionDetailDrawer
		open
		title="Bike repair"
		originLabel="Manual"
		subtitle="household"
		editable
		bind:date
		{originalDate}
		{today}
		details={amountAndNote}
	/>
</Story>

<Story name="Saving" asChild>
	<TransactionDetailDrawer
		open
		title="Bike repair"
		originLabel="Manual"
		subtitle="household"
		editable
		date={movedDate}
		{originalDate}
		{today}
		saving
		details={amountAndNote}
	/>
</Story>

<Story name="Refused" asChild>
	<TransactionDetailDrawer
		open
		title="Bike repair"
		originLabel="Manual"
		subtitle="household"
		editable
		date={movedDate}
		{originalDate}
		{today}
		error="A transaction with the same amount and description already exists on 14 Jul 2026."
		details={amountAndNote}
	/>
</Story>

<Story name="Imported" asChild>
	<TransactionDetailDrawer
		open
		title="Monthly salary"
		originLabel="ING"
		subtitle="salary"
		date={originalDate}
		{originalDate}
		{today}
		details={amountAndNote}
	/>
</Story>
