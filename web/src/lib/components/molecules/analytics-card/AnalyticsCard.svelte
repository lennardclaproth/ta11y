<script lang="ts">
	import type { Snippet } from 'svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';

	type Props = {
		/** Section heading for the chart or analytics content. */
		title?: string;
		/** Extra layout classes (e.g. grid column spans). */
		class?: string;
		children: Snippet;
	};

	let { title, class: className = '', children }: Props = $props();
</script>

<section
	class={['min-w-0 border-t border-slate-400 pt-3 pb-2', className].filter(Boolean).join(' ')}
>
	{#if title}
		<Heading level="h2" size="md" weight="medium" class="mb-3">{title}</Heading>
	{/if}
	{@render children()}
</section>
