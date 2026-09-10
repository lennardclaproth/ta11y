<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { ApiError } from '$lib/api/client';
	import type { Listing } from '$lib/api/types';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import {
		listingCurrencyOptions,
		listingMetadataFields,
		type ListingEditableKey,
		type ListingFieldPatch
	} from './listing-form.types';

	let {
		listing,
		onSave,
		onCancel,
		saving = $bindable(false)
	}: {
		listing: Listing;
		/** Receives only the fields that actually changed. */
		onSave: (patch: ListingFieldPatch) => Promise<void>;
		onCancel: () => void;
		saving?: boolean;
	} = $props();

	const editable = [...listingMetadataFields.map((field) => field.key), 'currency'] as const;

	/** The listing's current values, as form strings — null reads as an empty field. */
	function seed(from: Listing): Record<ListingEditableKey, string> {
		return Object.fromEntries(editable.map((key) => [key, from[key] ?? ''])) as Record<
			ListingEditableKey,
			string
		>;
	}

	// Seeded once, deliberately: the caller keys this form on the listing id, so a different
	// instrument gets a fresh component rather than one carrying the previous edits. `untrack`
	// says that the one-time read is the intent rather than a missed dependency.
	let values = $state(untrack(() => seed(listing)));
	let errors = $state<Record<string, string>>({});
	let failure = $state('');
	let form: HTMLFormElement;

	/**
	 * Only changed fields are sent. The endpoint patches what it is given and rejects an
	 * empty payload, so submitting an untouched form would be an error rather than a no-op.
	 */
	const patch = $derived.by(() => {
		const changed: ListingFieldPatch = {};
		for (const key of editable) {
			const next = values[key].trim();
			if (next !== (listing[key] ?? '')) changed[key] = next;
		}
		return changed;
	});

	const dirty = $derived(Object.keys(patch).length > 0);

	async function focusError() {
		await tick();
		form.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (saving || !dirty) return;
		errors = {};
		failure = '';
		saving = true;
		try {
			await onSave(patch);
		} catch (error) {
			if (error instanceof ApiError && error.status === 404) {
				failure = 'This listing no longer exists. Close this form and refresh the listings.';
			} else if (error instanceof ApiError && error.status === 400) {
				failure = 'The changes could not be saved. Check the details below and try again.';
				if (error.body && typeof error.body === 'object') {
					for (const [key, message] of Object.entries(error.body)) {
						if (key in values && typeof message === 'string') errors[key] = message;
					}
				}
			} else {
				failure =
					'Could not confirm the changes were saved. Check your connection and try again, then refresh the listings.';
			}
			await focusError();
		} finally {
			saving = false;
		}
	}
</script>

<form bind:this={form} onsubmit={submit} novalidate class="space-y-4" aria-busy={saving}>
	<!-- Symbol, name and source identify the instrument at the provider, so they are shown
	     rather than offered: changing one would repoint every price and position already
	     resolved through it. The API does not accept them either. -->
	<dl class="grid grid-cols-3 gap-3 rounded-lg border border-slate-200 bg-taupe-50 p-3 text-sm">
		<div class="min-w-0">
			<dt class="text-xs text-slate-500">Symbol</dt>
			<dd class="truncate font-medium text-slate-900">{listing.symbol}</dd>
		</div>
		<div class="min-w-0">
			<dt class="text-xs text-slate-500">Name</dt>
			<dd class="truncate text-slate-800">{listing.name}</dd>
		</div>
		<div class="min-w-0">
			<dt class="text-xs text-slate-500">Source</dt>
			<dd class="truncate text-slate-800">{listing.source}</dd>
		</div>
	</dl>

	<fieldset disabled={saving} class="grid gap-4 sm:grid-cols-2">
		<FormField label="Currency" id="edit-listing-currency" error={errors.currency}>
			{#snippet children(ctx)}
				<Select
					id={ctx.id}
					bind:value={values.currency}
					options={listingCurrencyOptions}
					ariaDescribedby={ctx.describedby}
					intent={ctx.invalid ? 'error' : 'default'}
				/>
			{/snippet}
		</FormField>
		{#each listingMetadataFields as field (field.key)}
			<FormField label={field.label} id={`edit-listing-${field.key}`} error={errors[field.key]}>
				{#snippet children(ctx)}
					<Input
						id={ctx.id}
						bind:value={values[field.key]}
						placeholder={field.placeholder}
						ariaDescribedby={ctx.describedby}
						intent={ctx.invalid ? 'error' : 'default'}
					/>
				{/snippet}
			</FormField>
		{/each}
	</fieldset>

	{#if failure}<p role="alert" class="text-red-700">{failure}</p>{/if}

	<div class="flex items-center justify-end gap-2 border-t border-slate-200 pt-4">
		<Button variant="ghost" intent="secondary" disabled={saving} onclick={onCancel}>Cancel</Button>
		<Button type="submit" disabled={saving || !dirty} loading={saving}>
			{saving ? 'Saving…' : 'Save changes'}
		</Button>
	</div>
</form>
