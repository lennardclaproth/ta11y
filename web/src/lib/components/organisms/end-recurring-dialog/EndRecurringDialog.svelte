<script lang="ts">
	// Ending is a date, not a delete: everything already paid stays readable afterwards.
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import { endMonthOptions } from '$lib/api/recurring';

	type Props = {
		open?: boolean;
		/** The item's name, so the dialog says what it is about to stop. */
		name: string;
		saving?: boolean;
		error?: string | null;
		onConfirm?: (month: string) => void;
	};

	let { open = $bindable(false), name, saving = false, error = null, onConfirm }: Props = $props();

	const months = endMonthOptions();
	// Next month by default: the current month is usually already paid.
	let month = $state(months[1]?.value ?? months[0].value);

	const chosen = $derived(months.find((option) => option.value === month)?.label ?? '');
</script>

<Dialog bind:open title="End {name}" size="sm" closeOnEscape={!saving}>
	<div class="flex flex-col gap-4">
		<Text size="sm">
			From the month you pick, {name} is no longer expected and new transactions are no longer linked
			to it automatically.
		</Text>

		<FormField
			label="Ends from"
			id="end-month"
			hint="Transactions before this month stay linked and keep their history."
			error={error ?? undefined}
		>
			{#snippet children(field)}
				<Select
					id={field.id}
					bind:value={month}
					options={months}
					disabled={saving}
					ariaDescribedby={field.describedby}
				/>
			{/snippet}
		</FormField>
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
		<Button shape="default" loading={saving} disabled={saving} onclick={() => onConfirm?.(month)}>
			End from {chosen}
		</Button>
	{/snippet}
</Dialog>
