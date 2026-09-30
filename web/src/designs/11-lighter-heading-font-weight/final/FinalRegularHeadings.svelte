<script lang="ts">
	// Final design for issue #11 — "Regular" (variant A, chosen by Lennard).
	//
	// Every EB Garamond heading in the app drops to regular 400: the one weight for which a real
	// font file already ships, so nothing is faked by the browser any more. Solved centrally, so
	// headings that never touch the Heading atom (bare <h1>/<h2>, the small drawer labels) follow
	// the same line. No new font asset, no new heading style, no change to sizes or leading.
	import '../shared/heading-weight.css';

	import HeadingContexts from '../shared/HeadingContexts.svelte';
	import WeightBand from '../shared/WeightBand.svelte';

	type Props = {
		state?: 'default' | 'loading' | 'empty' | 'error' | 'no-matches';
		/** The comparison band is only useful on the reference story. */
		showBand?: boolean;
	};

	let { state = 'default', showBand = true }: Props = $props();
</script>

<div class="min-h-dvh bg-taupe-100">
	{#if showBand}
		<WeightBand weightClass="d11-weight-400" />
	{/if}

	<div class="d11-weight-400 pt-4">
		<HeadingContexts {state} idPrefix="d11-{state}" />
	</div>
</div>
