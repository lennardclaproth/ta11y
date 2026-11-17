<script lang="ts">
	/**
	 * The single toast place: a stack in the top-right corner of the content area that floats over
	 * the page instead of pushing it down. Each toast is the notice band printed as a card, so it
	 * reads like every other notice in the app. Problems that stay, and refusals inside a dialog,
	 * are still bands in the flow at the content they concern — only the toast store floats.
	 *
	 * Positioned against the nearest positioned ancestor (the app shell's content area), so it sits
	 * under the masthead and never covers the navigation. Everything the store holds is shown: the
	 * store caps itself, so there is no toast here waiting out of sight for a turn. A toast that is
	 * hovered or focused holds its countdown, so it does not vanish while it is being read.
	 */
	import { flip } from 'svelte/animate';
	import { fly } from 'svelte/transition';
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { zClasses } from '$lib/styles/z-index';

	type Props = {
		class?: string;
	};

	let { class: className = '' }: Props = $props();

	// Someone who asked for less motion gets the same toast without the movement. This region lives
	// for the whole session, so the setting is followed while it is open, not only at startup.
	let reducedMotion = $state(false);
	$effect(() => {
		const query = window.matchMedia('(prefers-reduced-motion: reduce)');
		reducedMotion = query.matches;
		const onChange = (event: MediaQueryListEvent) => (reducedMotion = event.matches);
		query.addEventListener('change', onChange);
		return () => query.removeEventListener('change', onChange);
	});
	const motionDuration = $derived(reducedMotion ? 0 : 160);
</script>

{#if toast.items.length > 0}
	<!-- The stack itself lets clicks through; only the cards take them, so the page between and
	     beside them stays usable. -->
	<div
		class={[
			'pointer-events-none absolute inset-x-4 top-4 flex flex-col gap-2 sm:left-auto sm:w-96 lg:right-8',
			zClasses.toast,
			className
		]
			.filter(Boolean)
			.join(' ')}
		role="region"
		aria-label="Notifications"
	>
		{#each toast.items as item (item.id)}
			<div
				class="pointer-events-auto"
				role="presentation"
				animate:flip={{ duration: motionDuration }}
				transition:fly={{ x: 24, duration: motionDuration }}
				onmouseenter={() => toast.pause(item.id)}
				onmouseleave={() => toast.resume(item.id)}
				onfocusin={() => toast.pause(item.id)}
				onfocusout={() => toast.resume(item.id)}
			>
				<NoticeBand
					intent={item.intent}
					title={item.title}
					gutter="panel"
					dismissible={item.dismissible}
					onDismiss={() => toast.dismiss(item.id)}
					class="rounded-xl border-x shadow-md"
				>
					{item.message}
				</NoticeBand>
			</div>
		{/each}
	</div>
{/if}
