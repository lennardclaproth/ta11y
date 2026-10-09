<script lang="ts">
	// Design #14 - final: adding an item to an asset class, in two steps.
	// Step 1 picks what you own from the instruments that have a daily price; step 2 records the
	// first purchase. Both steps stay in one dialog so "what" and "how much" are never separated,
	// and the same step 2 is reused on its own for "Add purchase" on an existing item.
	import { untrack } from 'svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { longDate, quoteResults } from '../fixtures';
	import type { DesignQuote } from '../fixtures';

	type Props = {
		open?: boolean;
		step?: 'quote' | 'purchase';
		/** Pre-filled filter, so a story can show "no matches". */
		quoteQuery?: string;
		/** A save that came back as a failure: the dialog keeps what was typed. */
		saveError?: boolean;
		/** A save in flight: the button is busy and cannot be pressed twice. */
		saving?: boolean;
	};

	let {
		open = $bindable(false),
		step = $bindable('quote'),
		quoteQuery = '',
		saveError = false,
		saving = false
	}: Props = $props();

	// The story seeds the filter once; typing takes over from there.
	let query = $state(untrack(() => quoteQuery));
	let itemKind = $state('priced');
	let pickedQuote = $state<DesignQuote>(quoteResults[0]);
	let purchaseDate = $state('2026-05-20');
	let quantity = $state('0.05');
	let unitPrice = $state('60000.00');

	const unit = $derived(pickedQuote.symbol.split('/')[0]);
	const matches = $derived(
		quoteResults.filter((q) =>
			`${q.symbol} ${q.name} ${q.kind}`.toLowerCase().includes(query.trim().toLowerCase())
		)
	);
	const paidPreview = $derived(
		(Number.parseFloat(quantity) || 0) * (Number.parseFloat(unitPrice) || 0)
	);
</script>

