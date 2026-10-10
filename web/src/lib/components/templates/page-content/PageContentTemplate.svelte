<script lang="ts">
	import type { Snippet } from 'svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { zClasses } from '$lib/styles/z-index';

	type Props = {
		/**
		 * The page's own heading. It sits on the page rather than in the navigation bar, which
		 * carries nothing but destinations.
		 */
		title?: string;
		/** A short label beside the title, e.g. which kind of screen this is. */
		titleMeta?: string;
		/** Analytics section above the content (charts / KPIs); shrink-0. */
		analytics?: Snippet;
		/** Primary content; placed inside a flex-1 panel that owns its own scrolling. */
		children: Snippet;
		showFab?: boolean;
		fabIcon?: string;
		fabLabel?: string;
		onFabClick?: () => void;
		/** Render the content inside a panel surface (set false to lay out raw content). */
		panel?: boolean;
		class?: string;
	};

	let {
		title,
		titleMeta,
		analytics,
		children,
		showFab = false,
		fabIcon = 'heroicons:plus',
		fabLabel = 'Create',
		onFabClick,
		panel = true,
		class: className = ''
	}: Props = $props();
</script>

<div
	class={[
		'relative flex min-h-full flex-col gap-5 px-4 pb-6 lg:h-full lg:min-h-0 lg:px-8',
		className
	]
		.filter(Boolean)
		.join(' ')}
>
	{#if title}
		<div class="flex shrink-0 flex-wrap items-baseline justify-between gap-2 pt-1">
			<Heading level="h1" size="2xl" class="leading-none">{title}</Heading>
			{#if titleMeta}
				<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">{titleMeta}</Text>
			{/if}
		</div>
	{/if}

	{#if analytics}
		<div class="shrink-0">{@render analytics()}</div>
	{/if}

	{#if panel}
		<Panel
			variant="muted"
			shape="square"
			shadow="none"
			padding="none"
			class="flex h-[32rem] shrink-0 flex-col overflow-hidden lg:h-auto lg:min-h-0 lg:flex-1"
		>
			{@render children()}
		</Panel>
	{:else}
		<div class="flex h-[32rem] shrink-0 flex-col overflow-hidden lg:h-auto lg:min-h-0 lg:flex-1">
			{@render children()}
		</div>
	{/if}

	{#if showFab}
		<button
			type="button"
			aria-label={fabLabel}
			class={[
				'fixed right-6 bottom-6 inline-flex size-14 items-center justify-center rounded-full border border-amber-200 bg-slate-600 text-amber-200 shadow-lg',
				'transition-all duration-150 ease-out hover:bg-slate-500 active:scale-95',
				'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:ring-offset-2 focus-visible:outline-none',
				zClasses.fab
			].join(' ')}
			onclick={onFabClick}
		>
			<Icon icon={fabIcon} size="lg" />
		</button>
	{/if}
</div>
