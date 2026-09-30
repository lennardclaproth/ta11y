<script lang="ts">
	// Variant B — "Veld en lade": the create form keeps its compact date field (the popover now opens
	// inside the dialog); changing a date afterwards happens in a detail drawer that keeps the list
	// and the full transaction in view.
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { formatDisplayDate, todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import { scaledToNumber } from '$lib/api/money';
	import InFormDatePicker from '../shared/InFormDatePicker.svelte';
	import LedgerBackdrop from '../shared/LedgerBackdrop.svelte';
	import { backdatedDate, daysAgoLabel, designTransactions } from '../shared/design-data';

	type Props = {
		scene?: 'create' | 'edit';
	};

	let { scene = 'create' }: Props = $props();

	const today = todayISO();
	const editRow = designTransactions[0];

	// One pair per scene so each story starts from a fixed, non-reactive value.
	let createDate = $state(backdatedDate);
	let createPickerOpen = $state(false);
	let editDate = $state(backdatedDate);
	let amount = $state('89.50');
	let type = $state('expense');
	let description = $state('Bike repair');
	let tag = $state('household');
	let note = $state('');

	const typeOptions = [
		{ value: 'expense', label: 'Expense' },
		{ value: 'income', label: 'Income' }
	];

	// The prototype shows the calendar already open. The dialog itself is shown from an effect, so
	// wait a tick: the popover measures its trigger and would otherwise land in the top-left corner.
	$effect(() => {
		if (scene !== 'create') return;
		const timer = setTimeout(() => (createPickerOpen = true), 250);
		return () => clearTimeout(timer);
	});
</script>

<LedgerBackdrop />

{#if scene === 'create'}
	<Dialog open title="New transaction" size="md" dismissible>
		<div class="space-y-3">
			<div class="grid grid-cols-2 gap-3">
				<FormField label="Date" id="b-date" hint="Today or earlier">
					<InFormDatePicker
						bind:value={createDate}
						bind:open={createPickerOpen}
						max={today}
						class="w-full"
						ariaLabel="Transaction date"
					/>
				</FormField>

				<FormField label="Type" id="b-type">
					{#snippet children(ctx)}
						<Select
							id={ctx.id}
							bind:value={type}
							options={typeOptions}
							ariaLabel="Transaction type"
						/>
					{/snippet}
				</FormField>
			</div>

			<FormField label="Amount" id="b-amount">
				{#snippet children(ctx)}
					<CurrencyInput id={ctx.id} bind:value={amount} ariaDescribedby={ctx.describedby} />
				{/snippet}
			</FormField>

			<FormField label="Description" id="b-description">
				{#snippet children(ctx)}
					<Input id={ctx.id} bind:value={description} placeholder="e.g. Albert Heijn" />
				{/snippet}
			</FormField>

			<div class="grid grid-cols-2 gap-3">
				<FormField label="Tag" id="b-tag" hint="Optional">
					{#snippet children(ctx)}
						<Input id={ctx.id} bind:value={tag} placeholder="e.g. groceries" />
					{/snippet}
				</FormField>
				<FormField label="Note" id="b-note" hint="Optional">
					{#snippet children(ctx)}
						<Input id={ctx.id} bind:value={note} />
					{/snippet}
				</FormField>
			</div>
		</div>

		{#snippet footer()}
			<Button variant="ghost" intent="secondary">Cancel</Button>
			<Button intent="success">Save</Button>
		{/snippet}
	</Dialog>
{:else}
	<Drawer open title="Transaction" width="max-w-md">
		<div class="space-y-4">
			<div>
				<p class="font-heading text-2xl leading-tight text-slate-900">{editRow.description}</p>
				<div class="mt-1 flex items-center gap-2">
					<Badge intent="info" variant="soft" size="sm">Manual</Badge>
					<Text as="span" size="sm" tone="muted">{editRow.tag}</Text>
				</div>
			</div>

			<dl class="divide-y divide-slate-200 border-y border-slate-200">
				<div class="flex items-center justify-between py-3">
					<dt class="text-sm text-slate-500">Amount</dt>
					<dd><Money amount={scaledToNumber(editRow.amountCents)} currency="EUR" size="sm" /></dd>
				</div>
				<div class="flex items-center justify-between py-3">
					<dt class="text-sm text-slate-500">Note</dt>
					<dd class="text-sm text-slate-800">{editRow.note}</dd>
				</div>
				<div class="py-3">
					<div class="flex items-center justify-between gap-3">
						<dt class="text-sm text-slate-500">Date</dt>
						<dd>
							<InFormDatePicker
								bind:value={editDate}
								max={today}
								size="sm"
								ariaLabel="Transaction date"
							/>
						</dd>
					</div>
					<p class="mt-1.5 text-right text-sm text-slate-500">
						{daysAgoLabel(editDate, today)} · was {formatDisplayDate(editRow.date.slice(0, 10))}
					</p>
				</div>
			</dl>

			<Alert intent="info" title="Only the date can change">
				Amount, description and tag stay as entered. Imported transactions keep the date from their
				statement.
			</Alert>
		</div>

		{#snippet footer()}
			<div class="flex items-center justify-end gap-2">
				<Button variant="ghost" intent="secondary">Cancel</Button>
				<Button intent="success">Save date</Button>
			</div>
		{/snippet}
	</Drawer>
{/if}
