<script lang="ts">
	// Suggestions from your history. Nothing here counts until it is confirmed, so the
	// drawer sits beside the list it would add to rather than replacing it. Each card
	// quotes the statement text its name was read off, which is what makes the guess
	// checkable instead of something ta11y decided quietly.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { scaledToNumber } from '$lib/api/money';
	import { recurringDirectionLabel, rhythmLabels } from '$lib/api/recurring';
	import type { RecurringRhythm, RecurringSuggestion } from '$lib/api/types';

	type Props = {
		open?: boolean;
		suggestions: RecurringSuggestion[];
		loading?: boolean;
		error?: string | null;
		/** The match key currently being confirmed or dismissed, if any. */
		pending?: string | null;
		onConfirm?: (suggestion: RecurringSuggestion, name: string, rhythm: RecurringRhythm) => void;
		onDismiss?: (suggestion: RecurringSuggestion) => void;
	};

	let {
		open = $bindable(false),
		suggestions,
		loading = false,
		error = null,
		pending = null,
		onConfirm,
		onDismiss
	}: Props = $props();

	// Adjusting opens in place: changing a name should not cost you the list you were
	// reading, and the transactions behind the suggestion do not change either way.
	let adjusting = $state<string | null>(null);
	let draftName = $state('');
	// The Select atom speaks plain strings, so the rhythm narrows on its way back out.
	let draftRhythm = $state('monthly');

	const rhythmOptions = Object.entries(rhythmLabels).map(([value, label]) => ({ value, label }));

	function startAdjusting(suggestion: RecurringSuggestion) {
		adjusting = suggestion.match_key;
		draftName = suggestion.name;
		draftRhythm = suggestion.rhythm;
	}

	function confirm(suggestion: RecurringSuggestion) {
		const isAdjusted = adjusting === suggestion.match_key;
		onConfirm?.(
			suggestion,
			isAdjusted ? draftName : suggestion.name,
			isAdjusted ? (draftRhythm as RecurringRhythm) : suggestion.rhythm
		);
		adjusting = null;
	}
</script>

<Drawer bind:open title="Suggestions from your history" width="max-w-lg">
	<Text size="sm" tone="muted" class="mb-4">
		Patterns found in transactions you have not ignored. Nothing is added until you confirm it, and
		a dismissed suggestion does not come back.
	</Text>

	{#if loading}
		<div class="flex flex-col gap-3">
			{#each [0, 1, 2] as index (index)}
				<Panel variant="muted" shape="sm" padding="md" bordered class="flex flex-col gap-2">
					<Skeleton width="50%" />
					<Skeleton width="70%" />
				</Panel>
			{/each}
		</div>
	{:else if error}
		<p
			class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
			role="alert"
		>
			{error}
		</p>
	{:else if suggestions.length === 0}
		<p class="py-8 text-center text-sm text-slate-500">
			No patterns found in your history right now. Mark a transaction in Cashflow to start an item
			yourself.
		</p>
	{:else}
		<div class="flex flex-col gap-3">
			{#each suggestions as suggestion (suggestion.match_key)}
				{@const busy = pending === suggestion.match_key}
				<Panel
					variant="muted"
					shape="sm"
					padding="md"
					bordered
					class="flex flex-col gap-3 border-dashed"
				>
					<div class="flex items-start justify-between gap-3">
						<div class="min-w-0">
							<Heading level="h3" size="sm" class="truncate text-slate-900">
								{suggestion.name}
							</Heading>
							<Text as="span" size="xs" tone="muted">
								{recurringDirectionLabel(suggestion.direction)} · {suggestion.matches} transactions since
								{formatDisplayDate(suggestion.since.slice(0, 10))}
							</Text>
						</div>
						<Badge intent="neutral" variant="outline" size="sm">
							{rhythmLabels[suggestion.rhythm]}
						</Badge>
					</div>

					<div class="flex items-end justify-between gap-3">
						<Money
							amount={scaledToNumber(suggestion.amountCents)}
							currency="EUR"
							size="lg"
							weight="semibold"
						/>
						<span class="min-w-0 truncate text-xs text-slate-500">
							Named after “{suggestion.sample}”
						</span>
					</div>

					{#if adjusting === suggestion.match_key}
						<div class="flex flex-col gap-3 border-t border-slate-200 pt-3">
							<FormField
								label="To whom"
								id={`adjust-name-${suggestion.match_key}`}
								hint="Read off the statement text — change it to the name you recognise."
							>
								{#snippet children(field)}
									<Input id={field.id} bind:value={draftName} ariaDescribedby={field.describedby} />
								{/snippet}
							</FormField>
							<FormField label="How often" id={`adjust-rhythm-${suggestion.match_key}`}>
								{#snippet children(field)}
									<Select id={field.id} bind:value={draftRhythm} options={rhythmOptions} />
								{/snippet}
							</FormField>
						</div>
					{/if}

					<div class="flex flex-wrap items-center gap-2 border-t border-slate-200 pt-3">
						<Button
							size="sm"
							shape="default"
							loading={busy}
							disabled={busy}
							onclick={() => confirm(suggestion)}
						>
							Confirm
						</Button>
						{#if adjusting !== suggestion.match_key}
							<Button
								size="sm"
								shape="default"
								variant="outline"
								disabled={busy}
								onclick={() => startAdjusting(suggestion)}
							>
								Adjust name or rhythm
							</Button>
						{:else}
							<Button
								size="sm"
								shape="default"
								variant="outline"
								disabled={busy}
								onclick={() => (adjusting = null)}
							>
								Cancel changes
							</Button>
						{/if}
						<Button
							size="sm"
							shape="default"
							variant="ghost"
							intent="secondary"
							disabled={busy}
							onclick={() => onDismiss?.(suggestion)}
						>
							Dismiss
						</Button>
					</div>
				</Panel>
			{/each}
		</div>
	{/if}

	{#snippet footer()}
		<Button variant="outline" shape="default" onclick={() => (open = false)}>Close</Button>
	{/snippet}
</Drawer>
