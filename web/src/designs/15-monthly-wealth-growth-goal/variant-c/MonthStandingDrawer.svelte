<script lang="ts">
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import {
		bestStreak,
		currentStreak,
		goalMonths,
		monthStatusLabel,
		sharePercent
	} from '../goal-data';

	type Props = { open?: boolean };

	let { open = $bindable(false) }: Props = $props();

	const badgeIntent = {
		met: 'success',
		missed: 'warning',
		'in-progress': 'neutral'
	} as const;
</script>

<Drawer bind:open title="Monthly standing" width="max-w-xl">
	{#snippet header()}
		<Text size="sm" tone="muted">
			Goal 30% of income · {currentStreak} months in a row · best {bestStreak}
		</Text>
	{/snippet}

	<ul class="flex flex-col">
		{#each goalMonths as month (month.month)}
			<li class="flex flex-col gap-2 border-t border-slate-300 py-3 first:border-t-0">
				<div class="flex items-baseline justify-between gap-3">
					<span class="font-heading text-xl tracking-tight text-slate-900">{month.label}</span>
					<Badge intent={badgeIntent[month.status]} variant="soft" size="sm">
						{monthStatusLabel[month.status]}
					</Badge>
				</div>

				<dl class="grid grid-cols-3 gap-3">
					<div class="flex flex-col gap-0.5">
						<dt class="text-xs text-slate-500">Income</dt>
						<dd><Money amount={month.income} currency="EUR" size="sm" /></dd>
					</div>
					<div class="flex flex-col gap-0.5">
						<dt class="text-xs text-slate-500">To wealth</dt>
						<dd><Money amount={month.contributed} currency="EUR" size="sm" weight="semibold" /></dd>
					</div>
					<div class="flex flex-col items-end gap-0.5">
						<dt class="text-xs text-slate-500">Share · goal</dt>
						<dd class="text-sm text-slate-900 tabular-nums">
							{sharePercent(month)}% · {month.goalPercent}%
						</dd>
					</div>
				</dl>

				{#if month.unassigned > 0}
					<button
						type="button"
						class="w-fit text-sm text-sky-700 underline underline-offset-2 hover:text-sky-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
					>
						{month.unassigned} transactions not assigned yet
					</button>
				{/if}
			</li>
		{/each}
	</ul>

	{#snippet footer()}
		<Button variant="outline" size="md">Adjust goal</Button>
	{/snippet}
</Drawer>
