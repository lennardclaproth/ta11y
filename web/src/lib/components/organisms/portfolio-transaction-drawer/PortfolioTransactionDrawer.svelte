<script lang="ts">
	import { untrack } from 'svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import ListingSearchSelect from '$lib/components/molecules/listing-search-select/ListingSearchSelect.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import {
		formatDisplayDate,
		relativeDayLabel,
		todayISO
	} from '$lib/components/molecules/calendar/calendar.utils';
	import type {
		ListingSearchRow,
		PortfolioTransaction,
		PortfolioTransactionType,
		UpdatePortfolioTransactionRequest
	} from '$lib/api/types';
	import {
		portfolioTransactionTypeOptions,
		typeNeedsListing,
		typeNeedsQuantity
	} from './portfolio-transaction-drawer.types';

	type Props = {
		/** Two-way bindable open state. */
		open?: boolean;
		transaction: PortfolioTransaction;
		/** Only manually entered transactions can be edited or deleted. */
		editable?: boolean;
		/** Where the row came from, as a word: "Manual", "DEGIRO". */
		originLabel: string;
		/** Latest selectable day. */
		today?: string;
		saving?: boolean;
		deleting?: boolean;
		/** Refused save or delete, in words the reader can act on. */
		error?: string | null;
		onSave?: (body: UpdatePortfolioTransactionRequest) => void;
		onDelete?: () => void;
		onClose?: () => void;
	};

	let {
		open = $bindable(false),
		transaction,
		editable = false,
		originLabel,
		today = todayISO(),
		saving = false,
		deleting = false,
		error = null,
		onSave,
		onDelete,
		onClose
	}: Props = $props();

	type FormValues = {
		date: string;
		type: PortfolioTransactionType;
		quantity: string;
		amount: string;
		description: string;
	};

	function seed(from: PortfolioTransaction): FormValues {
		return {
			date: from.occurred_at.slice(0, 10),
			type: from.type,
			quantity: typeNeedsQuantity(from.type) ? from.quantity : '',
			amount: from.amount,
			description: from.description
		};
	}

	// Seeded once, deliberately: the caller keys this drawer on the transaction id, so another
	// row gets a fresh component rather than one carrying these edits.
	const initial = untrack(() => seed(transaction));
	let values = $state<FormValues>({ ...initial });
	// The listing stays the stored one until the reader picks another.
	let changingListing = $state(untrack(() => !transaction.listing_id));
	let pickedListing = $state<ListingSearchRow | null>(null);
	let confirmOpen = $state(false);

	const busy = $derived(saving || deleting);
	const needsListing = $derived(typeNeedsListing(values.type));
	const needsQuantity = $derived(typeNeedsQuantity(values.type));
	const listingId = $derived(
		needsListing
			? changingListing
				? (pickedListing?.id ?? null)
				: (transaction.listing_id ?? null)
			: null
	);
	const currentListingLabel = $derived(transaction.symbol ?? transaction.isin ?? 'Linked listing');

	// A type without a listing drops the pick, so returning to a trade type does not quietly
	// submit a listing the remounted search box no longer shows.
	$effect(() => {
		if (!needsListing) pickedListing = null;
	});

	function keepCurrentListing() {
		changingListing = false;
		pickedListing = null;
	}

	const dirty = $derived(
		values.date !== initial.date ||
			values.type !== initial.type ||
			values.amount.trim() !== initial.amount ||
			values.description.trim() !== initial.description ||
			(needsQuantity && values.quantity.trim() !== initial.quantity) ||
			listingId !== (typeNeedsListing(initial.type) ? (transaction.listing_id ?? null) : null)
	);

	// Mirrors the create form: these are the rules the API would refuse, said before the round trip.
	const problem = $derived.by(() => {
		if (values.amount.trim() === '') return 'Enter an amount.';
		if (needsListing && !listingId) return 'Pick the listing this transaction belongs to.';
		if (needsQuantity && values.quantity.trim() === '') return 'Enter a quantity.';
		return null;
	});

	function save() {
		if (busy || !dirty || problem) return;
		onSave?.({
			occurred_at: values.date,
			type: values.type,
			listing_id: listingId ?? undefined,
			amount: values.amount.trim(),
			quantity: needsQuantity ? values.quantity.trim() : undefined,
			description: values.description.trim() || undefined
		});
	}

	function confirmDelete() {
		if (busy) return;
		confirmOpen = false;
		onDelete?.();
	}
</script>

<!-- Not dismissible mid-request, so a refusal is seen, nor while the delete confirmation is up,
     so Escape closes only that dialog rather than both and the edits with them. -->
<Drawer
	bind:open
	title="Transaction"
	width="max-w-md"
	dismissible={!busy && !confirmOpen}
	closeOnBackdrop={!busy && !confirmOpen}
	{onClose}
