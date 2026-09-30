<script lang="ts">
	import type { IconifyIcon } from '@iconify/svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { pressClasses } from '../shared/press';

	/**
	 * Proposed Button variant `ruled`: a supporting action that hangs on a rule instead of a
	 * filled or outlined box, so the one dominant action per region stays the only filled button.
	 *
	 * Hover and press are the same as the existing `primary/ghost` button — amber-100 on hover,
	 * amber-200 while pressed — plus the shared black bar and amber ring, so every action in the
	 * design answers a click the same way.
	 */
	type Props = {
		label: string;
		icon?: string | IconifyIcon;
		disabled?: boolean;
		onclick?: (event: MouseEvent) => void;
	};

	let { label, icon, disabled = false, onclick }: Props = $props();
</script>

<button
	type="button"
	{disabled}
	{onclick}
	class="inline-flex h-10 items-center gap-1.5 rounded-t-sm border-b border-slate-400 px-1.5 text-sm
	       text-slate-700 transition-all duration-150 ease-out
	       hover:border-slate-700 hover:bg-amber-100 hover:text-slate-950
	       focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none
	       active:border-slate-900 active:bg-amber-200 active:text-slate-950
	       disabled:pointer-events-none disabled:opacity-50 {pressClasses}"
>
	{#if icon}<Icon {icon} size="sm" />{/if}
	{label}
</button>
