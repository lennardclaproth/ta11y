<script lang="ts">
	import type { IconifyIcon } from '@iconify/svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';

	/**
	 * Proposed press treatment for the Button atom. Today `primary/solid` sets
	 * `active:bg-amber-400`, which is its resting colour, so pressing it only scales the button
	 * by 2% — too little to read as "that click registered".
	 *
	 * This deepens the surface to amber-500, darkens the border and presses the face inwards.
	 * In the build this belongs in `button.variants.ts`, not in a wrapper.
	 */
	type Props = {
		label: string;
		icon?: string | IconifyIcon;
		loading?: boolean;
		disabled?: boolean;
		onclick?: (event: MouseEvent) => void;
	};

	let { label, icon, loading = false, disabled = false, onclick }: Props = $props();
</script>

<Button
	shape="default"
	{loading}
	{disabled}
	{onclick}
	class="active:border-slate-900 active:bg-amber-500 active:shadow-inner"
>
	{#if icon}<Icon {icon} size="sm" />{/if}
	{label}
</Button>
