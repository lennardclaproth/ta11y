<script lang="ts">
	// Design prototype for issue #16, variant B ("Agenda").
	// Orders the page by *when* the next transaction is expected, with the suggestions
	// from your history as a side rail instead of a separate screen.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import StatCard from '$lib/components/organisms/stat-card/StatCard.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
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
	import { dayOfMonth, daysUntil, monthKey, monthLabel, shortMonth } from '../recurring.format';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';

	type Props = {
		situation?: 'default' | 'loading' | 'error' | 'empty';
	};

	let { situation = 'default' }: Props = $props();

	let view = $state('upcoming');

	const views = $derived([
		{ value: 'upcoming', label: 'Expected' },
		{ value: 'all', label: `All items (${recurringExpenses.length + recurringIncome.length})` },
		{ value: 'ended', label: `Ended (${recurringEnded.length})` }
	]);

	const scheduled = $derived(
		[...recurringExpenses, ...recurringIncome]
			.filter((item) => item.nextExpected)
			.sort((a, b) => (a.nextExpected ?? '').localeCompare(b.nextExpected ?? ''))
	);

	/** Agenda rows grouped per month, so the reading order is the calendar's. */
	const months = $derived(
		situation === 'default'
			? scheduled.reduce<{ key: string; label: string; items: RecurringItem[] }[]>((acc, item) => {
					const key = monthKey(item.nextExpected as string);
					const group = acc.find((g) => g.key === key);
					if (group) group.items.push(item);
					else acc.push({ key, label: monthLabel(item.nextExpected as string), items: [item] });
					return acc;
				}, [])
			: []
	);

	const running = $derived([...recurringExpenses, ...recurringIncome]);
	const runningCount = $derived(running.length);
	const perRhythm = $derived([
		{ label: 'Monthly', count: running.filter((i) => i.rhythm === 'monthly').length },
		{ label: 'Quarterly', count: running.filter((i) => i.rhythm === 'quarterly').length },
		{ label: 'Yearly', count: running.filter((i) => i.rhythm === 'yearly').length },
		{ label: 'Ended', count: recurringEnded.length }
	]);

	function distance(iso: string): string {
		const days = daysUntil(iso);
		if (days === 0) return 'today';
		if (days === 1) return 'tomorrow';
		return `in ${days} days`;
	}
</script>

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
			<div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
				<StatCard label="Expenses per month" amount={monthlyExpenseTotal} currency="EUR" />
				<StatCard label="Income per month" amount={monthlyIncomeTotal} currency="EUR" />
				<AnalyticsCard title="Running items">
					<dl class="flex flex-wrap items-baseline gap-x-6 gap-y-1 text-sm">
						{#each perRhythm as entry (entry.label)}
							<div class="flex items-baseline gap-2">
								<dt class="text-slate-500">{entry.label}</dt>
								<dd class="font-medium text-slate-900">{entry.count}</dd>
							</div>
						{/each}
					</dl>
				</AnalyticsCard>
			</div>
		{/snippet}

		<div
			class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
		>
			<div class="flex min-w-0 flex-wrap items-center gap-4">
				<h2 class="text-2xl">Recurring items</h2>
				<Tabs tabs={views} bind:value={view} size="sm" ariaLabel="Recurring views" />
			</div>
			<Text as="span" size="sm" tone="muted">{runningCount} running</Text>
		</div>

		<div class="grid min-h-0 flex-1 grid-cols-1 lg:grid-cols-3">
			<!-- The agenda: what is expected, in calendar order. -->
			<div class="min-h-0 overflow-auto bg-white lg:col-span-2">
				{#if situation === 'loading'}
					<div class="flex flex-col gap-4 p-4">
						{#each [0, 1, 2, 3, 4, 5] as row (row)}
							<Skeleton class="h-10" />
						{/each}
					</div>
				{:else if situation === 'error'}
					<div class="p-4">
						<Alert intent="error" title="Could not load your recurring items">
							Check your connection and try again.
						</Alert>
					</div>
				{:else if situation === 'empty'}
					<div class="px-6 py-16 text-center">
						<Text size="sm" tone="muted">
							No recurring items yet. Point at a transaction in Cashflow to start one.
						</Text>
					</div>
				{:else}
					{#each months as month (month.key)}
						<div
							class="sticky top-0 border-b border-slate-400 bg-white px-4 pt-4 pb-2 text-sm font-semibold text-slate-700"
						>
							{month.label}
						</div>
						<ul>
							{#each month.items as item (item.id)}
								<li>
									<button
										type="button"
										class="flex w-full items-center gap-4 border-b border-slate-100 px-4 py-3 text-left transition-colors hover:bg-slate-50 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
									>
										<span class="w-12 shrink-0 text-center">
											<span class="block text-xs tracking-wide text-slate-500 uppercase">
												{shortMonth(item.nextExpected as string)}
											</span>
											<span class="block font-heading text-xl text-slate-900">
												{dayOfMonth(item.nextExpected as string)}
											</span>
										</span>
										<span class="min-w-0 flex-1">
											<span class="block truncate font-medium text-slate-900">{item.name}</span>
											<span class="block text-xs text-slate-500">
												{rhythmLabels[item.rhythm]} · {item.direction === 'in'
													? 'Income'
													: 'Expense'} · expected {distance(item.nextExpected as string)}
											</span>
										</span>
										<span class="shrink-0 text-right">
											<Money
												amount={item.direction === 'in' ? item.amount : -item.amount}
												currency="EUR"
												size="sm"
												signDisplay="always"
												colored
											/>
										</span>
									</button>
								</li>
							{/each}
						</ul>
					{/each}
				{/if}
			</div>

			<!-- Suggestions stay beside the agenda: nothing here counts until it is confirmed. -->
			<aside class="min-h-0 overflow-auto border-slate-200 bg-taupe-50 lg:border-l">
				<div class="border-b border-slate-200 px-4 py-3">
					<Heading level="h3" size="sm">From your history</Heading>
					<Text size="sm" tone="muted" class="mt-1">
						{recurringSuggestions.length} possible recurring items. Nothing is added until you confirm
						it.
					</Text>
				</div>

				{#each recurringSuggestions as suggestion (suggestion.id)}
					<div class="border-b border-slate-200 px-4 py-3">
						<div class="flex items-start justify-between gap-3">
							<div class="min-w-0">
								<p class="truncate font-medium text-slate-900">{suggestion.name}</p>
								<p class="mt-0.5 text-xs text-slate-500">
									{rhythmLabels[suggestion.rhythm]} · {suggestion.matches} transactions since
									{formatDisplayDate(suggestion.since)}
								</p>
							</div>
							<Money amount={suggestion.amount} currency="EUR" size="sm" />
						</div>
						<p class="mt-2 truncate text-xs text-slate-500">“{suggestion.sample}”</p>
						<div class="mt-3 flex flex-wrap items-center gap-2">
							<Button size="sm">Confirm</Button>
							<Button size="sm" variant="outline">Adjust</Button>
							<Button size="sm" variant="ghost" intent="secondary">Dismiss</Button>
						</div>
					</div>
				{/each}

				<div class="px-4 py-3">
					<Text size="xs" tone="muted">
						Transactions you ignore are left out, and a dismissed suggestion does not come back.
					</Text>
				</div>
			</aside>
		</div>
	</PageContentTemplate>
</AppShellTemplate>
