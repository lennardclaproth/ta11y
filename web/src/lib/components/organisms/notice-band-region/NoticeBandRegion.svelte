<script lang="ts">
	/**
	 * The single notification place, in normal document flow directly under the masthead. The store,
	 * its timers and its API are unchanged — only where the notices render and what they look like.
	 *
	 * Stacked bands share one hairline (`-mt-px`) so several notices read as one ruled block rather
	 * than a pile of cards.
	 */
	import { slide } from 'svelte/transition';
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import { toast } from '$lib/stores/toast.svelte';

	type Props = {
		class?: string;
	};

	let { class: className = '' }: Props = $props();

	/**
	 * How many notices are shown at once. A band sits in the flow, so every extra one pushes the page
	 * down; beyond a few at a time the jump costs more than the notice is worth. The store keeps the
	 * rest, with their timers, and the next one appears as soon as a band above it goes.
	 */
	const MAX_VISIBLE = 3;

	const visible = $derived(toast.items.slice(0, MAX_VISIBLE));

	// A band sits in the flow, so it opens and closes by height rather than flying in from a corner.
	// Someone who asked for less motion gets the same band without the movement. This region lives
	// for the whole session, so the setting is followed while it is open, not only at startup.
	let reducedMotion = $state(false);
	$effect(() => {
		const query = window.matchMedia('(prefers-reduced-motion: reduce)');
		reducedMotion = query.matches;
		const onChange = (event: MediaQueryListEvent) => (reducedMotion = event.matches);
		query.addEventListener('change', onChange);
		return () => query.removeEventListener('change', onChange);
	});
	const slideDuration = $derived(reducedMotion ? 0 : 160);
</script>

{#if visible.length > 0}
	<div
		class={['flex flex-col', className].filter(Boolean).join(' ')}
		role="region"
		aria-label="Notifications"
	>
		{#each visible as item, index (item.id)}
			<div transition:slide={{ duration: slideDuration }}>
				<NoticeBand
					intent={item.intent}
					title={item.title}
					gutter="page"
					dismissible={item.dismissible}
					onDismiss={() => toast.dismiss(item.id)}
					class={index > 0 ? '-mt-px' : ''}
				>
					{item.message}
				</NoticeBand>
			</div>
		{/each}
	</div>
{/if}
