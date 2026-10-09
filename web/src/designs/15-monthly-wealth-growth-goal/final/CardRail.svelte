<script lang="ts">
	// Proposal for a new molecule: the analytics band as one horizontally scrolling row of cards.
	// Cards are grouped ("Cashflow overview", "Wealth goal"); the group links above are both the
	// indicator of where you are and the way to jump, so a goal you have not scrolled to is still
	// named on screen. Snapping is mandatory, so the rail always comes to rest on a card edge.
	import { untrack, type Snippet } from 'svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';

	type Group = { id: string; label: string };

	type Props = {
		groups: Group[];
		/** `data-card` of the card the rail opens on; the real page simply starts at the left. */
		startAt?: string;
		/** Explanatory line beside the arrows. */
		hint?: string;
		children: Snippet;
	};

	let {
		groups,
		startAt,
		hint = 'Scroll sideways for the rest — which cards you see becomes a setting later.',
		children
	}: Props = $props();

	let track = $state<HTMLDivElement | null>(null);
	let active = $state(untrack(() => groups[0]?.id ?? ''));

	function cards(): HTMLElement[] {
		return track ? Array.from(track.querySelectorAll<HTMLElement>('[data-card]')) : [];
	}

	// The group whose first card still occupies the left edge of the viewport is the one you are
	// reading, so a half-scrolled rail does not flip the indicator early.
	function syncActive() {
		const node = track;
		if (!node) return;
		const edge = node.scrollLeft + 24;
		const current = cards().find((card) => card.offsetLeft + card.offsetWidth > edge);
		active = current?.dataset.group ?? groups[0]?.id ?? '';
	}

	function scrollTo(card: HTMLElement | undefined, smooth: boolean) {
		if (!track || !card) return;
		track.scrollTo({ left: card.offsetLeft, behavior: smooth ? 'smooth' : 'auto' });
	}

	function jump(groupId: string) {
		scrollTo(
			cards().find((card) => card.dataset.group === groupId),
			true
		);
	}

	function nudge(direction: 1 | -1) {
		const node = track;
		if (!node) return;
		const list = cards();
		const here = list.findIndex((card) => card.offsetLeft >= node.scrollLeft - 8);
		const index = direction === 1 ? here + 1 : here - 1;
		scrollTo(list[Math.min(list.length - 1, Math.max(0, index))], true);
	}

	// The prototype opens on a given card so a screenshot can show the goal half of the rail. The
	// second pass runs after the charts have taken their final width, which otherwise moves the
	// card out from under the scroll position.
	$effect(() => {
		if (!track) return;
		const align = () => {
			if (startAt) scrollTo(cards().find((card) => card.dataset.card === startAt), false);
			syncActive();
		};
		align();
		const timer = setTimeout(align, 400);
		window.addEventListener('resize', align);
		return () => {
			clearTimeout(timer);
			window.removeEventListener('resize', align);
		};
	});
</script>

<div class="flex flex-col gap-2">
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
			<Text size="xs" tone="muted" class="hidden sm:block">{hint}</Text>
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
		aria-label="Cashflow and wealth-goal cards"
		class="relative flex snap-x snap-mandatory gap-4 overflow-x-auto pb-2"
	>
		{@render children()}
	</div>
</div>
