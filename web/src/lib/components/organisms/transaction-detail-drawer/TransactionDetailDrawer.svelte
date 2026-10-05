<script lang="ts">
	import type { Snippet } from 'svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import {
		formatDisplayDate,
		relativeDayLabel,
		todayISO
	} from '$lib/components/molecules/calendar/calendar.utils';

	type Props = {
		/** Two-way bindable open state. */
		open?: boolean;
		/** Headline of the transaction (description, listing). */
		title: string;
		/** Where the row came from, as a word: "Manual", "ING", "Import". */
		originLabel: string;
		/** Secondary line next to the origin (tag, symbol). */
		subtitle?: string;
		/** Only manually entered transactions can be moved to another day. */
		editable?: boolean;
		/** Bindable working date ("YYYY-MM-DD"); differs from `originalDate` once changed. */
		date?: string;
		/** The date the transaction is stored on. */
		originalDate: string;
		/** Latest selectable day. */
		today?: string;
		saving?: boolean;
		/** Refused save, e.g. the new date would duplicate an existing transaction. */
		error?: string | null;
		onSave?: (date: string) => void;
		onClose?: () => void;
		/** Read-only rows (amount, note, quantity) rendered above the date. */
		details?: Snippet;
	};

	let {
		open = $bindable(false),
		title,
		originLabel,
		subtitle = '',
		editable = false,
		date = $bindable(''),
		originalDate,
		today = todayISO(),
		saving = false,
		error = null,
		onSave,
		onClose,
		details
	}: Props = $props();

	const changed = $derived(date !== originalDate);
</script>

<Drawer bind:open title="Transaction" width="max-w-md" {onClose}>
	<div class="space-y-4">
		<div>
			<p class="font-heading text-2xl leading-tight text-slate-900">{title}</p>
			<div class="mt-1 flex items-center gap-2">
				<Badge intent={editable ? 'info' : 'neutral'} variant="soft" size="sm">
					{originLabel}
				</Badge>
				{#if subtitle}
					<Text as="span" size="sm" tone="muted">{subtitle}</Text>
				{/if}
			</div>
		</div>

		{#if error}
			<!-- role="alert" so a refusal that appears after Save is announced, not just coloured. -->
			<div role="alert">
				<Alert intent="error" title="This date is already taken">{error}</Alert>
			</div>
		{/if}

		<dl class="divide-y divide-slate-200 border-y border-slate-200">
			{@render details?.()}
			<div class="py-3">
				<div class="flex items-center justify-between gap-3">
					<dt class="text-sm text-slate-500">Date</dt>
					<dd class="text-right">
						{#if editable}
							<!-- The drawer panel sits on the modal layer, so the portaled calendar has to be
							     raised above it or it opens behind the panel and no day can be clicked. -->
							<DatePicker
								value={date}
								max={today}
								disabled={saving}
								size="sm"
								layer="filterPopover"
								ariaLabel="Transaction date"
								onChange={(next) => (date = next ?? date)}
							/>
						{:else}
							<span class="text-sm text-slate-800">{formatDisplayDate(originalDate)}</span>
						{/if}
					</dd>
				</div>
				<p class="mt-1.5 text-right text-sm text-slate-500">
					{#if !editable}
						{relativeDayLabel(originalDate, today)} · from your statement
					{:else if changed}
						{relativeDayLabel(date, today)} · was {formatDisplayDate(originalDate)}
					{:else}
						{relativeDayLabel(originalDate, today)} · today or earlier
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
		<Button variant="ghost" intent="secondary" disabled={saving} onclick={() => (open = false)}>
			Cancel
		</Button>
		<Button
			intent="success"
			loading={saving}
			disabled={!editable || !changed}
			onclick={() => onSave?.(date)}
		>
			{saving ? 'Saving' : 'Save date'}
		</Button>
	{/snippet}
</Drawer>
