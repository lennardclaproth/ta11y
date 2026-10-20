<script lang="ts">
	// Adding an item to an asset class, in two steps. Step 1 picks how the item is valued and, for
	// a daily-priced one, what you own; step 2 records the first purchase. Both steps stay in one
	// dialog so "what" and "how much" are never separated. With `holding` set the dialog skips to
	// step 2 on its own, which is "Add purchase" on an item that already exists.
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
	import { decimalStringToNumber } from '$lib/api/money';
	import { formatDisplayDate, todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import { addPurchase, createAsset, createHolding, updatePurchase } from '$lib/services/assets';
	import { searchQuotes } from '$lib/services/marketdata';
	import type { AssetHolding, HoldingPurchase, Quote } from '$lib/api/types';
	import type { AddItemStep } from './add-asset-item-dialog.types';

	type Props = {
		open?: boolean;
		classId: string;
		/** The name of the asset class; `className` is the Tailwind passthrough elsewhere. */
		assetClassName: string;
		/**
		 * Set to add another purchase to an item that already follows a daily price. The
		 * instrument is then fixed and only step 2 is shown.
		 */
		holding?: AssetHolding | null;
		/** With `holding`, a recorded purchase of it to correct instead of adding a new one. */
		purchase?: HoldingPurchase | null;
		/** Fired after a successful save, so the caller can reload what it shows. */
		onSaved?: () => void;
	};

	let {
		open = $bindable(false),
		classId,
		assetClassName,
		holding = null,
		purchase = null,
		onSaved
	}: Props = $props();

	let step = $state<AddItemStep>('quote');
	let itemKind = $state('priced');
	let query = $state('');
	let quotes = $state<Quote[]>([]);
	let quotesLoading = $state(false);
	let quotesError = $state<string | null>(null);
	let picked = $state<Quote | null>(null);

	let manualName = $state('');
	let manualWorth = $state('');
	/** The day the worth applies from, so an item held for a while can start in the past. */
	let manualDate = $state<string | null>(todayISO());

	let purchaseDate = $state<string | null>(todayISO());
	let quantity = $state('');
	/** Whether `price` is what the whole purchase cost or what one unit cost. */
	let priceMode = $state('total');
	let price = $state('');

	let saving = $state(false);
	let saveError = $state<string | null>(null);
	let fieldErrors = $state<Record<string, string>>({});

	const today = todayISO();
	const instrument = $derived(picked ?? quoteFromHolding(holding));
	const unit = $derived(instrument ? instrument.symbol.split('/')[0] : '');
	// The API stores the price per unit; a total is divided by the quantity before it is sent.
	const quantityValue = $derived(Number.parseFloat(quantity) || 0);
	const priceValue = $derived(Number.parseFloat(price) || 0);
	const unitPrice = $derived(
		priceMode === 'unit' ? priceValue : quantityValue > 0 ? priceValue / quantityValue : 0
	);
	const paid = $derived(priceMode === 'unit' ? quantityValue * priceValue : priceValue);

	/** An existing holding stands in for a picked quote, so step 2 reads the same either way. */
	function quoteFromHolding(value: AssetHolding | null): Quote | null {
		if (!value) return null;
		return {
			symbol: value.symbol,
			name: value.name,
			kind: 'CRYPTO',
			currency: 'EUR',
			price: value.price,
			price_date: value.price_date,
			selectable: true
		};
	}

	function reset() {
		step = holding ? 'purchase' : 'quote';
		itemKind = 'priced';
		query = '';
		picked = null;
		manualName = '';
		manualWorth = '';
		manualDate = today;
		if (holding && purchase) {
			// A correction starts from what was recorded, which is a price per unit.
			purchaseDate = purchase.date;
			quantity = purchase.quantity;
			priceMode = 'unit';
			price = String(decimalStringToNumber(purchase.unit_price));
		} else {
			purchaseDate = today;
			quantity = '';
			priceMode = 'total';
			price = '';
		}
		saveError = null;
		fieldErrors = {};
	}

	async function loadQuotes() {
		// "Clear filter" calls this straight away and can overtake a request the debounced
		// typing already sent, leaving a list — and a pick derived from it — that does not
		// belong to the filter on screen.
		const wanted = query;
		const current = () => untrack(() => query) === wanted;
		quotesLoading = true;
		quotesError = null;
		try {
			const loaded = await searchQuotes({ q: wanted || undefined });
			if (!current()) return;
			quotes = loaded;
			// Keep a pick that is still in the result set; drop one the filter removed.
			if (picked && !quotes.some((q) => q.symbol === picked?.symbol)) picked = null;
			if (!picked) picked = quotes.find((q) => q.selectable) ?? null;
		} catch {
			if (!current()) return;
			quotesError = 'The instruments could not be loaded.';
			quotes = [];
		} finally {
			if (current()) quotesLoading = false;
		}
	}

	// Reopening starts over: a half-filled purchase from last time would be saved by accident.
	// Only `open` drives this; `loadQuotes` reads the filter, and tracking it would reset the
	// dialog on every keystroke.
	$effect(() => {
		if (open) {
			untrack(() => {
				reset();
				if (!holding) void loadQuotes();
			});
		}
	});

	function problemsFor(error: unknown): Record<string, string> {
		const body = error instanceof Error && 'body' in error ? error.body : null;
		if (!body || typeof body !== 'object') return {};
		return Object.fromEntries(
			Object.entries(body as Record<string, unknown>)
				.filter(([, value]) => typeof value === 'string')
				.map(([key, value]) => [key.replace('purchase.', ''), value as string])
		);
	}

	function validatePurchase(): boolean {
		const problems: Record<string, string> = {};
		if (!purchaseDate) problems.date = 'Pick the day you bought it.';
		const qty = Number.parseFloat(quantity);
		if (!Number.isFinite(qty) || qty <= 0) problems.quantity = 'Enter a quantity greater than zero.';
		const amount = Number.parseFloat(price);
		if (!Number.isFinite(amount) || amount < 0) problems.unit_price = 'Enter the price you paid.';
		fieldErrors = problems;
		return Object.keys(problems).length === 0;
	}

	async function saveManual() {
		const problems: Record<string, string> = {};
		if (manualName.trim() === '') problems.name = 'Give the item a name.';
		if (!manualDate) problems.effective_date = 'Pick the day this worth applies from.';
		fieldErrors = problems;
		if (Object.keys(problems).length > 0) return;
		saving = true;
		saveError = null;
		try {
			await createAsset({
				class_id: classId,
				name: manualName.trim(),
				initial_worth: (Number.parseFloat(manualWorth) || 0).toFixed(6),
				effective_date: manualDate as string
			});
			open = false;
			onSaved?.();
		} catch (error) {
			fieldErrors = problemsFor(error);
			saveError = 'This item could not be saved. Nothing was recorded.';
		} finally {
			saving = false;
		}
	}

	async function savePurchase() {
		if (!validatePurchase()) return;
		saving = true;
		saveError = null;
		try {
			const body = {
				date: purchaseDate as string,
				quantity: quantity.trim(),
				unit_price: unitPrice.toFixed(6)
			};
			if (holding && purchase) {
				await updatePurchase(holding.id, purchase.id, body);
			} else if (holding) {
				await addPurchase(holding.id, body);
			} else {
				if (!picked) return;
				await createHolding({
					class_id: classId,
					name: picked.name,
					symbol: picked.symbol,
					purchase: body
				});
			}
			open = false;
			onSaved?.();
		} catch (error) {
			fieldErrors = problemsFor(error);
			// A refusal that names no field (wrong class, untrackable instrument) is only visible here.
			const reason = fieldErrors.error ?? fieldErrors.purchase ?? fieldErrors.symbol;
			saveError = reason
				? `Nothing was recorded: ${reason}.`
				: 'This purchase could not be saved. Nothing was recorded; correct it and try again.';
		} finally {
			saving = false;
		}
	}
</script>

<Dialog
	bind:open
	title={holding
		? purchase
			? `Edit purchase of ${holding.name}`
			: `Add purchase to ${holding.name}`
		: `Add item to ${assetClassName}`}
	size="md"
>
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
					onSearch={() => void loadQuotes()}
				/>

				<Text as="p" size="sm" tone="muted">
					Instruments with a daily price in euro. Anything quoted in another currency stays a manual
					item for now.
				</Text>

				{#if quotesError}
					<Alert intent="error" title="The instruments could not be loaded">
						{quotesError}
					</Alert>
				{:else if quotesLoading}
					<p class="px-4 py-8 text-center text-sm text-slate-500" aria-busy="true">
						Loading instruments…
					</p>
				{:else if quotes.length === 0}
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
								onclick={() => {
									query = '';
									void loadQuotes();
								}}
							>
								Clear filter
							</Button>
						</div>
					</div>
				{:else}
					<ul class="divide-y divide-slate-200 rounded-md border border-slate-300">
						{#each quotes as quote (quote.symbol)}
							<li>
								<button
									type="button"
									disabled={!quote.selectable}
									aria-current={picked?.symbol === quote.symbol ? 'true' : undefined}
									class={[
										'flex w-full items-start justify-between gap-3 px-3 py-2.5 text-left transition-colors',
										quote.selectable ? 'hover:bg-slate-50' : 'cursor-not-allowed bg-taupe-50',
										picked?.symbol === quote.symbol ? 'bg-amber-50' : '',
										'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none'
									]
										.filter(Boolean)
										.join(' ')}
									onclick={() => (picked = quote)}
								>
									<span class="min-w-0">
										<span class="flex flex-wrap items-center gap-2">
											<span class="text-sm font-medium text-slate-900">{quote.symbol}</span>
											<span class="truncate text-sm text-slate-600">{quote.name}</span>
											<Badge intent={quote.selectable ? 'info' : 'neutral'} variant="soft" size="sm">
												{quote.currency}
											</Badge>
										</span>
										{#if quote.reason}
											<span class="mt-0.5 block text-xs text-slate-500">{quote.reason}</span>
										{/if}
									</span>
									{#if quote.price && quote.price_date}
										<span class="shrink-0 text-right">
											<Money
												amount={decimalStringToNumber(quote.price)}
												currency={quote.currency}
												size="sm"
											/>
											<span class="mt-0.5 block text-xs text-slate-500">
												{formatDisplayDate(quote.price_date)}
											</span>
										</span>
									{:else}
										<span class="shrink-0 text-right text-xs text-slate-500">
											Not tracked yet
										</span>
									{/if}
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			{:else}
				{#if saveError}
					<Alert intent="error" title="This item could not be saved">{saveError}</Alert>
				{/if}
				<FormField label="Name" id="add-manual-name" error={fieldErrors.name}>
					{#snippet children(ctx)}
						<Input
							id={ctx.id}
							bind:value={manualName}
							ariaDescribedby={ctx.describedby}
							intent={ctx.invalid ? 'error' : 'default'}
							placeholder="e.g. Gold coins"
						/>
					{/snippet}
				</FormField>
				<div class="grid gap-3 sm:grid-cols-2">
					<FormField label="Worth" id="add-manual-worth" error={fieldErrors.initial_worth}>
						{#snippet children(ctx)}
							<CurrencyInput
								id={ctx.id}
								bind:value={manualWorth}
								ariaDescribedby={ctx.describedby}
								intent={ctx.invalid ? 'error' : 'default'}
							/>
						{/snippet}
					</FormField>
					<FormField label="As of" id="add-manual-date" error={fieldErrors.effective_date}>
						{#snippet children(ctx)}
							<DatePicker
								id={ctx.id}
								bind:value={manualDate}
								max={today}
								portal={false}
								ariaLabel="Date the worth applies from"
								ariaDescribedby={ctx.describedby}
								invalid={ctx.invalid}
							/>
						{/snippet}
					</FormField>
				</div>
			{/if}
		{:else if instrument}
			<div
				class="-mx-5 -mt-4 mb-1 flex flex-wrap items-end justify-between gap-x-4 gap-y-2 border-b border-slate-300 bg-taupe-50 px-5 py-3"
			>
				<div class="min-w-0">
					<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">
						{holding
							? purchase
								? 'Recorded purchase'
								: 'New purchase'
							: 'Step 2 of 2 · First purchase'}
					</p>
					<p class="font-heading text-2xl leading-tight text-slate-900">{instrument.name}</p>
					<Text as="span" size="sm" tone="muted">
						{instrument.symbol}{#if instrument.price && instrument.price_date}
							· <Money
								amount={decimalStringToNumber(instrument.price)}
								currency="EUR"
								size="sm"
							/>
							on {formatDisplayDate(instrument.price_date)}
						{/if}
					</Text>
				</div>
				{#if !holding}
					<Button
						size="sm"
						variant="outline"
						intent="secondary"
						shape="default"
						onclick={() => (step = 'quote')}
					>
						Change
					</Button>
				{/if}
			</div>

			{#if saveError}
				<!-- What was typed stays on screen; the alert says what failed and the field says what
				     to change. -->
				<Alert intent="error" title="This purchase could not be saved">{saveError}</Alert>
			{/if}

			<div class="grid gap-3 sm:grid-cols-3">
				<FormField label="Date" id="add-date" error={fieldErrors.date}>
					{#snippet children(ctx)}
						<DatePicker
							id={ctx.id}
							bind:value={purchaseDate}
							max={today}
							portal={false}
							ariaLabel="Purchase date"
							ariaDescribedby={ctx.describedby}
							invalid={ctx.invalid}
						/>
					{/snippet}
				</FormField>
				<FormField
					label="Quantity"
					id="add-qty"
					hint={unit ? `In ${unit}` : undefined}
					error={fieldErrors.quantity}
				>
					{#snippet children(ctx)}
						<Input
							id={ctx.id}
							bind:value={quantity}
							ariaDescribedby={ctx.describedby}
							intent={ctx.invalid ? 'error' : 'default'}
							class="text-right tabular-nums"
						/>
					{/snippet}
				</FormField>
				<FormField
					label={priceMode === 'unit' ? 'Price paid per unit' : 'Total paid'}
					id="add-price"
					error={fieldErrors.unit_price}
				>
					{#snippet children(ctx)}
						<CurrencyInput
							id={ctx.id}
							bind:value={price}
							ariaDescribedby={ctx.describedby}
							intent={ctx.invalid ? 'error' : 'default'}
						/>
					{/snippet}
				</FormField>
			</div>

			<Tabs
				tabs={[
					{ value: 'total', label: 'Total paid' },
					{ value: 'unit', label: 'Price per unit' }
				]}
				bind:value={priceMode}
				size="sm"
				ariaLabel="What the price is for"
			/>

			<!-- The preview shows the figure that was not typed, so both are checked before saving. -->
			<div class="flex items-baseline justify-between gap-3 border-t border-slate-300 pt-3">
				{#if priceMode === 'unit'}
					<Text as="span" size="sm" tone="muted">Paid</Text>
					<Money amount={paid} currency="EUR" size="lg" weight="semibold" />
				{:else}
					<Text as="span" size="sm" tone="muted">Price per {unit || 'unit'}</Text>
					<Money amount={unitPrice} currency="EUR" size="lg" weight="semibold" />
				{/if}
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
				<Button disabled={!picked} onclick={() => (step = 'purchase')}>Next</Button>
			{:else}
				<Button intent="success" loading={saving} onclick={saveManual}>Save</Button>
			{/if}
		{:else}
			<!-- `loading` keeps the button busy and unpressable, so one save never becomes two. -->
			<Button intent="success" loading={saving} onclick={savePurchase}>Save</Button>
		{/if}
	{/snippet}
</Dialog>
