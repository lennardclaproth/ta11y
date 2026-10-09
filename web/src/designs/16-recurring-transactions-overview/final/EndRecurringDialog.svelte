<script lang="ts">
	// Ending is a date, not a delete: everything already paid stays readable afterwards.
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { endMonthOptions } from '../recurring.format';
	import type { RecurringItem } from '../recurring.fixture';

	type Props = { open?: boolean; item: RecurringItem };

	let { open = $bindable(false), item }: Props = $props();

	const months = endMonthOptions();
	let month = $state(months[1].value);
</script>

<Dialog bind:open title="End {item.name}" size="sm">
	<div class="flex flex-col gap-4">
		<Text size="sm">
			From the month you pick, {item.name} is no longer expected and new transactions are no longer linked
			to it automatically.
		</Text>

		<FormField
			label="Ends from"
			id="end-month"
			hint="Transactions before this month stay linked and keep their history."
		>
			{#snippet children(field)}
				<Select
					id={field.id}
					bind:value={month}
					options={months}
					ariaDescribedby={field.describedby}
				/>
			{/snippet}
		</FormField>
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" shape="default" onclick={() => (open = false)}>
			Cancel
		</Button>
		<Button shape="default">End from {months.find((m) => m.value === month)?.label}</Button>
	{/snippet}
</Dialog>
