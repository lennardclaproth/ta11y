<script lang="ts">
	// The only encouragement the pitch allows: the run of months you met, nothing else.
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import StreakStrip from './StreakStrip.svelte';
	import { figureClass } from './figure';
	import { bestStreak, currentStreak, goalMonths } from '../goal-data';

	type Props = { months?: typeof goalMonths };

	let { months = goalMonths }: Props = $props();
</script>

<div class="flex flex-col gap-3">
	<span class={figureClass}>{currentStreak} months in a row</span>
	<StreakStrip {months} />
	<Text size="sm" tone="muted">
		Best run so far: {bestStreak} months. July is still running, so it does not count yet.
	</Text>
</div>
