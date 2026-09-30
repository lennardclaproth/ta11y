<script lang="ts">
	import type { Snippet } from 'svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';

	/**
	 * LedgerToolbar with two additions: the table's own search box, and an `actions`
	 * snippet so a table can carry more than one action (add, import).
	 */
	type Props = {
		title: string;
		searchPlaceholder?: string;
		searchValue?: string;
		showSearch?: boolean;
		children?: Snippet;
		actions: Snippet;
	};

	let {
		title,
		searchPlaceholder = 'Search…',
		searchValue = '',
		showSearch = true,
		children,
		actions
	}: Props = $props();
</script>

<div
	class="flex shrink-0 flex-col gap-3 border-b border-slate-200 px-4 py-3 md:flex-row md:flex-wrap md:items-center md:justify-between"
>
	<div class="flex min-w-0 flex-col gap-3 md:flex-1 md:flex-row md:flex-wrap md:items-center md:gap-4">
		<h2 class="text-2xl leading-none">{title}</h2>
		{@render children?.()}
		{#if showSearch}
			<div class="w-full md:w-72">
				<SearchInput
					value={searchValue}
					placeholder={searchPlaceholder}
					size="md"
					shape="default"
					ariaLabel={searchPlaceholder}
				/>
			</div>
		{/if}
	</div>
	<div class="flex flex-wrap items-center gap-2">{@render actions()}</div>
</div>
