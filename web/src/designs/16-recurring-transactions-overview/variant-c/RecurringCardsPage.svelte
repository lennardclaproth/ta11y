<script lang="ts">
	// Design prototype for issue #16, variant C ("Cards").
	// One panel per recurring item, so the amount history is readable without opening anything.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';
	import {
		monthlyExpenseTotal,
		monthlyIncomeTotal,
		recurringEnded,
		recurringExpenses,
		recurringIncome,
		recurringSuggestions,
		rhythmLabels,
		type RecurringItem
	} from '../recurring.fixture';

	type Props = {
		situation?: 'default' | 'empty';
	};

	let { situation = 'default' }: Props = $props();

	const kpis: KpiItem[] = [
		{ label: 'Recurring expenses per month', amount: monthlyExpenseTotal, currency: 'EUR' },
		{ label: 'Recurring income per month', amount: monthlyIncomeTotal, currency: 'EUR' }
	];
</script>

<!-- One item. The large value is the most recent amount; the first amount and the line beside it
     show how it got there, without calling that good or bad. -->
{#snippet itemCard(item: RecurringItem)}
	<Panel variant="default" shape="sm" padding="md" bordered interactive class="flex flex-col gap-3">
		<div class="flex items-start justify-between gap-3">
			<div class="min-w-0">
				<Heading level="h4" size="sm" class="truncate">{item.name}</Heading>
				<Text as="span" size="xs" tone="muted">{item.linked} transactions</Text>
			</div>
			<Badge intent="neutral" variant="soft" size="sm">{rhythmLabels[item.rhythm]}</Badge>
		</div>

		<div class="flex items-end justify-between gap-3">
			<Money amount={item.amount} currency="EUR" size="lg" weight="semibold" />
			<Sparkline
				data={item.history}
				tone="neutral"
				width={112}
				height={34}
				fill
				ariaLabel={`Amounts for ${item.name}, oldest to newest`}
			/>
		</div>

		<div
			class="flex flex-wrap items-center justify-between gap-x-3 gap-y-1 border-t border-slate-200 pt-2"
		>
			<span class="flex items-baseline gap-1 text-xs text-slate-500">
				First
				<Money amount={item.history[0]} currency="EUR" size="sm" weight="normal" class="text-xs" />
			</span>
			{#if item.endedFrom}
				<span class="text-xs text-slate-500">Ended from {item.endedFrom}</span>
			{:else}
				<span class="text-xs text-slate-700">
					Next {formatDisplayDate(item.nextExpected)}
				</span>
			{/if}
		</div>
	</Panel>
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Recurring"
			showSearch
			searchPlaceholder="Search name…"
			accountName="Account"
			accountEmail="you@example.com"
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			<KpiRow items={kpis} columns={2} />
		{/snippet}

		<div
			class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
		>
			<h2 class="text-2xl">Recurring items</h2>
			<Text as="span" size="sm" tone="muted">
				{recurringExpenses.length + recurringIncome.length} running · {recurringEnded.length} ended
			</Text>
		</div>

		<div class="min-h-0 flex-1 overflow-auto px-4 pb-6">
			{#if situation === 'empty'}
				<div class="px-6 py-16 text-center">
					<Text size="sm" tone="muted">
						No recurring items yet. Point at a transaction in Cashflow to start one.
					</Text>
				</div>
			{:else}
				<AnalyticsCard title="From your history">
					<Text size="sm" tone="muted" class="mb-3">
						{recurringSuggestions.length} possible recurring items, found in transactions you have not
						ignored. Nothing is added until you confirm it.
					</Text>
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
						{#each recurringSuggestions as suggestion (suggestion.id)}
							<Panel
								variant="muted"
								shape="sm"
								padding="md"
								bordered
								class="flex flex-col gap-3 border-dashed"
							>
								<div class="flex items-start justify-between gap-3">
									<div class="min-w-0">
										<Heading level="h4" size="sm" class="truncate">{suggestion.name}</Heading>
										<Text as="span" size="xs" tone="muted">
											{suggestion.matches} transactions since {formatDisplayDate(suggestion.since)}
										</Text>
									</div>
									<Badge intent="info" variant="soft" size="sm">
										{rhythmLabels[suggestion.rhythm]}
									</Badge>
								</div>
								<div class="flex items-end justify-between gap-3">
									<Money amount={suggestion.amount} currency="EUR" size="lg" weight="semibold" />
									<span class="truncate text-xs text-slate-500">“{suggestion.sample}”</span>
								</div>
								<div class="flex flex-wrap items-center gap-2 border-t border-slate-200 pt-2">
									<Button size="sm">Confirm</Button>
									<Button size="sm" variant="outline">Adjust</Button>
									<Button size="sm" variant="ghost" intent="secondary">Dismiss</Button>
								</div>
							</Panel>
						{/each}
					</div>
				</AnalyticsCard>

				<AnalyticsCard title="Expenses and subscriptions" class="mt-5">
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
						{#each recurringExpenses as item (item.id)}
							{@render itemCard(item)}
						{/each}
					</div>
				</AnalyticsCard>

				<AnalyticsCard title="Recurring income" class="mt-5">
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
						{#each recurringIncome as item (item.id)}
							{@render itemCard(item)}
						{/each}
					</div>
				</AnalyticsCard>

				<AnalyticsCard title="Ended" class="mt-5">
					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
						{#each recurringEnded as item (item.id)}
							{@render itemCard(item)}
						{/each}
					</div>
				</AnalyticsCard>
			{/if}
		</div>
	</PageContentTemplate>
</AppShellTemplate>
