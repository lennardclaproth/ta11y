<script lang="ts">
	// Puts today's heading (semibold, faked by the browser) next to the variant's weight,
	// at the page-title size, so the difference is judgeable before reading the rest.
	import Text from '$lib/components/atoms/typography/Text.svelte';

	type Props = {
		/** One of the d11-weight-* classes from heading-weight.css. */
		weightClass: string;
		proposedLabel: string;
	};

	let { weightClass, proposedLabel }: Props = $props();
</script>

<section class="border-b border-slate-300 px-4 pt-6 pb-6 lg:px-8">
	<Text size="xs" tone="muted" class="block tracking-wider uppercase">
		Page title — today vs this variant
	</Text>
	<div class="mt-3 grid gap-6 sm:grid-cols-2">
		<div>
			<Text size="xs" tone="muted" class="mb-1 block">Today · semibold, faked by the browser</Text>
			<p class="font-heading text-4xl leading-tight font-semibold tracking-tight text-slate-900">
				Cashflow
			</p>
		</div>
		<div class={weightClass}>
			<Text size="xs" tone="muted" class="mb-1 block">{proposedLabel}</Text>
			<p class="d11-sample font-heading text-4xl leading-tight tracking-tight text-slate-900">
				Cashflow
			</p>
		</div>
	</div>
</section>
