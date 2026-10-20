<script lang="ts">
	// Changing the worth of an item whose worth is set by hand. Money moving in or out and the
	// item growing on its own are different things — one changes what was paid in, the other is
	// growth — so the dialog asks which one it is before it asks for an amount.
	import { untrack } from 'svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { decimalStringToNumber } from '$lib/api/money';
	import { todayISO } from '$lib/components/molecules/calendar/calendar.utils';
	import { adjustAssetWorth, setAssetWorth, updateMutation } from '$lib/services/assets';
	import type { AssetMutation } from '$lib/api/types';
	import type { WorthDialogAsset } from './asset-worth-dialog.types';

	type Props = {
		open?: boolean;
		asset: WorthDialogAsset | null;
		/**
		 * A recorded change to correct instead of a new one. It keeps its kind, so the choice
		 * between adding and setting is not offered; the history after it is recalculated.
		 */
		editing?: AssetMutation | null;
		/** The change being edited is the item's opening worth. */
		opening?: boolean;
		/** Fired after a successful save, so the caller can reload what it shows. */
		onSaved?: () => void;
	};

	let {
		open = $bindable(false),
		asset,
		editing = null,
		opening = false,
		onSaved
	}: Props = $props();

	// Plain strings because Tabs binds one; the values are `WorthChangeMode` and the direction.
	let mode = $state<string>('adjust');
	let direction = $state<string>('INCREASE');
	let amount = $state('');
	let date = $state<string | null>(todayISO());
	let note = $state('');

	let saving = $state(false);
	let saveError = $state<string | null>(null);
	let fieldErrors = $state<Record<string, string>>({});

	const today = todayISO();
	const dialogTitle = $derived(
		editing ? 'Edit change' : asset ? `Update ${asset.name}` : 'Update worth'
	);
	const current = $derived(asset ? decimalStringToNumber(asset.current_worth) : 0);
	const amountValue = $derived(Number.parseFloat(amount) || 0);
	const newWorth = $derived(
		mode === 'set' ? amountValue : direction === 'DECREASE' ? current - amountValue : current + amountValue
	);

	$effect(() => {
		if (open) {
			untrack(() => {
				if (editing) {
					mode = editing.change_type.toUpperCase() === 'ADJUST' ? 'adjust' : 'set';
					direction = editing.direction?.toUpperCase() === 'DECREASE' ? 'DECREASE' : 'INCREASE';
					amount = String(decimalStringToNumber(editing.amount));
					date = editing.effective_date.slice(0, 10);
					note = editing.note ?? '';
				} else {
					mode = 'adjust';
					direction = 'INCREASE';
					amount = '';
					date = today;
					note = '';
				}
				saveError = null;
				fieldErrors = {};
			});
		}
	});

	function validate(): boolean {
		const problems: Record<string, string> = {};
		const value = Number.parseFloat(amount);
		if (!Number.isFinite(value) || value < 0 || (mode === 'adjust' && value === 0)) {
			problems.amount = mode === 'set' ? 'Enter what it is worth now.' : 'Enter an amount.';
		} else if (!editing && mode === 'adjust' && direction === 'DECREASE' && value > current) {
			// Only checked for a new withdrawal: an edited one is weighed against the worth on its
			// own day, which only the recalculated history knows.
			problems.amount = 'You cannot withdraw more than it is worth.';
		}
		if (!date) problems.effective_date = 'Pick the day of the change.';
		fieldErrors = problems;
		return Object.keys(problems).length === 0;
	}

	async function save() {
		if (!asset || !validate()) return;
		saving = true;
		saveError = null;
		const effective_date = date as string;
		const trimmedNote = note.trim() || undefined;
		try {
			if (editing) {
				await updateMutation(asset.id, editing.id, {
					amount: amountValue.toFixed(6),
					direction:
						mode === 'adjust' ? (direction === 'DECREASE' ? 'DECREASE' : 'INCREASE') : undefined,
					effective_date,
					note: trimmedNote
				});
			} else if (mode === 'set') {
				await setAssetWorth(asset.id, {
					worth: amountValue.toFixed(6),
					effective_date,
					note: trimmedNote
				});
			} else {
				await adjustAssetWorth(asset.id, {
					direction: direction === 'DECREASE' ? 'DECREASE' : 'INCREASE',
					amount: amountValue.toFixed(6),
					effective_date,
					note: trimmedNote
				});
			}
			open = false;
			onSaved?.();
		} catch (error) {
			const body = error instanceof Error && 'body' in error ? error.body : null;
			const problems =
				body && typeof body === 'object'
					? (Object.fromEntries(
							Object.entries(body).filter(([, value]) => typeof value === 'string')
						) as Record<string, string>)
					: {};
			// The adjust endpoint names its amount problem "worth".
			fieldErrors = { ...problems, amount: problems.amount ?? problems.worth };
			const reason = problems.error;
			saveError = reason
				? `Nothing was recorded: ${reason}.`
				: 'The change could not be saved. Nothing was recorded; correct it and try again.';
		} finally {
			saving = false;
		}
	}
