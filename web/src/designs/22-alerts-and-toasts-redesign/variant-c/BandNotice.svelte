<script lang="ts">
	/**
	 * Variant C — proposed replacement for the Alert molecule. One band shape for every
	 * notification; the intent is a solid chip, the message stays slate.
	 */
	import type { Snippet } from 'svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';
	import { bandBaseClasses, bandChipClasses, bandIcons, bandLabels } from './band-notice.variants';

	type Props = {
		intent?: AlertIntent;
		title?: string;
		dismissible?: boolean;
		onDismiss?: () => void;
		/** Recovery action ("Try again"); only for notices that stay. */
		action?: Snippet;
		class?: string;
		children?: Snippet;
	};

	let {
		intent = 'info',
		title,
		dismissible = false,
		onDismiss,
		action,
		class: className = '',
		children
	}: Props = $props();

	const classes = $derived([bandBaseClasses, className].filter(Boolean).join(' '));

	const role = $derived(intent === 'error' || intent === 'warning' ? 'alert' : 'status');
</script>

<div class={classes} {role}>
	<span
		class={['inline-flex size-7 shrink-0 items-center justify-center', bandChipClasses[intent]].join(
			' '
		)}
	>
		<Icon icon={bandIcons[intent]} size="md" />
		<span class="sr-only">{bandLabels[intent]}</span>
	</span>

	<div class="min-w-0 flex-1">
		{#if title}
			<p class="text-sm font-medium text-slate-900">{title}</p>
		{/if}
		{#if children}
			<p class="text-sm text-slate-700">{@render children()}</p>
		{/if}
	</div>

	{#if action}
		<div class="shrink-0">{@render action()}</div>
	{/if}

	{#if dismissible}
		<IconButton
			icon="heroicons:x-mark"
			ariaLabel="Dismiss"
			size="sm"
			variant="ghost"
			intent="secondary"
			class="-mr-2 shrink-0"
			onclick={onDismiss}
		/>
	{/if}
</div>
