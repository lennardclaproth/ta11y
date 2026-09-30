<script lang="ts">
	// Proposal (new molecule, `transaction-date-header`): the ruled date line at the top of the
	// "New transaction" form. The date is the headline of the form, not a field among fields:
	// label, serif date, a plain-language distance, and the controls on the right.
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import InFormDatePicker from '../shared/InFormDatePicker.svelte';
	import { daysAgoLabel } from '../shared/design-data';

	type Props = {
		value: string;
		today: string;
		open?: boolean;
		disabled?: boolean;
	};

	let {
		value = $bindable(),
		today,
		open = $bindable(false),
		disabled = false
	}: Props = $props();

	const isToday = $derived(value === today);
</script>

<div
	class="-mx-5 -mt-4 mb-1 flex flex-wrap items-end justify-between gap-x-4 gap-y-3 border-b border-slate-300 bg-taupe-50 px-5 py-3"
>
	<div class="min-w-0">
		<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">Date</p>
		<p class="font-heading text-2xl leading-tight text-slate-900">
			{formatDisplayDate(value)}
		</p>
		<Text as="span" size="sm" tone="muted">
			{daysAgoLabel(value, today)} · today or earlier
		</Text>
	</div>
	<div class="flex shrink-0 items-center gap-2">
		<Button
			size="sm"
			variant="outline"
			intent="secondary"
			shape="default"
			disabled={disabled || isToday}
			ariaLabel="Set the date to today"
			onclick={() => (value = today)}
		>
			Today
		</Button>
		<InFormDatePicker
			bind:value
			bind:open
			max={today}
			{disabled}
			size="sm"
			placeholder="Pick a day"
			ariaLabel="Transaction date"
		/>
	</div>
</div>