</script>

<Dialog bind:open title={dialogTitle} size="md">
	<div class="space-y-4">
		{#if !editing}
			<Tabs
				tabs={[
					{ value: 'adjust', label: 'Add or withdraw' },
					{ value: 'set', label: 'Set new worth' }
				]}
				bind:value={mode}
				ariaLabel="What changed"
			/>
		{/if}
		<Text as="p" size="sm" tone="muted">
			{#if editing}
				The history after this change is recalculated, so later changes and your net worth
				follow the correction.
			{:else if mode === 'adjust'}
				Money you put in or took out, like a transfer to your savings. It counts towards paid in.
			{:else}
				What it is worth now on its own, like savings after interest or a revalued house. The
				difference counts as growth.
			{/if}
		</Text>

		{#if saveError}
			<Alert intent="error" title="This change could not be saved">{saveError}</Alert>
		{/if}

		{#if mode === 'adjust'}
			<Tabs
				tabs={[
					{ value: 'INCREASE', label: 'Add' },
					{ value: 'DECREASE', label: 'Withdraw' }
				]}
				bind:value={direction}
				size="sm"
				ariaLabel="Added or withdrawn"
			/>
		{/if}

		<div class="grid gap-3 sm:grid-cols-2">
			<FormField
				label={opening
					? 'Opening worth'
					: mode === 'set'
						? editing
							? 'Worth set to'
							: 'Worth now'
						: direction === 'DECREASE'
							? 'Withdrawn'
							: 'Added'}
				id="worth-amount"
				error={fieldErrors.amount}
			>
				{#snippet children(ctx)}
					<CurrencyInput
						id={ctx.id}
						bind:value={amount}
						ariaDescribedby={ctx.describedby}
						intent={ctx.invalid ? 'error' : 'default'}
					/>
				{/snippet}
			</FormField>
			<FormField label="Date" id="worth-date" error={fieldErrors.effective_date}>
				{#snippet children(ctx)}
					<DatePicker
						id={ctx.id}
						bind:value={date}
						max={today}
						portal={false}
						ariaLabel="Date of the change"
						ariaDescribedby={ctx.describedby}
						invalid={ctx.invalid}
					/>
				{/snippet}
			</FormField>
		</div>

		<FormField label="Note" id="worth-note" hint="Optional">
			{#snippet children(ctx)}
				<Input
					id={ctx.id}
					bind:value={note}
					ariaDescribedby={ctx.describedby}
					placeholder="e.g. Monthly transfer"
				/>
			{/snippet}
		</FormField>

		{#if !editing}
			<div class="flex items-baseline justify-between gap-3 border-t border-slate-300 pt-3">
				<Text as="span" size="sm" tone="muted">
					Worth after this change (now <Money amount={current} currency="EUR" size="sm" />)
				</Text>
				<Money amount={newWorth} currency="EUR" size="lg" weight="semibold" />
			</div>
		{/if}
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (open = false)}>Cancel</Button>
		<Button intent="success" loading={saving} onclick={save}>Save</Button>
	{/snippet}
</Dialog>
