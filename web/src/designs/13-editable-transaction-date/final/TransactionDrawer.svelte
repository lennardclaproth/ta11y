<script lang="ts">
	// Final — "Change date" (variant B's detail drawer). Opening a row keeps the ledger in view and
	// shows the whole transaction; only the date is editable, and only for manual transactions.
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { formatDisplayDate, todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import { scaledToNumber } from '$lib/api/money';
	import InFormDatePicker from '../shared/InFormDatePicker.svelte';
	import { daysAgoLabel, isManual, sourceLabel } from '../shared/design-data';
	import type { CashflowTransaction } from '$lib/api/types';

	type Props = {
		row: CashflowTransaction;
		/** Date shown in the picker; differs from the row's own date once it is changed. */
		date?: string;
		saving?: boolean;
		/** Refused save, e.g. the new date makes this transaction a duplicate. */
		error?: string | null;
	};

	const today = todayISO();

	let { row, date = $bindable(''), saving = false, error = null }: Props = $props();

	const originalDate = $derived(row.date.slice(0, 10));
	const editable = $derived(isManual(row));
	const changed = $derived(date !== originalDate);
</script>

<Drawer open title="Transaction" width="max-w-md">
	<div class="space-y-4">
		<div>
			<p class="font-heading text-2xl leading-tight text-slate-900">{row.description}</p>
			<div class="mt-1 flex items-center gap-2">
				<Badge intent={editable ? 'info' : 'neutral'} variant="soft" size="sm">
					{sourceLabel(row)}
				</Badge>
				<Text as="span" size="sm" tone="muted">{row.tag}</Text>
			</div>
		</div>

		{#if error}
			<Alert intent="error" title="This date is already taken">
				{error}
			</Alert>
		{/if}

		<dl class="divide-y divide-slate-200 border-y border-slate-200">
			<div class="flex items-center justify-between gap-3 py-3">
				<dt class="text-sm text-slate-500">Amount</dt>
				<dd><Money amount={scaledToNumber(row.amountCents)} currency="EUR" size="sm" /></dd>
			</div>
			<div class="flex items-center justify-between gap-3 py-3">
				<dt class="text-sm text-slate-500">Note</dt>
				<dd class="text-sm text-slate-800">{row.note}</dd>
			</div>
			<div class="py-3">
				<div class="flex items-center justify-between gap-3">
					<dt class="text-sm text-slate-500">Date</dt>
					<dd class="text-right">
						{#if editable}
							<InFormDatePicker
								bind:value={date}
								max={today}
								disabled={saving}
								size="sm"
								ariaLabel="Transaction date"
							/>
						{:else}
							<span class="text-sm text-slate-800">{formatDisplayDate(originalDate)}</span>
						{/if}
					</dd>
				</div>
				<p class="mt-1.5 text-right text-sm text-slate-500">
					{#if !editable}
						{daysAgoLabel(originalDate, today)} · from your statement
					{:else if changed}
						{daysAgoLabel(date, today)} · was {formatDisplayDate(originalDate)}
					{:else}
						{daysAgoLabel(originalDate, today)} · today or earlier
					{/if}
				</p>
			</div>
		</dl>

		{#if editable}
			<Alert intent="info" title="Only the date can change">
				Amount, description and tag stay as entered. Change them by removing this transaction and
				adding it again.
			</Alert>
		{:else}
			<Alert intent="info" title="Imported transactions keep their statement date">
				This transaction came from an import, so its date is fixed. Only transactions you entered
				yourself can be moved to another day.
			</Alert>
		{/if}
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" disabled={saving}>Cancel</Button>
		<Button intent="success" loading={saving} disabled={!editable || !changed}>
			{saving ? 'Saving' : 'Save date'}
		</Button>
	{/snippet}
</Drawer>
