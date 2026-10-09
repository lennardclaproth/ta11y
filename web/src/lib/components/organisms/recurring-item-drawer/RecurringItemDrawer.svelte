<script lang="ts">
	// One recurring item: how the amount moved, which transactions are linked, and the
	// way out. The chart is slate, not red — it reports the amounts as they were paid.
	// ta11y never calls a rise "more expensive"; reading that is the point of showing it.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import { chartColors } from '$lib/charts/theme';
	import { scaledToNumber } from '$lib/api/money';
	import {
		endedFromLabel,
		nextExpectedLabel,
		recurringDirectionLabel,
		recurringTickLabel,
		rhythmLabels
	} from '$lib/api/recurring';
	import { cashflowOriginLabel } from '$lib/api/transactions';
	import type { RecurringItemDetail, RecurringTransaction } from '$lib/api/types';

	type Props = {
		open?: boolean;
		item: RecurringItemDetail | null;
		loading?: boolean;
		error?: string | null;
		/** The transaction id currently being unlinked, if any. */
		unlinking?: string | null;
		onUnlink?: (transaction: RecurringTransaction) => void;
		onEnd?: () => void;
		onClose?: () => void;
	};

	let {
		open = $bindable(false),
		item,
		loading = false,
		error = null,
		unlinking = null,
		onUnlink,
		onEnd,
		onClose
	}: Props = $props();

	const labels = $derived(item ? item.history.map((point) => point.date) : []);
	const amounts = $derived(
		item ? item.history.map((point) => scaledToNumber(point.amountCents)) : []
	);
	const status = $derived(
		!item
			? ''
			: item.ended_from
				? `Ended from ${endedFromLabel(item.ended_from)}`
				: `Next expected ${nextExpectedLabel(item.next_expected)}`
	);

	// `source` carries the origin the same way it does in Cashflow, so a row entered by
	// hand and a row from a statement are told apart in one word.
	function originLabel(transaction: RecurringTransaction): string {
		return cashflowOriginLabel({
			id: transaction.id,
			description: transaction.description,
			note: '',
			source: transaction.source,
			amountCents: transaction.amountCents,
			direction: 'out',
			date: transaction.date,
			tag: '',
			ignored: false
		});
	}
</script>

<Drawer bind:open title={item?.name ?? 'Recurring item'} width="max-w-2xl" {onClose}>
	{#snippet header()}
		{#if item}
			<div class="mt-1 flex flex-wrap items-center gap-2">
				<Badge intent="neutral" variant="soft" size="sm">
					{recurringDirectionLabel(item.direction)}
				</Badge>
				<Badge intent="neutral" variant="outline" size="sm">{rhythmLabels[item.rhythm]}</Badge>
				<Text as="span" size="sm" tone="muted">{status}</Text>
			</div>
		{/if}
	{/snippet}

	{#if loading}
		<div class="flex flex-col gap-3">
			<Skeleton variant="rect" height="10rem" />
			<Skeleton width="60%" />
			<Skeleton width="40%" />
		</div>
	{:else if error}
		<p class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700" role="alert">
			{error}
		</p>
	{:else if item}
		<AnalyticsCard title="Amount over time">
			{#if amounts.length > 1}
				<TimeSeriesChart
					height="h-40"
					{labels}
					xTickFormat={recurringTickLabel}
					ariaLabel={`Amounts paid for ${item.name}, oldest to newest`}
					datasets={[{ label: 'Amount', data: amounts, color: chartColors.net, fill: true }]}
				/>
			{:else}
				<Text size="sm" tone="muted">
					One amount so far. A second one gives this item a line to show.
				</Text>
			{/if}

			{#if amounts.length > 0}
				<div class="mt-2 flex flex-wrap items-baseline justify-between gap-2">
					<Text as="span" size="xs" tone="muted">
						{amounts.length}
						{amounts.length === 1 ? 'amount' : 'amounts'}, {formatDisplayDate(
							labels[0].slice(0, 10)
						)} – {formatDisplayDate(labels[labels.length - 1].slice(0, 10))}
					</Text>
					<span class="flex items-baseline gap-2 text-xs text-slate-500">
						First <Money amount={amounts[0]} currency="EUR" size="sm" weight="normal" />
						· Last
						<Money amount={scaledToNumber(item.lastAmountCents)} currency="EUR" size="sm" />
					</span>
				</div>
			{/if}
		</AnalyticsCard>

		<AnalyticsCard title="Linked transactions" class="mt-5">
			<Text size="sm" tone="muted" class="mb-2">
				{item.transactions.length} transactions, newest first. Unlink one that does not belong — the
				transaction itself stays in Cashflow.
			</Text>

			{#if item.transactions.length === 0}
				<p class="py-6 text-center text-sm text-slate-500">
					No transactions linked yet. Mark one in Cashflow to add it here.
				</p>
			{:else}
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
						{#each item.transactions as transaction (transaction.id)}
							<tr class="border-b border-slate-100">
								<td class="py-2 pr-3 whitespace-nowrap text-slate-700">
									{formatDisplayDate(transaction.date.slice(0, 10))}
								</td>
								<td class="py-2 pr-3 text-slate-700">{transaction.description}</td>
								<td class="py-2 pr-3">
									<Badge intent="neutral" size="sm">{originLabel(transaction)}</Badge>
								</td>
								<td class="py-2 pr-3 text-right">
									<Money amount={scaledToNumber(transaction.amountCents)} currency="EUR" size="sm" />
								</td>
								<td class="py-2 text-right">
									<Button
										size="sm"
										variant="ghost"
										intent="secondary"
										shape="default"
										loading={unlinking === transaction.id}
										disabled={unlinking === transaction.id}
										onclick={() => onUnlink?.(transaction)}
									>
										Unlink
									</Button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</AnalyticsCard>
	{/if}

	{#snippet footer()}
		<Button variant="outline" shape="default" onclick={() => (open = false)}>Close</Button>
		{#if item && !item.ended_from}
			<Button variant="outline" intent="error" shape="default" onclick={onEnd}>
				End from month…
			</Button>
		{/if}
	{/snippet}
</Drawer>
