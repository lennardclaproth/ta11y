<script lang="ts">
	// Prototype-only wrapper. The real AppShellTemplate is `h-dvh` with an internally scrolling
	// main region, which is correct in the app but means a narrow-screen screenshot stops at the
	// fold. `flow` renders the same regions in document flow so a mobile preview shows the whole
	// page. Nothing below this file changes; the app keeps using AppShellTemplate.
	import type { Snippet } from 'svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';

	type Props = {
		top: Snippet;
		children: Snippet;
		flow?: boolean;
	};

	let { top, children, flow = false }: Props = $props();
</script>

{#if flow}
	<div class="flex min-h-dvh flex-col bg-taupe-100 text-slate-800">
		<div class="shrink-0">{@render top()}</div>
		<main class="flex-1">{@render children()}</main>
	</div>
{:else}
	<AppShellTemplate {top}>
		{@render children()}
	</AppShellTemplate>
{/if}
