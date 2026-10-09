<script lang="ts">
	/**
	 * Variant B — proposed replacement for the Alert molecule. The intent lives in a stub column
	 * (coloured edge, icon, word); the message keeps the ordinary slate reading colour.
	 */
	import type { Snippet } from 'svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';
	import {
		ledgerBaseClasses,
		ledgerIcons,
		ledgerStubClasses,
		ledgerStubLabels,
		ledgerSurfaceClasses
	} from './ledger-notice.variants';

	type Props = {
		intent?: AlertIntent;
		surface?: 'inline' | 'paper';
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
		surface = 'inline',
		title,
		dismissible = false,
		onDismiss,
		action,
		class: className = '',
		children
	}: Props = $props();

	const classes = $derived(
		[ledgerBaseClasses, ledgerSurfaceClasses[surface], ledgerStubClasses[intent], className]
			.filter(Boolean)
			.join(' ')
	);

	const role = $derived(intent === 'error' || intent === 'warning' ? 'alert' : 'status');
</script>

<div class={classes} {role}>
	<div
		class="flex w-24 shrink-0 flex-col gap-1 border-r border-slate-200 px-3 py-3 sm:w-28 sm:px-4"
	>
		<Icon icon={ledgerIcons[intent]} size="md" />
		<span class="text-xs font-semibold tracking-wider uppercase">{ledgerStubLabels[intent]}</span>
	</div>

	<div class="flex min-w-0 flex-1 flex-col gap-1 px-3 py-3 sm:px-4">
		{#if title}
			<p class="text-sm font-medium text-slate-900">{title}</p>
		{/if}
		{#if children}
			<div class="text-sm text-slate-700">{@render children()}</div>
		{/if}
		{#if action}
			<div class="pt-1">{@render action()}</div>
		{/if}
	</div>

	{#if dismissible}
		<div class="shrink-0 py-2 pr-2">
			<IconButton
				icon="heroicons:x-mark"
				ariaLabel="Dismiss"
				size="sm"
				variant="ghost"
				intent="secondary"
				onclick={onDismiss}
			/>
		</div>
	{/if}
</div>
