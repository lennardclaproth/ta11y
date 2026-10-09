<script lang="ts">
	// No goal set yet: the standing has nothing to compare against, so it sends you to the
	// goal first instead of showing a zero.
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { goalFigureClass } from './wealth-goal-cards.variants';

	type Props = { onSet?: () => void };

	let { onSet }: Props = $props();
</script>

<div class="flex flex-col items-start gap-3 py-2">
	<span class={goalFigureClass}>No monthly goal yet</span>
	<Text size="sm" tone="muted" class="max-w-prose">
		Set the share of your income that should go towards wealth each month — savings, a broker
		deposit or an asset purchase. Then mark those transactions on Cashflow and this board keeps the
		score.
	</Text>
	<Button size="md" onclick={onSet}>Set monthly goal</Button>
</div>
