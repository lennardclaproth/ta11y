<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		label: string;
		hint?: string;
		children: Snippet;
	};

	let { label, hint, children }: Props = $props();
</script>

<section class="flex flex-col gap-3">
	<div class="flex flex-wrap items-baseline gap-x-3 gap-y-1 border-b border-slate-800 pb-1">
		<h3 class="font-heading text-xl text-slate-900">{label}</h3>
		{#if hint}<p class="text-xs tracking-wider text-slate-500 uppercase">{hint}</p>{/if}
	</div>
	{@render children()}
</section>
