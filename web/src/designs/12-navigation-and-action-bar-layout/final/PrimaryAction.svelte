<script lang="ts">
	import type { IconifyIcon } from '@iconify/svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { pressClasses } from '../shared/press';

	/**
	 * Proposed press treatment for the Button atom. Today `primary/solid` sets
	 * `active:bg-amber-400`, which is its resting colour, so a press only scales the button by 2%
	 * and a mouse click shows nothing at all (`focus-visible` stays off for pointer input).
	 *
	 * The fix keeps the button as it is and adds two things: the app's own amber ring while
	 * pressed, and a black bar under it. In the build this belongs in `button.variants.ts`.
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

<Button shape="default" {loading} {disabled} {onclick} class={pressClasses}>
	{#if icon}<Icon {icon} size="sm" />{/if}
	{label}
</Button>
