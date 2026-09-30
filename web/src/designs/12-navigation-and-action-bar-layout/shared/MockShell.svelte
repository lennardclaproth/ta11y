<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Stand-in for AppShellTemplate + PageContentTemplate: same canvas, spacing and stacking,
	 * but it grows with its content so a full-page screenshot shows the whole prototype
	 * instead of one viewport. The real screens keep the existing templates.
	 */
	type Props = { top: Snippet; children: Snippet };

	let { top, children }: Props = $props();
</script>

<div class="flex min-h-dvh flex-col bg-taupe-100 text-slate-800">
	<div class="shrink-0">{@render top()}</div>
	<main class="flex flex-1 flex-col gap-5 px-4 pb-6 lg:px-8">
		{@render children()}
	</main>
</div>
