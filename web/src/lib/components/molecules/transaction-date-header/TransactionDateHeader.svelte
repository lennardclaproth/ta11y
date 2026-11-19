<script lang="ts">
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import {
		formatDisplayDate,
		relativeDayLabel,
		todayISO
	} from '$lib/components/molecules/calendar/calendar.utils';

	type Props = {
		/** Bindable transaction date ("YYYY-MM-DD"). */
		value: string;
		/** Latest selectable day; also the target of the "Today" shortcut. */
		today?: string;
		/** Bindable open state of the calendar, so a form can show it deliberately. */
		open?: boolean;
		disabled?: boolean;
		class?: string;
	};

	let {
		value = $bindable(),
		today = todayISO(),
		open = $bindable(false),
		disabled = false,
		class: className = ''
	}: Props = $props();

	const isToday = $derived(value === today);
</script>

<!-- The date is the headline of a transaction form, not a field among fields: it is the one
     value a backdated entry hinges on. The bleed matches the Dialog's own padding so the rule
     runs the full width of the form. -->
<div
	class={[
		'-mx-5 -mt-4 mb-1 flex flex-wrap items-end justify-between gap-x-4 gap-y-3',
		'border-b border-slate-300 bg-taupe-50 px-5 py-3',
		className
	]
		.filter(Boolean)
		.join(' ')}
>
	<div class="min-w-0">
		<p class="text-xs font-semibold tracking-wide text-slate-500 uppercase">Date</p>
		<p class="font-heading text-2xl leading-tight text-slate-900">{formatDisplayDate(value)}</p>
		<Text as="span" size="sm" tone="muted">
			{relativeDayLabel(value, today)} · today or earlier
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
		<!-- Inside a dialog the calendar must stay out of the portal, or the top layer hides it. -->
		<DatePicker
			{value}
			bind:open
			max={today}
			{disabled}
			portal={false}
			placement="bottom-end"
			size="sm"
			placeholder="Pick a day"
			ariaLabel="Transaction date"
			onChange={(next) => (value = next ?? value)}
		/>
	</div>
</div>
