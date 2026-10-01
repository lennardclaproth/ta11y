<script lang="ts">
	import type { Snippet } from 'svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';

	/**
	 * The ledger header is where everything that acts on the rows lives: the section rule carries
	 * the title, how many rows there are, the filters that narrow them and the actions that
	 * change them, and searching is the ruled line directly underneath, the full width of the
	 * table. A control's place says what it affects, so nothing about these rows is reachable
	 * from the navigation bar any more.
	 */
	type Props = {
		title?: string;
		/** Row count or load state, e.g. `128 rows · 3 selected`. */
		meta?: string;
		/** One sentence under the title for screens that need to explain themselves. */
		description?: string;
		showSearch?: boolean;
		searchValue?: string;
		searchPlaceholder?: string;
		/** Debounce window (ms) before `onSearch` fires. */
		debounceMs?: number;
		/** Debounced query callback; fires immediately when the search is cleared. */
		onSearch?: (query: string) => void;
		/** Context next to the title, e.g. the view tabs. */
		before?: Snippet;
		/** Filters over the rows, on the section rule ahead of the actions. */
		filters?: Snippet;
		/** Actions on the rows: the supporting ones ruled, the one dominant action filled. */
		actions?: Snippet;
		class?: string;
	};

	let {
		title,
		meta,
		description,
		showSearch = false,
		searchValue = $bindable(''),
		searchPlaceholder = 'Search…',
		debounceMs = 300,
		onSearch,
		before,
		filters,
		actions,
		class: className = ''
	}: Props = $props();

	/** Local echo of the field, so typing stays responsive while `onSearch` waits out the pause. */
	let typed = $state(searchValue);
	let mounted = false;

	// Follow the committed value when it is changed from outside (clearing a filter elsewhere).
	$effect(() => {
		typed = searchValue;
	});

	$effect(() => {
		const current = typed;
		if (!mounted) {
			mounted = true;
			return;
		}
		if (current === searchValue) return;
		const timer = setTimeout(() => {
			searchValue = current;
			onSearch?.(current);
		}, debounceMs);
		return () => clearTimeout(timer);
	});

	function clear() {
		typed = '';
		searchValue = '';
		onSearch?.('');
	}
</script>

<div
	class={['shrink-0 border-b border-slate-300 px-4 pt-3', className].filter(Boolean).join(' ')}
	data-ledger-toolbar
>
	<div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2 pb-2">
		<div class="flex min-w-0 flex-wrap items-baseline gap-3">
			{#if title}<h2 class="text-2xl leading-none">{title}</h2>{/if}
			{#if meta}
				<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">{meta}</Text>
			{/if}
			{@render before?.()}
		</div>
		{#if filters || actions}
			<div class="flex flex-wrap items-center gap-3">
				{#if filters}
					{@render filters()}
					{#if actions}
						<span class="h-5 w-px bg-slate-300" aria-hidden="true"></span>
					{/if}
				{/if}
				{@render actions?.()}
			</div>
		{/if}
	</div>

	{#if description}
		<p class="max-w-prose pb-2 text-sm text-slate-600">{description}</p>
	{/if}

	{#if showSearch}
		<!-- Focus is the line itself: a pale amber wash and a darker glyph. It has to be
		     noticeable without shouting over the rows it filters. -->
		<div
			class="group flex items-center gap-2 border-t border-slate-200 px-1 transition-colors duration-150 ease-out focus-within:bg-amber-100/40"
		>
			<Icon
				icon="heroicons:magnifying-glass"
				size="sm"
				class="shrink-0 text-slate-500 transition-colors group-focus-within:text-slate-800"
			/>
			<input
				type="search"
				bind:value={typed}
				placeholder={searchPlaceholder}
				aria-label={searchPlaceholder}
				class="h-10 w-full min-w-0 bg-transparent text-sm text-slate-900 placeholder:text-slate-500 focus:outline-none [&::-webkit-search-cancel-button]:appearance-none"
			/>
			{#if typed}
				<button
					type="button"
					aria-label="Clear search"
					onclick={clear}
					class="inline-flex size-8 shrink-0 items-center justify-center rounded-md text-slate-500 transition-colors hover:bg-slate-200/70 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none active:ring-2 active:ring-amber-300"
				>
					<Icon icon="heroicons:x-mark" size="sm" />
				</button>
			{/if}
		</div>
	{/if}
</div>
