<script lang="ts">
	// Variant A — "Kalender in beeld": the existing Calendar molecule is rendered inline inside the
	// form, so the date is never hidden behind a popover that the modal swallows.
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Calendar from '$lib/components/molecules/calendar/Calendar.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import {
		formatDisplayDate,
		parseISODate,
		startOfMonthUTC,
		todayISO
	} from '$lib/components/molecules/calendar/calendar.utils';
	import { scaledToNumber } from '$lib/api/money';
	import LedgerBackdrop from '../shared/LedgerBackdrop.svelte';
	import { backdatedDate, daysAgoLabel, designTransactions } from '../shared/design-data';

	type Props = {
		/** `create` = New transaction, `edit` = change the date of an existing manual row. */
		scene?: 'create' | 'edit';
	};

	let { scene = 'create' }: Props = $props();

	const today = todayISO();
	const editRow = designTransactions[0];

	// One pair per scene so each story starts from a fixed, non-reactive value.
	let createDate = $state(backdatedDate);
	let createMonth = $state(startOfMonthUTC(parseISODate(backdatedDate) ?? new Date()));
	let editDate = $state(backdatedDate);
	let editMonth = $state(startOfMonthUTC(parseISODate(backdatedDate) ?? new Date()));
	let amount = $state('89.50');
	let type = $state('expense');
	let description = $state('Bike repair');
	let tag = $state('household');
	let note = $state('');

	const typeOptions = [
		{ value: 'expense', label: 'Expense' },
		{ value: 'income', label: 'Income' }
	];
</script>

<LedgerBackdrop />

{#snippet dateSummary(iso: string)}
	<div class="mt-3 border-t border-slate-200 pt-3">
		<p class="font-heading text-xl leading-tight text-slate-900">{formatDisplayDate(iso)}</p>
		<Text as="p" size="sm" tone="muted">
			{daysAgoLabel(iso, today)} · dates after today are not selectable
		</Text>
	</div>
{/snippet}

{#if scene === 'create'}
	<Dialog open title="New transaction" size="xl" dismissible>
		<div class="grid gap-5 md:grid-cols-[minmax(0,1fr)_17rem]">
			<!-- Date leads on narrow screens: it is what this form is being opened for. -->
			<div
				class="order-1 border-b border-slate-200 pb-4 md:order-2 md:border-b-0 md:border-l md:pb-0 md:pl-5"
			>
				<h3
					class="mb-2 border-b border-slate-200 pb-1 text-xs font-semibold tracking-wide text-slate-700 uppercase"
				>
					Date
				</h3>
				<Calendar
					bind:month={createMonth}
					mode="single"
					selected={createDate}
					max={today}
					onSelect={(iso) => (createDate = iso)}
					class="mx-auto"
				/>
				{@render dateSummary(createDate)}
			</div>

			<div class="order-2 space-y-3 md:order-1">
				<FormField label="Type" id="a-type">
					{#snippet children(ctx)}
						<Select
							id={ctx.id}
							bind:value={type}
							options={typeOptions}
							ariaLabel="Transaction type"
						/>
					{/snippet}
				</FormField>

				<FormField label="Amount" id="a-amount">
					{#snippet children(ctx)}
						<CurrencyInput id={ctx.id} bind:value={amount} ariaDescribedby={ctx.describedby} />
					{/snippet}
				</FormField>

				<FormField label="Description" id="a-description">
					{#snippet children(ctx)}
						<Input id={ctx.id} bind:value={description} placeholder="e.g. Albert Heijn" />
					{/snippet}
				</FormField>

				<div class="grid grid-cols-2 gap-3">
					<FormField label="Tag" id="a-tag" hint="Optional">
						{#snippet children(ctx)}
							<Input id={ctx.id} bind:value={tag} placeholder="e.g. groceries" />
						{/snippet}
					</FormField>
					<FormField label="Note" id="a-note" hint="Optional">
						{#snippet children(ctx)}
							<Input id={ctx.id} bind:value={note} />
						{/snippet}
					</FormField>
				</div>
			</div>

		</div>

		{#snippet footer()}
			<Button variant="ghost" intent="secondary">Cancel</Button>
			<Button intent="success">Save</Button>
		{/snippet}
	</Dialog>
{:else}
	<Dialog open title="Change date" size="md" dismissible>
		<div class="space-y-4">
			<div class="flex items-baseline justify-between gap-3 border-b border-slate-200 pb-3">
				<div class="min-w-0">
					<p class="truncate text-slate-800">{editRow.description}</p>
					<Text as="span" size="sm" tone="muted">
						{editRow.tag} · entered manually
					</Text>
				</div>
				<Money amount={scaledToNumber(editRow.amountCents)} currency="EUR" size="sm" />
			</div>

			<div class="flex flex-col items-center">
				<Calendar
					bind:month={editMonth}
					mode="single"
					selected={editDate}
					max={today}
					onSelect={(iso) => (editDate = iso)}
				/>
			</div>

			<div class="border-t border-slate-200 pt-3">
				<p class="font-heading text-xl leading-tight text-slate-900">
					{formatDisplayDate(editDate)}
				</p>
				<Text as="p" size="sm" tone="muted">
					{daysAgoLabel(editDate, today)} · was {formatDisplayDate(editRow.date.slice(0, 10))}
				</Text>
				<Text as="p" size="sm" tone="muted">
					Only the date changes; amount, description and tag stay as entered.
				</Text>
			</div>
		</div>

		{#snippet footer()}
			<Button variant="ghost" intent="secondary">Cancel</Button>
			<Button intent="success">Save changes</Button>
		{/snippet}
	</Dialog>
{/if}
