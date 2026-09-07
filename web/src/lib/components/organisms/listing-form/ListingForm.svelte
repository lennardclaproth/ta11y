<script lang="ts">
	import { tick } from 'svelte';
	import type { CreateListingRequest } from '$lib/api/types';
	import { ApiError } from '$lib/api/client';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';

	let {
		onSave,
		onCancel,
		saving = $bindable(false)
	}: {
		onSave: (body: CreateListingRequest) => Promise<void>;
		onCancel: () => void;
		saving?: boolean;
	} = $props();
	let values = $state({
		symbol: '',
		name: '',
		source: '',
		currency: '',
		exchange: '',
		isin: '',
		ticker: '',
		region: '',
		type: '',
		description: ''
	});
	let errors = $state<Record<string, string>>({});
	let failure = $state('');
	let form: HTMLFormElement;
	const sources = [
		{ value: 'market_stack', label: 'Marketstack' },
		{ value: 'alpha_vantage', label: 'Alpha Vantage' },
		{ value: 'brandnewday', label: 'Brand New Day (manual uploads)' }
	];
	const metadata = [
		{ key: 'exchange', label: 'Exchange', placeholder: 'e.g. XAMS' },
		{ key: 'isin', label: 'ISIN', placeholder: 'e.g. IE00B3RBWM25' },
		{ key: 'ticker', label: 'Ticker', placeholder: 'e.g. VWRL' },
		{ key: 'region', label: 'Region', placeholder: 'e.g. Netherlands' },
		{ key: 'type', label: 'Type', placeholder: 'e.g. ETF' },
		{ key: 'description', label: 'Description', placeholder: 'Additional information' }
	] as const;

	async function focusError() {
		await tick();
		form.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus();
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (saving) return;
		errors = {};
		failure = '';
		for (const key of ['symbol', 'name', 'source'] as const) {
			if (!values[key].trim())
				errors[key] = `Enter ${key === 'source' ? 'a data source' : `a ${key}`}.`;
		}
		if (Object.keys(errors).length) {
			await focusError();
			return;
		}
		const body: CreateListingRequest = {
			symbol: values.symbol.trim().toUpperCase(),
			name: values.name.trim(),
			source: values.source
		};
		for (const key of ['currency', ...metadata.map((field) => field.key)] as const) {
			const value = values[key].trim();
			if (value) body[key] = key === 'isin' ? value.toUpperCase() : value;
		}
		saving = true;
		try {
			await onSave(body);
		} catch (error) {
			if (error instanceof ApiError && error.status === 409) {
				failure =
					'A listing with this symbol and source already exists. Check the listings or use a different symbol or source.';
			} else if (error instanceof ApiError && error.status === 400) {
				failure = 'The listing could not be added. Check the details below and try again.';
				if (error.body && typeof error.body === 'object') {
					for (const [key, message] of Object.entries(error.body)) {
						if (key in values && typeof message === 'string') errors[key] = message;
					}
				}
			} else {
				failure =
					'Could not confirm that the listing was saved. Check your connection and try again. If it already exists, close this form and refresh the listings.';
			}
			await focusError();
		} finally {
			saving = false;
		}
	}
</script>

<form bind:this={form} onsubmit={submit} novalidate class="space-y-4" aria-busy={saving}>
	<p class="text-slate-600">
		Add an instrument for portfolio transactions and price history. Fields marked * are required.
	</p>
	<fieldset disabled={saving} class="space-y-4">
		{#each [{ key: 'symbol', label: 'Symbol', placeholder: 'e.g. VWRL' }, { key: 'name', label: 'Name', placeholder: 'e.g. Vanguard FTSE All-World' }] as field (field.key)}
			<FormField label={field.label} id={`listing-${field.key}`} required error={errors[field.key]}>
				{#snippet children(ctx)}
					<Input
						id={ctx.id}
						bind:value={values[field.key as 'symbol' | 'name']}
						required
						placeholder={field.placeholder}
						ariaDescribedby={ctx.describedby}
						intent={ctx.invalid ? 'error' : 'default'}
					/>
				{/snippet}
			</FormField>
		{/each}
		<FormField
			label="Data source"
			id="listing-source"
			required
			error={errors.source}
			hint={values.source === 'brandnewday'
				? 'Prices are supplied through manual daily uploads.'
				: 'Use the symbol recognized by this provider. Automatic prices require a configured provider.'}
		>
			{#snippet children(ctx)}
				<Select
					id={ctx.id}
					bind:value={values.source}
					options={sources}
					placeholder="Choose a data source"
					required
					ariaDescribedby={ctx.describedby}
					intent={ctx.invalid ? 'error' : 'default'}
				/>
			{/snippet}
		</FormField>
		<details
			class="border-t border-slate-200 pt-4"
			open={metadata.some((field) => errors[field.key]) || !!errors.currency}
		>
			<summary
				class="cursor-pointer font-medium focus-visible:outline-2 focus-visible:outline-offset-4"
				>Optional details</summary
			>
			<div class="mt-4 grid gap-4 sm:grid-cols-2">
				<FormField label="Currency" id="listing-currency" error={errors.currency}>
					{#snippet children(ctx)}
						<Select
							id={ctx.id}
							bind:value={values.currency}
							options={[
								{ value: '', label: 'Not specified' },
								...['EUR', 'USD', 'GBP', 'JPY'].map((value) => ({ value, label: value }))
							]}
							ariaDescribedby={ctx.describedby}
							intent={ctx.invalid ? 'error' : 'default'}
						/>
					{/snippet}
				</FormField>
				{#each metadata as field (field.key)}
					<FormField label={field.label} id={`listing-${field.key}`} error={errors[field.key]}>
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
			</div>
		</details>
	</fieldset>
	{#if failure}<p role="alert" class="text-red-700">{failure}</p>{/if}
	<div class="flex justify-end gap-2 border-t border-slate-200 pt-4">
		<Button variant="ghost" intent="secondary" disabled={saving} onclick={onCancel}>Cancel</Button>
		<Button type="submit" disabled={saving} loading={saving}
			>{saving ? 'Adding listing…' : 'Add listing'}</Button
		>
	</div>
</form>
