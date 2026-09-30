<script lang="ts">
	import type { Snippet } from 'svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { icons } from '../shared/icons';

	/**
	 * Editorial table header: the section rule carries the title, the row count and the
	 * supporting actions; search is the ruled line directly underneath, the width of the table.
	 */
	type Props = {
		title: string;
		meta: string;
		searchPlaceholder?: string;
		showSearch?: boolean;
		before?: Snippet;
		actions: Snippet;
	};

	let { title, meta, searchPlaceholder = 'Search…', showSearch = true, before, actions }: Props =
		$props();
</script>

<div class="shrink-0 border-b border-slate-200 px-4 pt-3 pb-2">
	<div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
		<div class="flex min-w-0 flex-wrap items-baseline gap-3">
			<h2 class="text-2xl leading-none">{title}</h2>
			<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">{meta}</Text>
			{@render before?.()}
		</div>
		<div class="flex flex-wrap items-center gap-4">{@render actions()}</div>
	</div>

	{#if showSearch}
		<div class="mt-2 flex items-center gap-2 border-t border-slate-200 pt-2">
			<Icon icon={icons.search} size="sm" class="text-slate-500" />
			<input
				type="search"
				placeholder={searchPlaceholder}
				aria-label={searchPlaceholder}
				class="h-8 w-full border-0 bg-transparent text-sm text-slate-900 placeholder:text-slate-500 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700"
			/>
		</div>
	{/if}
</div>
