<script lang="ts">
	// One recurring item: how the amount moved, which transactions are linked, and the way out.
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { chartColors } from '$lib/charts/theme';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { linkedTransactions, rhythmLabels, type RecurringItem } from '../recurring.fixture';
	import { historyDates, nextExpectedLabel, tickLabel } from '../recurring.format';

	type Props = {
		open?: boolean;
		item: RecurringItem;
		onEnd?: () => void;
	};

	let { open = $bindable(false), item, onEnd }: Props = $props();

	const labels = $derived(historyDates(item.lastSeen, item.rhythm, item.history.length));
	const directionWord = $derived(item.direction === 'in' ? 'Income' : 'Expense');
</script>

<Drawer bind:open title={item.name} width="max-w-2xl">
	{#snippet header()}
		<div class="mt-1 flex flex-wrap items-center gap-2">
			<Badge intent="neutral" variant="soft" size="sm">{directionWord}</Badge>
			<Badge intent="neutral" variant="outline" size="sm">{rhythmLabels[item.rhythm]}</Badge>
			<Text as="span" size="sm" tone="muted">
				{item.endedFrom
					? `Ended from ${item.endedFrom}`
					: `Next expected ${nextExpectedLabel(item.nextExpected)}`}
			</Text>
		</div>
	{/snippet}

	<!-- Slate, not red: the chart reports the amounts as they were paid. ta11y never calls a rise
	     "more expensive" — reading that is the point of showing it. -->
	<AnalyticsCard title="Amount over time">
		<TimeSeriesChart
			height="h-40"
			{labels}
			xTickFormat={tickLabel}
			ariaLabel={`Amounts paid for ${item.name}, oldest to newest`}
			datasets={[{ label: 'Amount', data: item.history, color: chartColors.net, fill: true }]}
		/>
		<div class="mt-2 flex flex-wrap items-baseline justify-between gap-2">
			<Text as="span" size="xs" tone="muted">
				{item.history.length} amounts, {formatDisplayDate(labels[0])} – {formatDisplayDate(
					item.lastSeen
				)}
			</Text>
			<span class="flex items-baseline gap-2 text-xs text-slate-500">
				First <Money amount={item.history[0]} currency="EUR" size="sm" weight="normal" />
				· Last <Money amount={item.amount} currency="EUR" size="sm" />
			</span>
		</div>
	</AnalyticsCard>

	<AnalyticsCard title="Linked transactions" class="mt-5">
		<Text size="sm" tone="muted" class="mb-2">
			{item.linked} transactions, newest first. Unlink one that does not belong — the transaction itself
			stays in Cashflow.
		</Text>
		<table class="w-full border-collapse text-sm">
			<thead>
				<tr class="border-b border-slate-400 text-xs font-semibold text-slate-700">
					<th scope="col" class="py-2 pr-3 text-left">Date</th>
					<th scope="col" class="py-2 pr-3 text-left">Description</th>
					<th scope="col" class="w-24 py-2 pr-3 text-left">Entered</th>
					<th scope="col" class="w-24 py-2 pr-3 text-right">Amount</th>
					<th scope="col" class="w-20 py-2 text-right"><span class="sr-only">Action</span></th>
				</tr>
			</thead>
			<tbody>
				{#each linkedTransactions as transaction (transaction.id)}
					<tr class="border-b border-slate-100">
						<td class="py-2 pr-3 whitespace-nowrap text-slate-700">
							{formatDisplayDate(transaction.date)}
						</td>
						<td class="py-2 pr-3 text-slate-700">{transaction.description}</td>
						<td class="py-2 pr-3">
							<Badge intent={transaction.origin === 'Manual' ? 'info' : 'neutral'} size="sm">
								{transaction.origin}
							</Badge>
						</td>
						<td class="py-2 pr-3 text-right">
							<Money amount={transaction.amount} currency="EUR" size="sm" />
						</td>
						<td class="py-2 text-right">
							<Button size="sm" variant="ghost" intent="secondary" shape="default">Unlink</Button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</AnalyticsCard>

	{#snippet footer()}
		<Button variant="outline" shape="default" onclick={() => (open = false)}>Close</Button>
		<Button variant="outline" intent="error" shape="default" onclick={onEnd}>End from month…</Button>
	{/snippet}
</Drawer>
