<script lang="ts">
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Popover from '$lib/components/molecules/popover/Popover.svelte';
	import Calendar from '$lib/components/molecules/calendar/Calendar.svelte';
	import type { PopoverPlacement } from '$lib/components/molecules/popover/popover.types';
	import {
		formatDisplayDate,
		parseISODate,
		startOfMonthUTC
	} from '$lib/components/molecules/calendar/calendar.utils';

	type Size = 'sm' | 'md' | 'lg';

	type Props = {
		/** Bindable selected date ("YYYY-MM-DD") or null. */
		value?: string | null;
		placeholder?: string;
		min?: string | null;
		max?: string | null;
		size?: Size;
		disabled?: boolean;
		/**
		 * Append the calendar to `<body>`. Leave it on to escape clipping ancestors; turn it
		 * **off** inside a `<dialog>`, whose top layer would otherwise cover the panel and make
		 * the date unpickable.
		 */
		portal?: boolean;
		/** Where the calendar hangs off the trigger; `bottom-end` keeps it inside a narrow dialog. */
		placement?: PopoverPlacement;
		/** Bindable open state of the calendar. */
		open?: boolean;
		ariaLabel?: string;
		onChange?: (value: string | null) => void;
		class?: string;
	};

	let {
		value = $bindable(null),
		placeholder = 'Select date',
		min = null,
		max = null,
		size = 'md',
		disabled = false,
		portal = true,
		placement = 'bottom-start',
		open = $bindable(false),
		ariaLabel = 'Select date',
		onChange,
		class: className = ''
	}: Props = $props();

	let month = $state(startOfMonthUTC(parseISODate(value) ?? new Date()));

	// Re-center the calendar on the selected value whenever the popover opens.
	$effect(() => {
		if (open) month = startOfMonthUTC(parseISODate(value) ?? new Date());
	});

	const sizeClasses = {
		sm: 'h-8 px-2.5 text-sm',
		md: 'h-10 px-3 text-sm',
		lg: 'h-12 px-3 text-base'
	} satisfies Record<Size, string>;

	const label = $derived(value ? formatDisplayDate(value) : placeholder);

	function handleSelect(iso: string) {
		value = iso;
		onChange?.(iso);
		open = false;
	}
</script>

<Popover bind:open {placement} {portal} class="p-3">
	{#snippet trigger(api)}
		<button
			type="button"
			{disabled}
			aria-label={ariaLabel}
			aria-expanded={api.open}
			class={[
				'inline-flex items-center gap-2 rounded-xl border bg-white whitespace-nowrap',
				'transition-colors hover:bg-slate-50',
				'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
				'disabled:pointer-events-none disabled:opacity-50',
				api.open ? 'border-slate-400' : 'border-slate-300',
				value ? 'text-slate-800' : 'text-slate-500',
				sizeClasses[size],
				className
			].join(' ')}
			onclick={api.toggle}
		>
			<Icon icon="heroicons:calendar-days" size="sm" class="text-slate-500" />
			<span>{label}</span>
		</button>
	{/snippet}

	<Calendar bind:month mode="single" selected={value} {min} {max} onSelect={handleSelect} />
</Popover>
