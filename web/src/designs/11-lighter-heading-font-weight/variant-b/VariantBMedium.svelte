<script lang="ts">
	// Variant B — "Medium". Every EB Garamond heading goes to 500, shipped as a real EB Garamond
	// Medium file declared on its own `font-weight: 500` descriptor. Slightly more presence than
	// variant A, still no browser-faked bold.
	//
	// The prototype approximates 500 with a hairline stroke because only the regular file is in
	// the repo today — see shared/heading-weight.css.
	import '../shared/heading-weight.css';

	import HeadingContexts from '../shared/HeadingContexts.svelte';
	import WeightBand from '../shared/WeightBand.svelte';
</script>

<div class="min-h-dvh bg-taupe-100">
	<WeightBand
		weightClass="d11-weight-500"
		proposedLabel="Variant B · medium 500, extra font file (approximated here)"
	/>

	<div class="d11-weight-500 pt-4">
		<HeadingContexts idPrefix="d11b" />
	</div>
</div>
