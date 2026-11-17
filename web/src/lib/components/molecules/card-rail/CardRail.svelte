<script lang="ts">
	// The analytics band as one horizontally scrolling row of cards, grouped. The group links
	// above are both the indicator of where you are and the way to jump, so a group you have
	// not scrolled to is still named on screen. Snapping is mandatory, so the rail always
	// comes to rest on a card edge rather than leaving one half in view.
	//
	// Children carry `data-card` (a stable id) and `data-group` (a group id); the rail reads
	// those instead of owning the card list, so a page can lay its cards out as it likes.
	import type { Snippet } from 'svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import type { CardRailGroup } from './card-rail.types';

	type Props = {
		groups: CardRailGroup[];
		/** Accessible name for the scrolling region. */
		ariaLabel: string;
		/** Explanatory line beside the arrows. */
		hint?: string;
		children: Snippet;
		class?: string;
	};

	let { groups, ariaLabel, hint, children, class: className = '' }: Props = $props();

	let track = $state<HTMLDivElement | null>(null);
	// Empty until the rail has measured itself; until then the first group is the one you are
	// reading, which is where an un-scrolled rail starts.
	let scrolledTo = $state('');
	const active = $derived(scrolledTo || (groups[0]?.id ?? ''));

	function cards(): HTMLElement[] {
		return track ? Array.from(track.querySelectorAll<HTMLElement>('[data-card]')) : [];
	}

	// The group whose first card still occupies the left edge of the viewport is the one you
	// are reading, so a half-scrolled rail does not flip the indicator early.
	function syncActive() {
		const node = track;
		if (!node) return;
		const edge = node.scrollLeft + 24;
		const current = cards().find((card) => card.offsetLeft + card.offsetWidth > edge);
		scrolledTo = current?.dataset.group ?? '';
	}

	function scrollToCard(card: HTMLElement | undefined) {
		if (!track || !card) return;
		track.scrollTo({ left: card.offsetLeft, behavior: 'smooth' });
	}

	function jump(groupId: string) {
		scrollToCard(cards().find((card) => card.dataset.group === groupId));
	}

	function nudge(direction: 1 | -1) {
		const node = track;
		if (!node) return;
		const list = cards();
		const here = list.findIndex((card) => card.offsetLeft >= node.scrollLeft - 8);
		const index = direction === 1 ? here + 1 : here - 1;
		scrollToCard(list[Math.min(list.length - 1, Math.max(0, index))]);
	}

	// The charts settle on their final width after the first paint, which moves the cards out
	// from under the scroll position; re-reading on resize keeps the indicator honest.
	$effect(() => {
		if (!track) return;
		syncActive();
		window.addEventListener('resize', syncActive);
		return () => window.removeEventListener('resize', syncActive);
	});
</script>

<div class={['flex flex-col gap-2', className].filter(Boolean).join(' ')}>
	<div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
		<nav class="flex items-center gap-1" aria-label="Card groups">
			{#each groups as group (group.id)}
				<button
					type="button"
					aria-current={active === group.id ? 'true' : undefined}
					class={[
						'rounded-md px-2.5 py-1 text-sm transition-colors',
						'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
						active === group.id
							? 'bg-amber-100 text-slate-900'
							: 'text-slate-600 hover:bg-slate-100 hover:text-slate-800'
					].join(' ')}
					onclick={() => jump(group.id)}
				>
					{group.label}
				</button>
			{/each}
		</nav>

		<div class="flex items-center gap-3">
			{#if hint}
				<Text size="xs" tone="muted" class="hidden sm:block">{hint}</Text>
			{/if}
			<!-- The arrows are the keyboard (and non-touch) way through the rail, so they stay
			     visible at every width. -->
			<div class="flex gap-1">
				<IconButton
					icon="heroicons:chevron-left"
					ariaLabel="Previous card"
					size="sm"
					variant="outline"
					onclick={() => nudge(-1)}
				/>
				<IconButton
					icon="heroicons:chevron-right"
					ariaLabel="Next card"
					size="sm"
					variant="outline"
					onclick={() => nudge(1)}
				/>
			</div>
		</div>
	</div>

	<div
		bind:this={track}
		onscroll={syncActive}
		role="group"
		aria-label={ariaLabel}
		class="relative flex snap-x snap-mandatory gap-4 overflow-x-auto pb-2"
	>
		{@render children()}
	</div>
</div>
