<script lang="ts">
	/**
	 * The transaction form body, so a save-level notice can be judged against real fields.
	 * The failing field mirrors FormField's markup with the proposed FieldError line.
	 */
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Label from '$lib/components/atoms/label/Label.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import FieldError from './FieldError.svelte';

	type Props = {
		/** Show the per-field error on Amount. */
		invalid?: boolean;
	};

	let { invalid = true }: Props = $props();

	const typeOptions = [
		{ value: 'expense', label: 'Expense' },
		{ value: 'income', label: 'Income' }
	];
</script>

<div class="space-y-3">
	<div class="flex items-center justify-between gap-3 border-b border-slate-200 pb-2">
		<span class="font-heading text-xl text-slate-900">Mar 2, 2026</span>
		<span class="text-sm text-slate-500">Change date</span>
	</div>

	<FormField label="Type" id="demo-type">
		{#snippet children(ctx)}
			<Select id={ctx.id} value="expense" options={typeOptions} ariaLabel="Transaction type" />
		{/snippet}
	</FormField>

	{#if invalid}
		<div class="flex flex-col gap-1.5">
			<Label for="demo-amount" required>Amount</Label>
			<CurrencyInput id="demo-amount" value="" intent="error" ariaDescribedby="demo-amount-error" />
			<FieldError id="demo-amount-error" message="Amount is required" />
		</div>
	{:else}
		<div class="flex flex-col gap-1.5">
			<Label for="demo-amount" required>Amount</Label>
			<CurrencyInput id="demo-amount" value="1200.00" />
		</div>
	{/if}

	<FormField label="Description" id="demo-description">
		{#snippet children(ctx)}
			<Input id={ctx.id} value="Monthly rent" placeholder="e.g. Albert Heijn" />
		{/snippet}
	</FormField>
</div>