<Dialog bind:open title="Add item to Crypto & metals" size="md">
	<div class="space-y-4">
		{#if step === 'quote'}
			<div>
				<p class="mb-2 text-xs font-semibold tracking-wide text-slate-500 uppercase">
					Step 1 of 2 · What do you own?
				</p>
				<!-- The choice between the two kinds of item comes first: a manual item never reaches
				     step 2, it only needs a name and a worth. -->
				<Tabs
					tabs={[
						{ value: 'priced', label: 'Follows a daily price' },
						{ value: 'manual', label: 'Worth I set myself' }
					]}
					bind:value={itemKind}
					ariaLabel="How the item is valued"
				/>
			</div>

			{#if itemKind === 'priced'}
				<SearchInput
					bind:value={query}
					placeholder="Filter bitcoin, ethereum, gold…"
					ariaLabel="Filter the daily-priced instruments"
				/>

				<Text as="p" size="sm" tone="muted">
					Instruments with a daily price in euro. Anything quoted in another currency stays a
					manual item for now.
				</Text>

				{#if matches.length === 0}
					<!-- No matches is not an empty account: the filter is what is standing in the way, so
					     the way out is clearing it. -->
					<div class="rounded-md border border-slate-300 bg-taupe-50 px-4 py-8 text-center">
						<Text as="p" size="sm" tone="muted">
							No instrument matches “{query}”. Try the name or the symbol, for example “bitcoin” or
							“BTC”.
						</Text>
						<div class="mt-3">
							<Button
								size="sm"
								variant="outline"
								intent="secondary"
								shape="default"
								onclick={() => (query = '')}
							>
								Clear filter
							</Button>
						</div>
					</div>
				{:else}
					<ul class="divide-y divide-slate-200 rounded-md border border-slate-300">
						{#each matches as quote (quote.id)}
							<li>
								<button
									type="button"
									disabled={!quote.selectable}
									aria-current={pickedQuote.id === quote.id ? 'true' : undefined}
									class={[
										'flex w-full items-start justify-between gap-3 px-3 py-2.5 text-left transition-colors',
										quote.selectable ? 'hover:bg-slate-50' : 'cursor-not-allowed bg-taupe-50',
										pickedQuote.id === quote.id ? 'bg-amber-50' : '',
										'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none'
									]
										.filter(Boolean)
										.join(' ')}
									onclick={() => (pickedQuote = quote)}
								>
									<span class="min-w-0">
										<span class="flex flex-wrap items-center gap-2">
											<span class="text-sm font-medium text-slate-900">{quote.symbol}</span>
											<span class="truncate text-sm text-slate-600">{quote.name}</span>
											<Badge
												intent={quote.selectable ? 'info' : 'neutral'}
												variant="soft"
												size="sm"
											>
												{quote.currency}
											</Badge>
										</span>
										{#if quote.reason}
											<span class="mt-0.5 block text-xs text-slate-500">{quote.reason}</span>
										{/if}
									</span>
									<span class="shrink-0 text-right">
										<Money amount={quote.price} currency={quote.currency} size="sm" />
										<span class="mt-0.5 block text-xs text-slate-500">
											{longDate(quote.price_date)}
										</span>
									</span>
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			{:else}
				<FormField label="Name" id="add-manual-name">
					{#snippet children(ctx)}
						<Input id={ctx.id} value="" placeholder="e.g. Gold coins" />
					{/snippet}
				</FormField>
				<FormField label="Worth" id="add-manual-worth">
					{#snippet children(ctx)}
						<CurrencyInput id={ctx.id} value="" />
					{/snippet}
				</FormField>
			{/if}
		{:else}
			<div
				class="-mx-5 -mt-4 mb-1 flex flex-wrap items-end justify-between gap-x-4 gap-y-2 border-b border-slate-300 bg-taupe-50 px-5 py-3"
			>
				<div class="min-w-0">
					<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">
						Step 2 of 2 · First purchase
					</p>
					<p class="font-heading text-2xl leading-tight text-slate-900">{pickedQuote.name}</p>
					<Text as="span" size="sm" tone="muted">
						{pickedQuote.symbol} ·
						<Money amount={pickedQuote.price} currency="EUR" size="sm" />
						on {longDate(pickedQuote.price_date)}
					</Text>
				</div>
				<Button
					size="sm"
					variant="outline"
					intent="secondary"
					shape="default"
					onclick={() => (step = 'quote')}
				>
					Change
				</Button>
			</div>

			{#if saveError}
				<!-- What was typed stays on screen; the alert says what failed and the field says what
				     to change. -->
				<Alert intent="error" title="This purchase could not be saved">
					The quantity was not accepted. Nothing was recorded; correct it and save again.
				</Alert>
			{/if}

			<div class="grid gap-3 sm:grid-cols-3">
				<FormField label="Date" id="add-date">
					<DatePicker
						bind:value={purchaseDate}
						max="2026-06-17"
						portal={false}
						ariaLabel="Purchase date"
					/>
				</FormField>
				<FormField
					label="Quantity"
					id="add-qty"
					hint={`In ${unit}`}
					error={saveError ? 'Enter a quantity greater than zero.' : undefined}
				>
					{#snippet children(ctx)}
						<Input
							id={ctx.id}
							bind:value={quantity}
							ariaDescribedby={ctx.describedby}
							intent={saveError ? 'error' : 'default'}
							class="text-right tabular-nums"
						/>
					{/snippet}
				</FormField>
				<FormField label="Price paid per unit" id="add-price">
					{#snippet children(ctx)}
						<CurrencyInput id={ctx.id} bind:value={unitPrice} />
					{/snippet}
				</FormField>
			</div>

			<div class="flex items-baseline justify-between gap-3 border-t border-slate-300 pt-3">
				<Text as="span" size="sm" tone="muted">Paid</Text>
				<Money amount={paidPreview} currency="EUR" size="lg" weight="semibold" />
			</div>
			<Text as="p" size="sm" tone="muted">
				From this day on the value follows the daily price. You can add more purchases later.
			</Text>
		{/if}
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (open = false)}>Cancel</Button>
		{#if step === 'quote'}
			{#if itemKind === 'priced'}
				<Button disabled={matches.length === 0} onclick={() => (step = 'purchase')}>Next</Button>
			{:else}
				<Button intent="success" loading={saving}>Save</Button>
			{/if}
		{:else}
			<!-- `loading` keeps the button busy and unpressable, so one save never becomes two. -->
			<Button intent="success" loading={saving}>Save</Button>
		{/if}
	{/snippet}
</Dialog>
