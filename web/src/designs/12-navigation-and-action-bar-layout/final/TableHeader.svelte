<script lang="ts">
	import type { Snippet } from 'svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { icons } from '../shared/icons';
	import { pressClasses } from '../shared/press';

	/**
	 * Editorial table header: the section rule carries the title, the row count, the row filters
	 * and the actions that act on the rows. Searching sits on the ruled line directly underneath,
	 * the full width of the table.
	 *
	 * Review of the final design:
	 * - the row filters (Open / Closed / All) moved off the search line onto the section rule,
	 *   next to the actions, so the search line is only the search;
	 * - the focus treatment is lighter: a pale amber wash and a darker glyph, no thick bar under
	 *   the field. It has to be noticeable without shouting.
	 */
	type Props = {
		title: string;
		meta?: string;
		searchPlaceholder?: string;
		showSearch?: boolean;
		searchValue?: string;
		/** Tabs or other context rendered next to the title. */
		before?: Snippet;
		/** Filters that belong to the rows, rendered on the section rule before the actions. */
		filters?: Snippet;
		actions: Snippet;
	};

	let {
		title,
		meta,
		searchPlaceholder = 'Search…',
		showSearch = true,
		searchValue = $bindable(''),
		before,
		filters,
		actions
	}: Props = $props();
</script>

<div class="shrink-0 border-b border-slate-300 px-4 pt-3">
	<div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 pb-2">
		<div class="flex min-w-0 flex-wrap items-baseline gap-3">
			<h2 class="text-2xl leading-none">{title}</h2>
			{#if meta}
				<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">{meta}</Text>
			{/if}
			{@render before?.()}
		</div>
		<div class="flex flex-wrap items-center gap-3">
			{#if filters}
				{@render filters()}
				<span class="h-5 w-px bg-slate-300" aria-hidden="true"></span>
			{/if}
			{@render actions()}
		</div>
	</div>

	{#if showSearch}
		<div
			class="group flex items-center gap-2 border-t border-slate-200 px-1 transition-colors duration-150 ease-out focus-within:bg-amber-100/40"
		>
			<Icon
				icon={icons.search}
				size="sm"
				class="shrink-0 text-slate-500 transition-colors group-focus-within:text-slate-800"
			/>
			<input
				type="search"
				bind:value={searchValue}
				placeholder={searchPlaceholder}
				aria-label={searchPlaceholder}
				class="h-10 w-full min-w-0 bg-transparent text-sm text-slate-900 placeholder:text-slate-500 focus:outline-none [&::-webkit-search-cancel-button]:appearance-none"
			/>
			{#if searchValue}
				<button
					type="button"
					aria-label="Clear search"
					onclick={() => (searchValue = '')}
					class="inline-flex size-8 shrink-0 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-200/70 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none {pressClasses}"
				>
					<Icon icon={icons.close} size="sm" />
				</button>
			{/if}
		</div>
	{/if}
</div>
