<script lang="ts">
	import type { IconifyIcon } from '@iconify/svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';

	/**
	 * Proposed Button variant `ruled`: a supporting action that hangs on a rule instead of a
	 * filled or outlined box, so the one dominant action per region stays the only filled button.
	 *
	 * Press feedback is explicit and does not rely on motion alone: while pressed the rule
	 * thickens to slate-900, the label darkens, the amber press surface appears and the label
	 * shifts down one pixel. Focus uses the app's amber ring, like Button.
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
	       hover:border-slate-700 hover:bg-amber-100/60 hover:text-slate-950
	       focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none
	       active:translate-y-px active:border-b-2 active:border-slate-900 active:bg-amber-200 active:text-slate-950
	       disabled:pointer-events-none disabled:opacity-50"
>
	{#if icon}<Icon {icon} size="sm" />{/if}
	{label}
</button>
