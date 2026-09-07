<script lang="ts">
	import type { Snippet } from 'svelte';
	import ToastHost from '$lib/components/organisms/toast-host/ToastHost.svelte';

	type Props = {
		/** Sticky top region (typically the TopNavbar). */
		top?: Snippet;
		/** Main scrollable area. */
		children: Snippet;
		/** Render the single app-level toast host. */
		withToastHost?: boolean;
		class?: string;
	};

	let { top, children, withToastHost = true, class: className = '' }: Props = $props();
</script>

<div
	class={['flex h-dvh min-h-0 flex-col bg-taupe-100 text-slate-800', className]
		.filter(Boolean)
		.join(' ')}
>
	<a
		href="#app-main"
		class="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-50 focus:bg-taupe-50 focus:p-3 focus:outline-2"
		>Skip to content</a
	>
	{#if top}
		<div class="shrink-0">{@render top()}</div>
	{/if}

	<main id="app-main" tabindex="-1" class="min-h-0 flex-1 overflow-auto lg:overflow-hidden">
		{@render children()}
	</main>

	{#if withToastHost}
		<ToastHost />
	{/if}
</div>
