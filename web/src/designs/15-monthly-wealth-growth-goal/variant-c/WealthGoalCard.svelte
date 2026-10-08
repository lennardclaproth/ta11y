<script lang="ts">
	// Variant C's one new surface on Cashflow: a fifth analytics card for the running month.
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import ProgressBar from '$lib/components/atoms/progress-bar/ProgressBar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { currentStreak, goalAmount, goalMonths, sharePercent } from '../goal-data';

	type Props = { onOpen?: () => void };

	let { onOpen }: Props = $props();

	const running = goalMonths[0];
</script>

<AnalyticsCard title="Wealth goal">
	<div class="flex flex-col gap-2">
		<Money amount={running.contributed} currency="EUR" size="xl" weight="semibold" />
		<Text size="sm" tone="muted">of €{goalAmount(running).toLocaleString('en')} needed in July</Text>
		<ProgressBar
			value={running.contributed}
			max={goalAmount(running)}
			intent="primary"
			ariaLabel="{sharePercent(running)}% of July income put towards wealth, goal {running.goalPercent}%"
		/>
		<Text size="sm" tone="muted">
			{sharePercent(running)}% of income · goal {running.goalPercent}%
		</Text>
		<Text size="sm" tone="muted">Streak: {currentStreak} months in a row</Text>
		<Button size="sm" variant="ghost" intent="secondary" class="w-fit px-0" onclick={onOpen}>
			Monthly standing
		</Button>
	</div>
</AnalyticsCard>
