<script lang="ts">
	import type { Snippet } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import { zClasses } from '$lib/styles/z-index';

	type Props = {
		/** Two-way bindable open state. */
		open?: boolean;
		title?: string;
		side?: 'right' | 'left';
		/** Tailwind max-width utility for the panel. */
		width?: string;
		dismissible?: boolean;
		closeOnBackdrop?: boolean;
		/** Pad the body; set false for content that fills the panel edge to edge, like a table. */
		padded?: boolean;
		onClose?: () => void;
		header?: Snippet;
		footer?: Snippet;
		children: Snippet;
		class?: string;
	};

	let {
		open = $bindable(false),
		title,
		side = 'right',
		width = 'max-w-md',
		dismissible = true,
		closeOnBackdrop = true,
		padded = true,
		onClose,
		header,
		footer,
		children,
		class: className = ''
	}: Props = $props();

	function close() {
		if (open === false) return;
		open = false;
		onClose?.();
	}

	// Close on Escape while open, unless a layer above already claimed the key: an open popover
	// inside the drawer calls preventDefault() so one press steps out of the popover rather than
	// discarding the whole drawer. stopPropagation() cannot do that — both listeners sit on
	// `document`, so neither stops the other.
	// Only the topmost layer answers: an open native <dialog> sits above every drawer, and of two
	// drawers the later one paints over the earlier. Without this, one press over a drawer opened
	// from another drawer would close both.
	let panel = $state<HTMLDivElement | null>(null);

	function isTopmost(): boolean {
		if (!panel) return false;
		const dialogs = Array.from(document.querySelectorAll('dialog[open]'));
		if (dialogs.some((dialog) => !panel?.contains(dialog))) return false;
		// A drawer on its way out is inert while it animates, and no longer counts as a layer.
		const drawers = Array.from(document.querySelectorAll('[data-drawer-panel]')).filter(
			(drawer) => !drawer.closest('[inert]')
		);
		return drawers[drawers.length - 1] === panel;
	}

	$effect(() => {
		if (!open) return;
		function onKeydown(event: KeyboardEvent) {
			if (event.key === 'Escape' && dismissible && !event.defaultPrevented && isTopmost()) close();
		}
		document.addEventListener('keydown', onKeydown, true);
		return () => document.removeEventListener('keydown', onKeydown, true);
	});
</script>

{#if open}
	<div
		class={[
			'fixed inset-0 flex',
			side === 'right' ? 'justify-end' : 'justify-start',
			zClasses.modal
		].join(' ')}
	>
		<button
			type="button"
			aria-label="Close"
			class="absolute inset-0 bg-slate-900/50"
			transition:fade={{ duration: 200 }}
			onclick={() => closeOnBackdrop && close()}
		></button>

		<div
			bind:this={panel}
			data-drawer-panel
			role="dialog"
			aria-modal="true"
			aria-label={title}
			class={['relative flex h-full w-full flex-col bg-white shadow-2xl', width, className]
				.filter(Boolean)
				.join(' ')}
			transition:fly={{ x: side === 'right' ? 360 : -360, duration: 200 }}
		>
			{#if title || header || dismissible}
				<header class="flex items-start justify-between gap-4 border-b border-slate-200 px-5 py-4">
					<div class="min-w-0 flex-1">
						{#if title}
							<!-- Matches Dialog: the shared serif scale and weight, keeping the base-layer colour. -->
							<Heading level="h2" size="lg" class="text-slate-900">{title}</Heading>
						{/if}
						{#if header}
							{@render header()}
						{/if}
					</div>
					{#if dismissible}
						<button
							type="button"
							aria-label="Close"
							class="-mt-1 -mr-1 inline-flex size-8 shrink-0 items-center justify-center rounded-full text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
							onclick={close}
						>
							<Icon icon="heroicons:x-mark" size="md" />
						</button>
					{/if}
				</header>
			{/if}

			<div
				class={padded
					? 'min-h-0 flex-1 overflow-y-auto px-5 py-4'
					: 'flex min-h-0 flex-1 flex-col overflow-hidden'}
			>
				{@render children()}
			</div>

			{#if footer}
				<footer class="flex items-center justify-end gap-2 border-t border-slate-200 px-5 py-4">
					{@render footer()}
				</footer>
			{/if}
		</div>
	</div>
{/if}
