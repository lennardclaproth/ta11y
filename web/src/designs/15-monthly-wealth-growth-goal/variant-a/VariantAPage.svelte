<script lang="ts">
	// Variant A — "Ledger tab": the monthly standing is a second view on Cashflow, in the same
	// ruled ledger header Portfolio already uses for its view tabs.
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import LedgerToolbar from '$lib/components/organisms/ledger-toolbar/LedgerToolbar.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import ProgressBar from '$lib/components/atoms/progress-bar/ProgressBar.svelte';
	import MonthStandingTable from './MonthStandingTable.svelte';
	import PurposeTransactionsTable from '../shared/PurposeTransactionsTable.svelte';
	import CashflowAnalytics from '../shared/CashflowAnalytics.svelte';
	import {
		bestStreak,
		currentStreak,
		goalAmount,
		goalMonths,
		goalTransactions,
		sharePercent
	} from '../goal-data';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	type Props = { tab?: 'transactions' | 'goal' };

	let { tab = $bindable('goal') }: Props = $props();

	let selectedIds = $state<string[]>(['px-04', 'px-05']);

	const running = goalMonths[0];
	const navActions: MenuItem[] = [{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' }];
	const tabs = [
		{ value: 'transactions', label: 'Transactions' },
		{ value: 'goal', label: 'Monthly goal' }
	];

	const figureClass = 'font-heading text-3xl leading-tight tracking-tight text-slate-900';
</script>

<!-- The shell is viewport-tall in the app; in the prototype it grows with its content so a
     screenshot shows the whole screen instead of the first fold. -->
<AppShellTemplate class="h-auto! min-h-dvh">
	{#snippet top()}
		<TopNavbar
			title="Cashflow"
			showSearch
			showDateRange
			actions={navActions}
			accountName="Design preview"
			accountEmail="preview@example.com"
		/>
	{/snippet}

	<PageContentTemplate>
		{#snippet analytics()}
			{#if tab === 'goal'}
				<div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
					<AnalyticsCard title="July 2026 so far">
						<div class="flex flex-col gap-2">
							<div class="flex items-baseline gap-2">
								<Money amount={running.contributed} currency="EUR" size="xl" weight="semibold" />
								<Text as="span" size="sm" tone="muted">
									of €{goalAmount(running).toLocaleString('en')} needed
								</Text>
							</div>
							<ProgressBar
								value={running.contributed}
								max={goalAmount(running)}
								intent="primary"
								ariaLabel="Share of July 2026 income put towards wealth"
							/>
							<Text size="sm" tone="muted">
								{sharePercent(running)}% of €{running.income.toLocaleString('en')} marked income ·
								goal {running.goalPercent}%
							</Text>
						</div>
					</AnalyticsCard>

					<AnalyticsCard title="Monthly goal">
						<div class="flex items-start justify-between gap-4">
							<div class="flex flex-col gap-1">
								<span class={figureClass}>{running.goalPercent}%</span>
								<Text size="sm" tone="muted">of your income, from July 2026</Text>
							</div>
							<Button size="sm" variant="outline">Adjust goal</Button>
						</div>
					</AnalyticsCard>

					<AnalyticsCard title="Streak">
						<div class="flex flex-col gap-1">
							<span class={figureClass}>{currentStreak} months in a row</span>
							<Text size="sm" tone="muted">Best run so far: {bestStreak} months</Text>
							<Text size="sm" tone="muted">July is still running, so it does not count yet.</Text>
						</div>
					</AnalyticsCard>
				</div>
			{:else}
				<CashflowAnalytics />
			{/if}
		{/snippet}

		<LedgerToolbar actionLabel="Add transaction" onAdd={() => {}}>
			<Tabs {tabs} bind:value={tab} ariaLabel="Cashflow view" />
			{#if tab === 'transactions'}
				<Button size="sm" variant="ghost" intent="secondary">Mark all 24 matches…</Button>
			{/if}
		</LedgerToolbar>

		{#if tab === 'goal'}
			<MonthStandingTable months={goalMonths} />
		{:else}
			<PurposeTransactionsTable rows={goalTransactions} bind:selectedIds />
		{/if}
	</PageContentTemplate>
</AppShellTemplate>
