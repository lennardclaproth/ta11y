<script lang="ts">
	/**
	 * Variant A — proposed replacement for the Alert molecule. Same props as today's Alert
	 * (`intent`, `title`, `dismissible`, `onDismiss`) plus `surface` and an optional `action`.
	 */
	import type { Snippet } from 'svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';
	import {
		ruledIcons,
		ruledKickerClasses,
		ruledKickerLabels,
		ruledRuleClasses,
		ruledSurfaceClasses
	} from './ruled-notice.variants';

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
		[
			'flex flex-col gap-1',
			ruledSurfaceClasses[surface],
			ruledRuleClasses[intent],
			className
		]
			.filter(Boolean)
			.join(' ')
	);

	// Errors and warnings interrupt; info and success announce politely.
	const role = $derived(intent === 'error' || intent === 'warning' ? 'alert' : 'status');
</script>

<div class={classes} {role}>
	<div class="flex items-start justify-between gap-3">
		<p
			class={['flex items-center gap-1.5 text-xs font-semibold tracking-wider uppercase', ruledKickerClasses[intent]].join(' ')}
		>
			<Icon icon={ruledIcons[intent]} size="sm" />
			{ruledKickerLabels[intent]}
		</p>

		{#if dismissible}
			<IconButton
				icon="heroicons:x-mark"
				ariaLabel="Dismiss"
				size="sm"
				variant="ghost"
				intent="secondary"
				class="-mt-1.5 -mr-2 shrink-0"
				onclick={onDismiss}
			/>
		{/if}
	</div>

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
