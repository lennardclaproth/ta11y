<script lang="ts">
	// The goal itself: what you asked of yourself, from when it applies, and the way back to it.
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import { figureClass } from './figure';

	type Props = {
		percent: number;
		/** The goal in money for the month on screen, so the percentage is not the only unit. */
		amount?: number;
		onAdjust?: () => void;
	};

	let { percent, amount, onAdjust }: Props = $props();
</script>

<div class="flex flex-col gap-3">
	<span class={figureClass}>{percent}% of income</span>
	{#if amount !== undefined}
		<Text size="sm" tone="muted">
			<Money {amount} currency="EUR" size="sm" /> on July's income so far.
		</Text>
	{/if}
	<Text size="sm" tone="muted">
		Applies from July 2026. Earlier months keep the goal they had.
	</Text>
	<Button size="sm" variant="outline" class="w-fit" onclick={onAdjust}>Adjust goal</Button>
</div>
