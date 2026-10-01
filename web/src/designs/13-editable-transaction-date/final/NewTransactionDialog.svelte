<script lang="ts">
	// Final — "New transaction" (variant C's ruled date header). The date opens on today and is
	// changed either with the Today shortcut or the existing picker, whose calendar now opens
	// inside the dialog. Same dialog on Cashflow and Portfolio; only the fields below differ.
	import { untrack } from 'svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import DateHeader from './DateHeader.svelte';

	type Props = {
		/** Starting date; the real form always opens on today. */
		initialDate?: string;
		/** Prototypes open the calendar so the fixed screenshot shows the repaired popover. */
		showCalendar?: boolean;
		/** Refused save: the chosen date makes this an exact duplicate of an existing transaction. */
		error?: string | null;
		saving?: boolean;
	};

	const today = todayISO();

	let { initialDate = today, showCalendar = false, error = null, saving = false }: Props = $props();

	let date = $state(untrack(() => initialDate));
	let pickerOpen = $state(false);
	let amount = $state('89.50');
	let type = $state('expense');
	let description = $state('Bike repair');
	let tag = $state('household');
	let note = $state('');

	const typeOptions = [
		{ value: 'expense', label: 'Expense' },
		{ value: 'income', label: 'Income' }
	];

	// The dialog is mounted from an effect, so wait a tick: the popover measures its trigger and
	// would otherwise land in the top-left corner of the screenshot.
	$effect(() => {
		if (!showCalendar) return;
		const timer = setTimeout(() => (pickerOpen = true), 250);
		return () => clearTimeout(timer);
	});
</script>

<Dialog open title="New transaction" size="md" dismissible>
	<div class="space-y-3">
		<DateHeader bind:value={date} bind:open={pickerOpen} {today} disabled={saving} />

		{#if error}
			<Alert intent="error" title="This date is already taken">
				{error}
			</Alert>
		{/if}

		<FormField label="Type" id="f-type">
			{#snippet children(ctx)}
				<Select id={ctx.id} bind:value={type} options={typeOptions} ariaLabel="Transaction type" />
			{/snippet}
		</FormField>

		<FormField label="Amount" id="f-amount">
			{#snippet children(ctx)}
				<CurrencyInput id={ctx.id} bind:value={amount} ariaDescribedby={ctx.describedby} />
			{/snippet}
		</FormField>

		<FormField label="Description" id="f-description">
			{#snippet children(ctx)}
				<Input id={ctx.id} bind:value={description} placeholder="e.g. Albert Heijn" />
			{/snippet}
		</FormField>

		<div class="grid grid-cols-2 gap-3">
			<FormField label="Tag" id="f-tag" hint="Optional">
				{#snippet children(ctx)}
					<Input id={ctx.id} bind:value={tag} placeholder="e.g. groceries" />
				{/snippet}
			</FormField>
			<FormField label="Note" id="f-note" hint="Optional">
				{#snippet children(ctx)}
					<Input id={ctx.id} bind:value={note} />
				{/snippet}
			</FormField>
		</div>
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" disabled={saving}>Cancel</Button>
		<Button intent="success" loading={saving}>{saving ? 'Saving' : 'Save'}</Button>
	{/snippet}
</Dialog>
