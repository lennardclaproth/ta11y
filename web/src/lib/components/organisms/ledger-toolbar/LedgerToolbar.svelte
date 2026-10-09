<script lang="ts">
	import type { Snippet } from 'svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	let {
		title,
		actionLabel,
		actionIcon,
		onAdd,
		children
	}: {
		title?: string;
		/** Omit the label to leave the header without an action. */
		actionLabel?: string;
		/** Iconify id for an action that is not a creation; defaults to the bundled plus. */
		actionIcon?: string;
		onAdd?: () => void;
		children?: Snippet;
	} = $props();
	// Bundle this essential control icon so creation remains recognizable offline.
	const plus = {
		width: 24,
		height: 24,
		body: '<path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4.5v15m7.5-7.5h-15"/>'
	};
</script>

<div
	class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
	data-ledger-toolbar
>
	<div class="flex min-w-0 flex-wrap items-center gap-4">
		{#if title}<h2 class="text-2xl">{title}</h2>{/if}
		{@render children?.()}
	</div>
	{#if actionLabel}
		<Button shape="default" onclick={onAdd}><Icon icon={actionIcon ?? plus} />{actionLabel}</Button>
	{/if}
</div>
