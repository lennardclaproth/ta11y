<script lang="ts">
	import type { Snippet } from 'svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';

	/**
	 * Search-led table bar: searching the rows is the dominant control, the row actions sit
	 * behind a divider on the right. The page heading already names the table, so this bar
	 * carries no second heading.
	 */
	type Props = {
		searchPlaceholder: string;
		searchLabel: string;
		showSearch?: boolean;
		before?: Snippet;
		actions: Snippet;
	};

	let { searchPlaceholder, searchLabel, showSearch = true, before, actions }: Props = $props();
</script>

<div
	class="flex shrink-0 flex-wrap items-center gap-3 border-b border-slate-200 px-4 py-3 md:flex-nowrap"
>
	{#if before}
		<div class="flex shrink-0 items-center gap-3">{@render before()}</div>
	{/if}
	{#if showSearch}
		<div class="min-w-0 flex-1 basis-full md:basis-auto">
			<SearchInput
				placeholder={searchPlaceholder}
				ariaLabel={searchLabel}
				size="md"
				shape="default"
			/>
		</div>
	{:else}
		<div class="flex-1"></div>
	{/if}
	<div class="flex flex-wrap items-center gap-2 md:border-l md:border-slate-200 md:pl-3">
		{@render actions()}
	</div>
</div>
