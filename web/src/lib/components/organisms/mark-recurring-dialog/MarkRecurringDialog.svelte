<script lang="ts">
	// The entry point from Cashflow: point a transaction (or a selection) at a recurring
	// item. Two routes, one dialog — adding to something that exists is the common one,
	// so it comes first. Marking changes no tag, no ignored status and no monthly total.
	import { untrack } from 'svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Radio from '$lib/components/atoms/radio/Radio.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { scaledToNumber } from '$lib/api/money';
	import { rhythmLabels } from '$lib/api/recurring';
	import type {
		CashflowDirection,
		CashflowTransaction,
		RecurringItem,
		RecurringRhythm
	} from '$lib/api/types';
	import type { MarkRecurringMode, MarkRecurringValue } from './mark-recurring-dialog.types';

	type Props = {
		open?: boolean;
		/** The transactions the user selected on Cashflow. */
		selection: CashflowTransaction[];
		/** Items already confirmed, offered as the first route. */
		items: RecurringItem[];
		saving?: boolean;
		error?: string | null;
		onSubmit?: (value: MarkRecurringValue) => void;
	};

	let {
		open = $bindable(false),
		selection,
		items,
		saving = false,
		error = null,
		onSubmit
	}: Props = $props();

	// The Select and Radio atoms speak plain strings; the values narrow on submit.
	let mode = $state<MarkRecurringMode>('existing');
	let existing = $state('');
	let name = $state('');
	let direction = $state('out');
	let rhythm = $state('monthly');

	const itemOptions = $derived(
		items.map((item) => ({ value: item.id, label: `${item.name} · ${rhythmLabels[item.rhythm]}` }))
	);
	const rhythmOptions = Object.entries(rhythmLabels).map(([value, label]) => ({ value, label }));

	// Opening the dialog seeds the form from the selection: the description is where the
	// counterparty hides, and the direction is already known from the transactions.
	// An account with no items yet has only one route, so it opens on "new".
	// Only opening seeds it: the items arrive after the dialog is already up, and a
	// realtime refresh replaces the selection, so following those would wipe the form
	// under the user's hands.
	$effect(() => {
		if (!open) return;
		untrack(() => {
			const first = selection[0];
			name = first?.description ?? '';
			direction = first?.direction ?? 'out';
			existing = items[0]?.id ?? '';
			mode = items.length > 0 ? 'existing' : 'new';
		});
	});

	const canSubmit = $derived(
		selection.length > 0 && (mode === 'existing' ? existing !== '' : name.trim() !== '') && !saving
	);

	function submit() {
		if (!canSubmit) return;
		onSubmit?.(
			mode === 'existing'
				? { mode: 'existing', itemId: existing }
				: {
						mode: 'new',
						name: name.trim(),
						direction: direction as CashflowDirection,
						rhythm: rhythm as RecurringRhythm
					}
		);
	}
</script>

<Dialog bind:open title="Mark as recurring" size="md" closeOnEscape={!saving}>
	<div class="flex flex-col gap-5">
		<!-- What you picked, repeated here: the dialog covers the table it came from. -->
		<div class="border-t border-slate-400 pt-3">
			<Text as="span" size="xs" tone="muted">
				{selection.length}
				{selection.length === 1 ? 'transaction' : 'transactions'} selected
			</Text>
			<ul class="mt-2 flex max-h-40 flex-col gap-1 overflow-y-auto">
				{#each selection as transaction (transaction.id)}
					<li class="flex items-baseline justify-between gap-3 text-sm">
						<span class="min-w-0 truncate text-slate-700">
							{formatDisplayDate(transaction.date.slice(0, 10))} · {transaction.description}
						</span>
						<Money amount={scaledToNumber(transaction.amountCents)} currency="EUR" size="sm" />
					</li>
				{/each}
			</ul>
		</div>

		<fieldset class="flex flex-col gap-3">
			<legend class="sr-only">Where these transactions go</legend>

			{#if items.length > 0}
				<label class="flex items-center gap-2 text-sm text-slate-800">
					<Radio bind:group={mode} value="existing" name="mark-mode" disabled={saving} />
					Add to an existing item
				</label>
				{#if mode === 'existing'}
					<div class="pl-7">
						<FormField label="Recurring item" id="mark-existing">
							{#snippet children(field)}
								<Select
									id={field.id}
									bind:value={existing}
									options={itemOptions}
									disabled={saving}
								/>
							{/snippet}
						</FormField>
					</div>
				{/if}
			{/if}

			<label class="flex items-center gap-2 text-sm text-slate-800">
				<Radio bind:group={mode} value="new" name="mark-mode" disabled={saving} />
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
							<Input
								id={field.id}
								bind:value={name}
								disabled={saving}
								ariaDescribedby={field.describedby}
							/>
						{/snippet}
					</FormField>

					<fieldset class="flex flex-col gap-2">
						<legend class="mb-1 text-sm font-medium text-slate-800">Direction</legend>
						<label class="flex items-center gap-2 text-sm text-slate-800">
							<Radio bind:group={direction} value="out" name="mark-direction" disabled={saving} />
							Expense — money you pay
						</label>
						<label class="flex items-center gap-2 text-sm text-slate-800">
							<Radio bind:group={direction} value="in" name="mark-direction" disabled={saving} />
							Income — money you receive
						</label>
					</fieldset>

					<FormField label="How often" id="mark-rhythm">
						{#snippet children(field)}
							<Select id={field.id} bind:value={rhythm} options={rhythmOptions} disabled={saving} />
						{/snippet}
					</FormField>
				</div>
			{/if}
		</fieldset>

		{#if error}
			<p
				class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700"
				role="alert"
			>
				{error}
			</p>
		{/if}

		<Text size="xs" tone="muted">
			Marking does not change the tag, the ignored status or your monthly totals.
		</Text>
	</div>

	{#snippet footer()}
		<Button
			variant="ghost"
			intent="secondary"
			shape="default"
			disabled={saving}
			onclick={() => (open = false)}
		>
			Cancel
		</Button>
		<Button shape="default" loading={saving} disabled={!canSubmit} onclick={submit}>Save</Button>
	{/snippet}
</Dialog>
