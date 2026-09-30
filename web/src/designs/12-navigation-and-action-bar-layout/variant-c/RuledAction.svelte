<script lang="ts">
	import type { IconifyIcon } from '@iconify/svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';

	/**
	 * Proposed Button variant `ruled`: a text action that hangs on a rule instead of a filled
	 * or outlined box. Used for the supporting actions of a section, so the one dominant
	 * action per region stays the only filled button.
	 */
	type Props = { label: string; icon?: string | IconifyIcon };

	let { label, icon }: Props = $props();
</script>

<button
	type="button"
	class="inline-flex h-8 items-center gap-1.5 border-b border-slate-300 text-sm text-slate-700 transition-colors hover:border-slate-700 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700"
>
	{#if icon}<Icon {icon} size="sm" />{/if}
	{label}
</button>
