<script lang="ts">
	import type { Snippet } from 'svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';

	type Props = {
		/** Section heading for the chart or analytics content. */
		title?: string;
		/**
		 * Actions that act on this chart or the calculation behind it (rebuilding, refreshing).
		 * They sit on the section rule, so what a control affects is visible from where it hangs.
		 */
		actions?: Snippet;
		/** Extra layout classes (e.g. grid column spans). */
		class?: string;
		children: Snippet;
	};

	let { title, actions, class: className = '', children }: Props = $props();
</script>

<section
	class={['min-w-0 border-t border-slate-400 pt-3 pb-2', className].filter(Boolean).join(' ')}
>
	{#if title || actions}
		<div class="mb-3 flex min-h-8 flex-wrap items-center justify-between gap-2">
			{#if title}
				<Heading level="h2" size="md">{title}</Heading>
			{/if}
			{#if actions}
				<div class="flex flex-wrap items-center gap-2">{@render actions()}</div>
			{/if}
		</div>
	{/if}
	{@render children()}
</section>
