<script lang="ts">
	import type { Snippet } from 'svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { icons } from '../shared/icons';

	/**
	 * Editorial table header: the section rule carries the title, the row count and the actions
	 * that act on the rows; searching and filtering sit on the ruled line directly underneath,
	 * the full width of the table.
	 *
	 * Search notes from the review of round 1:
	 * - the offset outline ring is gone; focus is the line itself thickening to slate-800 with
	 *   the amber press surface behind it, which is the same amber the buttons focus with;
	 * - Chromium's own clear glyph for `type="search"` is suppressed, so the single X on the
	 *   right is this component's, and it only appears once something is typed.
	 */
	type Props = {
		title: string;
		meta?: string;
		searchPlaceholder?: string;
		showSearch?: boolean;
		searchValue?: string;
		/** Tabs or other context rendered next to the title. */
		before?: Snippet;
		/** Filters that belong to the rows, rendered at the end of the search line. */
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
		<div class="flex flex-wrap items-center gap-3">{@render actions()}</div>
	</div>

	{#if showSearch}
		<div
			class="flex items-center gap-2 border-t border-slate-200 px-1 transition-colors duration-150 ease-out focus-within:bg-amber-100/60"
		>
			<Icon icon={icons.search} size="sm" class="shrink-0 text-slate-500" />
			<input
				type="search"
				bind:value={searchValue}
				placeholder={searchPlaceholder}
				aria-label={searchPlaceholder}
				class="h-10 w-full min-w-0 border-b-2 border-transparent bg-transparent text-sm text-slate-900 placeholder:text-slate-500 focus:border-slate-800 focus:outline-none [&::-webkit-search-cancel-button]:appearance-none"
			/>
			{#if searchValue}
				<button
					type="button"
					aria-label="Clear search"
					onclick={() => (searchValue = '')}
					class="inline-flex size-8 shrink-0 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-200/70 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none active:bg-slate-300"
				>
					<Icon icon={icons.close} size="sm" />
				</button>
			{/if}
			{#if filters}
				<div class="flex shrink-0 items-center gap-2 border-l border-slate-200 pl-2">
					{@render filters()}
				</div>
			{/if}
		</div>
	{/if}
</div>
