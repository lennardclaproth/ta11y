<script lang="ts">
	/**
	 * Final design for #22 — proposed replacement for the Alert molecule.
	 *
	 * Same public API as today's Alert (`intent`, `title`, `dismissible`, `onDismiss`, `children`),
	 * plus `action` for an in-place recovery control and `surface` for the dialog case. Everything
	 * that renders a notification today — the toast host, the persistent page notices, the save
	 * error in a modal, the hand-rolled error boxes in admin pages, forms and drawers — renders this.
	 */
	import type { Snippet } from 'svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import type { AlertIntent } from '$lib/components/molecules/alert/alert.types';
	import {
		noticeBandBaseClasses,
		noticeBandChipClasses,
		noticeBandGutterClasses,
		noticeBandIcons,
		noticeBandLabels,
		noticeBandSurfaceClasses,
		type NoticeBandGutter,
		type NoticeBandSurface
	} from './notice-band.variants';

	type Props = {
		intent?: AlertIntent;
		title?: string;
		dismissible?: boolean;
		onDismiss?: () => void;
		surface?: NoticeBandSurface;
		gutter?: NoticeBandGutter;
		/** Recovery action ("Try again"). Only for notices that stay. */
		action?: Snippet;
		class?: string;
		children?: Snippet;
	};

	let {
		intent = 'info',
		title,
		dismissible = false,
		onDismiss,
		surface = 'paper',
		gutter = 'page',
		action,
		class: className = '',
		children
	}: Props = $props();

	const classes = $derived(
		[
			noticeBandBaseClasses,
			noticeBandSurfaceClasses[surface],
			noticeBandGutterClasses[gutter],
			className
		]
			.filter(Boolean)
			.join(' ')
	);

	// Errors and warnings interrupt (assertive); success and info announce politely. Same split as
	// today's Alert, so nothing regresses for screen readers.
	const role = $derived(intent === 'error' || intent === 'warning' ? 'alert' : 'status');
</script>

<div class={classes} {role}>
	<span
		class={[
			'inline-flex size-7 shrink-0 items-center justify-center',
			noticeBandChipClasses[intent]
		].join(' ')}
	>
		<Icon icon={noticeBandIcons[intent]} size="md" />
		<span class="sr-only">{noticeBandLabels[intent]}</span>
	</span>

	<div class="min-w-0 flex-1 self-center">
		{#if title}
			<p class="text-sm font-medium text-slate-900">{title}</p>
		{/if}
		{#if children}
			<p class="text-sm text-slate-700">{@render children()}</p>
		{/if}
	</div>

	{#if action}
		<div class="shrink-0 self-center">{@render action()}</div>
	{/if}

	{#if dismissible}
		<IconButton
			icon="heroicons:x-mark"
			ariaLabel="Dismiss"
			size="sm"
			variant="ghost"
			intent="secondary"
			class="-mr-2 shrink-0 self-center"
			onclick={onDismiss}
		/>
	{/if}
</div>