>
	<div class="space-y-4">
		<div>
			<p class="font-heading text-2xl leading-tight text-slate-900">
				{transaction.symbol ?? (transaction.description || 'Transaction')}
			</p>
			<div class="mt-1 flex items-center gap-2">
				<Badge intent={editable ? 'info' : 'neutral'} variant="soft" size="sm">{originLabel}</Badge>
				<Text as="span" size="sm" tone="muted">{transaction.type}</Text>
			</div>
		</div>

		{#if error}
			<!-- An error band carries role="alert", so a refusal after Save is announced. -->
			<NoticeBand
				intent="error"
				surface="inset"
				gutter="dialog"
				title="This transaction was not changed"
				class="-mx-5"
			>
				{error}
			</NoticeBand>
		{/if}

		{#if editable}
			<div class="space-y-3">
				<FormField label="Date" id="ptd-date" hint={relativeDayLabel(values.date, today)}>
					<!-- The drawer panel sits on the modal layer, so the portaled calendar has to be
					     raised above it or it opens behind the panel and no day can be clicked. -->
					<DatePicker
						value={values.date}
						max={today}
						disabled={busy}
						size="sm"
						layer="filterPopover"
						ariaLabel="Transaction date"
						onChange={(next) => (values.date = next ?? values.date)}
					/>
				</FormField>

				<FormField label="Type" id="ptd-type">
					{#snippet children(ctx)}
						<Select
							id={ctx.id}
							bind:value={values.type}
							options={[...portfolioTransactionTypeOptions]}
							disabled={busy}
							ariaLabel="Transaction type"
						/>
					{/snippet}
				</FormField>

				{#if needsListing}
					<FormField label="Listing" id="ptd-listing">
						{#if changingListing}
							<ListingSearchSelect bind:value={pickedListing} disabled={busy} ariaLabel="Listing" />
							{#if transaction.listing_id}
								<Button
									size="sm"
									variant="ghost"
									intent="secondary"
									class="mt-1"
									disabled={busy}
									onclick={keepCurrentListing}
								>
									Keep {currentListingLabel}
								</Button>
							{/if}
						{:else}
							<div class="flex items-center justify-between gap-3">
								<span class="text-sm text-slate-800">{currentListingLabel}</span>
								<Button
									size="sm"
									variant="ghost"
									intent="secondary"
									disabled={busy}
									onclick={() => (changingListing = true)}
								>
									Change
								</Button>
							</div>
						{/if}
					</FormField>
				{/if}

				<div class="grid grid-cols-2 gap-3">
					{#if needsQuantity}
						<FormField label="Quantity" id="ptd-quantity">
							{#snippet children(ctx)}
								<Input
									id={ctx.id}
									bind:value={values.quantity}
									disabled={busy}
									placeholder="e.g. 10"
									ariaDescribedby={ctx.describedby}
								/>
							{/snippet}
						</FormField>
					{/if}
					<FormField label="Amount" id="ptd-amount">
						{#snippet children(ctx)}
							<CurrencyInput
								id={ctx.id}
								bind:value={values.amount}
								disabled={busy}
								ariaDescribedby={ctx.describedby}
							/>
						{/snippet}
					</FormField>
				</div>

				<FormField label="Description" id="ptd-description" hint="Optional">
					{#snippet children(ctx)}
						<Input
							id={ctx.id}
							bind:value={values.description}
							disabled={busy}
							ariaDescribedby={ctx.describedby}
						/>
					{/snippet}
				</FormField>

				<!-- Live, so the reason Save stays disabled is announced as it appears. -->
				<div role="status">
					{#if dirty && problem}
						<Text as="p" size="sm" tone="muted">{problem}</Text>
					{/if}
				</div>
			</div>
		{:else}
			<dl class="divide-y divide-slate-200 border-y border-slate-200">
				<div class="flex items-center justify-between gap-3 py-3">
					<dt class="text-sm text-slate-500">Date</dt>
					<dd class="text-sm text-slate-800">{formatDisplayDate(initial.date)}</dd>
				</div>
				<div class="flex items-center justify-between gap-3 py-3">
					<dt class="text-sm text-slate-500">Amount</dt>
					<dd class="text-sm text-slate-800">{transaction.amount}</dd>
				</div>
				<div class="flex items-center justify-between gap-3 py-3">
					<dt class="text-sm text-slate-500">Quantity</dt>
					<dd class="text-sm text-slate-800">{transaction.quantity}</dd>
				</div>
				{#if transaction.description}
					<div class="flex items-center justify-between gap-3 py-3">
						<dt class="text-sm text-slate-500">Description</dt>
						<dd class="text-right text-sm text-slate-800">{transaction.description}</dd>
					</div>
				{/if}
			</dl>
			<NoticeBand
				intent="info"
				surface="inset"
				gutter="dialog"
				title="Imported transactions stay as the statement records them"
				class="-mx-5"
			>
				This transaction came from an import, so it cannot be changed or deleted here. Re-uploading
				the statement recognises it by exactly these values.
			</NoticeBand>
		{/if}
	</div>

	{#snippet footer()}
		{#if editable}
			<Button
				variant="ghost"
				intent="error"
				class="mr-auto"
				disabled={busy}
				loading={deleting}
				onclick={() => (confirmOpen = true)}
			>
				{deleting ? 'Deleting' : 'Delete'}
			</Button>
			<Button variant="ghost" intent="secondary" disabled={busy} onclick={() => (open = false)}>
				Cancel
			</Button>
			<Button
				intent="success"
				loading={saving}
				disabled={busy || !dirty || !!problem}
				onclick={save}
			>
				{saving ? 'Saving' : 'Save'}
			</Button>
		{:else}
			<Button variant="ghost" intent="secondary" onclick={() => (open = false)}>Close</Button>
		{/if}
	{/snippet}
</Drawer>

<Dialog bind:open={confirmOpen} title="Delete this transaction?" size="sm">
	<Text as="p" size="sm">
		{transaction.type} of {transaction.amount} on {formatDisplayDate(initial.date)} will be removed and
		the portfolio rebuilt without it. This cannot be undone.
	</Text>
	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (confirmOpen = false)}>Keep it</Button
		>
		<Button intent="error" onclick={confirmDelete}>Delete transaction</Button>
	{/snippet}
</Dialog>
