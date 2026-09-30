<script lang="ts">
	// Variant C — "In de ledger": the date gets a ruled header line at the top of the form, and an
	// existing manual row is corrected in place in the table — no second dialog or drawer.
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { formatDisplayDate, todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import InFormDatePicker from '../shared/InFormDatePicker.svelte';
	import LedgerBackdrop from '../shared/LedgerBackdrop.svelte';
	import { backdatedDate, daysAgoLabel, designTransactions, isManual } from '../shared/design-data';
	import type { CashflowTransaction } from '$lib/api/types';

	type Props = {
		scene?: 'create' | 'edit';
	};

	let { scene = 'create' }: Props = $props();

	const today = todayISO();
	const editingId = 'design-3';

	let date = $state(backdatedDate);
	let pickerOpen = $state(false);
	let rowDate = $state('2026-07-14');
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
		const timer = setTimeout(() => (pickerOpen = true), 250);
		return () => clearTimeout(timer);
	});
</script>

{#snippet dateCell(row: CashflowTransaction)}
	{#if row.id === editingId}
		<InFormDatePicker bind:value={rowDate} max={today} size="sm" ariaLabel="Transaction date" />
	{:else if isManual(row)}
		<button
			type="button"
			class="group -mx-1 inline-flex items-center gap-1.5 rounded-md px-1 py-0.5 text-slate-700 underline decoration-slate-300 decoration-dotted underline-offset-4 transition-colors hover:bg-slate-100 hover:text-slate-900 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
		>
			<span>{formatDisplayDate(row.date.slice(0, 10))}</span>
			<Icon
				icon="heroicons:pencil-square"
				size="sm"
				class="text-slate-400 group-hover:text-slate-600"
			/>
		</button>
	{:else}
		<span>{formatDisplayDate(row.date.slice(0, 10))}</span>
	{/if}
{/snippet}

{#snippet actionCell(row: CashflowTransaction)}
	{#if row.id === editingId}
		<div class="flex items-center justify-end gap-2">
			<Button size="sm" variant="ghost" intent="secondary">Cancel</Button>
			<Button size="sm" intent="success">Save date</Button>
		</div>
	{/if}
{/snippet}

{#if scene === 'create'}
	<LedgerBackdrop />

	<Dialog open title="New transaction" size="md" dismissible>
		<div class="space-y-3">
			<div
				class="-mx-5 -mt-4 mb-1 flex flex-wrap items-end justify-between gap-3 border-b border-slate-200 bg-taupe-50 px-5 py-3"
			>
				<div class="min-w-0">
					<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">Date</p>
					<p class="font-heading text-2xl leading-tight text-slate-900">
						{formatDisplayDate(date)}
					</p>
					<Text as="span" size="sm" tone="muted">
						{daysAgoLabel(date, today)} · today or earlier
					</Text>
				</div>
				<div class="flex items-center gap-2">
					<Button size="sm" variant="outline" intent="secondary">Today</Button>
					<InFormDatePicker
						bind:value={date}
						bind:open={pickerOpen}
						max={today}
						size="sm"
						placeholder="Pick a day"
						ariaLabel="Transaction date"
					/>
				</div>
			</div>

			<FormField label="Type" id="c-type">
				{#snippet children(ctx)}
					<Select id={ctx.id} bind:value={type} options={typeOptions} ariaLabel="Transaction type" />
				{/snippet}
			</FormField>

			<FormField label="Amount" id="c-amount">
				{#snippet children(ctx)}
					<CurrencyInput id={ctx.id} bind:value={amount} ariaDescribedby={ctx.describedby} />
				{/snippet}
			</FormField>

			<FormField label="Description" id="c-description">
				{#snippet children(ctx)}
					<Input id={ctx.id} bind:value={description} placeholder="e.g. Albert Heijn" />
				{/snippet}
			</FormField>

			<div class="grid grid-cols-2 gap-3">
				<FormField label="Tag" id="c-tag" hint="Optional">
					{#snippet children(ctx)}
						<Input id={ctx.id} bind:value={tag} placeholder="e.g. groceries" />
					{/snippet}
				</FormField>
				<FormField label="Note" id="c-note" hint="Optional">
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
	<LedgerBackdrop rows={designTransactions} {dateCell} {actionCell} />
{/if}
