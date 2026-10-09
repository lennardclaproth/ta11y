<script lang="ts">
	// The entry point from Cashflow: point a transaction (or a selection) at a recurring item.
	// Two routes, one dialog — adding to something that exists is the common one, so it comes first.
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Radio from '$lib/components/atoms/radio/Radio.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import {
		recurringExpenses,
		recurringIncome,
		rhythmLabels,
		type CashflowRow
	} from '../recurring.fixture';

	type Props = { open?: boolean; selection: CashflowRow[] };

	let { open = $bindable(false), selection }: Props = $props();

	let mode = $state('new');
	let name = $state('Cloud Locker');
	let direction = $state('out');
	let rhythm = $state('monthly');
	let existing = $state('rc-006');

	const itemOptions = [...recurringExpenses, ...recurringIncome].map((item) => ({
		value: item.id,
		label: `${item.name} · ${rhythmLabels[item.rhythm]}`
	}));

	const rhythmOptions = Object.entries(rhythmLabels).map(([value, label]) => ({ value, label }));
</script>

<Dialog bind:open title="Mark as recurring" size="md">
	<div class="flex flex-col gap-5">
		<!-- What you picked, repeated here: the dialog covers the table it came from. -->
		<div class="border-t border-slate-400 pt-3">
			<Text as="span" size="xs" tone="muted">
				{selection.length}
				{selection.length === 1 ? 'transaction' : 'transactions'} selected
			</Text>
			<ul class="mt-2 flex flex-col gap-1">
				{#each selection as row (row.id)}
					<li class="flex items-baseline justify-between gap-3 text-sm">
						<span class="min-w-0 truncate text-slate-700">
							{formatDisplayDate(row.date)} · {row.description}
						</span>
						<Money amount={row.amount} currency="EUR" size="sm" />
					</li>
				{/each}
			</ul>
		</div>

		<fieldset class="flex flex-col gap-3">
			<legend class="sr-only">Where these transactions go</legend>
			<label class="flex items-center gap-2 text-sm text-slate-800">
				<Radio bind:group={mode} value="existing" name="mark-mode" />
				Add to an existing item
			</label>
			{#if mode === 'existing'}
				<div class="pl-7">
					<FormField label="Recurring item" id="mark-existing">
						{#snippet children(field)}
							<Select id={field.id} bind:value={existing} options={itemOptions} />
						{/snippet}
					</FormField>
				</div>
			{/if}

			<label class="flex items-center gap-2 text-sm text-slate-800">
				<Radio bind:group={mode} value="new" name="mark-mode" />
				Start a new item
			</label>
			{#if mode === 'new'}
				<div class="flex flex-col gap-4 pl-7">
					<FormField
						label="To whom"
						id="mark-name"
						hint="Taken from the description — change it to the name you recognise."
					>
						{#snippet children(field)}
							<Input id={field.id} bind:value={name} ariaDescribedby={field.describedby} />
						{/snippet}
					</FormField>

					<fieldset class="flex flex-col gap-2">
						<legend class="mb-1 text-sm font-medium text-slate-800">Direction</legend>
						<label class="flex items-center gap-2 text-sm text-slate-800">
							<Radio bind:group={direction} value="out" name="mark-direction" />
							Expense — money you pay
						</label>
						<label class="flex items-center gap-2 text-sm text-slate-800">
							<Radio bind:group={direction} value="in" name="mark-direction" />
							Income — money you receive
						</label>
					</fieldset>

					<FormField label="How often" id="mark-rhythm">
						{#snippet children(field)}
							<Select id={field.id} bind:value={rhythm} options={rhythmOptions} />
						{/snippet}
					</FormField>
				</div>
			{/if}
		</fieldset>

		<Text size="xs" tone="muted">
			Marking does not change the tag, the ignored status or your monthly totals.
		</Text>
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" shape="default" onclick={() => (open = false)}>
			Cancel
		</Button>
		<Button shape="default">Save</Button>
	{/snippet}
</Dialog>
