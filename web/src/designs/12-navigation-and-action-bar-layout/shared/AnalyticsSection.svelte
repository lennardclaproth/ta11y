<script lang="ts">
	import type { Snippet } from 'svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';

	/**
	 * Proposed extension of AnalyticsCard: the same ruled section, plus an optional
	 * `actions` snippet on the rule for actions that belong to the chart (rebuild, refresh).
	 */
	type Props = {
		title?: string;
		actions?: Snippet;
		class?: string;
		children: Snippet;
	};

	let { title, actions, class: className = '', children }: Props = $props();
</script>

<section
	class={['min-w-0 border-t border-slate-400 pt-3 pb-2', className].filter(Boolean).join(' ')}
>
	<div class="mb-3 flex min-h-8 flex-wrap items-center justify-between gap-2">
		{#if title}
			<Heading level="h2" size="md" weight="medium">{title}</Heading>
		{/if}
		{#if actions}
			<div class="flex items-center gap-2">{@render actions()}</div>
		{/if}
	</div>
	{@render children()}
</section>
