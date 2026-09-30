<script lang="ts">
	// Variant A — "Regular". Every EB Garamond heading drops to 400, the one weight for which a
	// real font file already ships. No new asset, no synthetic bold anywhere.
	import '../shared/heading-weight.css';

	import HeadingContexts from '../shared/HeadingContexts.svelte';
	import WeightBand from '../shared/WeightBand.svelte';
</script>

<div class="min-h-dvh bg-taupe-100">
	<WeightBand weightClass="d11-weight-400" proposedLabel="Variant A · regular 400, real font file" />

	<div class="d11-weight-400 pt-4">
		<HeadingContexts idPrefix="d11a" />
	</div>
</div>
