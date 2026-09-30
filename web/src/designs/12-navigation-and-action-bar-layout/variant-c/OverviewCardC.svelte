<script lang="ts">
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import { icons } from '../shared/icons';
	import { account, netWorth, netWorthChange, netWorthChangePct, period } from '../shared/mock-data';

	/** Admin pages have no period, so that block drops out. */
	let { showPeriod = true }: { showPeriod?: boolean } = $props();

	let range = $state('6m');
</script>

<Panel variant="floating" shape="sm" shadow="md" padding="none" class="w-80 sm:w-96">
	<div class="p-4">
		<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">Net worth</Text>
		<div class="mt-1 flex flex-wrap items-baseline gap-x-3 gap-y-1">
			<Money amount={netWorth} currency="EUR" size="xl" weight="semibold" />
			<TrendIndicator value={netWorthChangePct} size="sm" />
		</div>
		<Text as="p" size="xs" tone="muted">
			{netWorthChange >= 0 ? '+' : '−'}€{Math.abs(netWorthChange).toLocaleString('en', {
				maximumFractionDigits: 0
			})} over {period.label}
		</Text>
	</div>

	{#if showPeriod}
		<div class="border-t border-slate-300 p-4">
			<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">Period</Text>
			<div class="mt-2">
				<Tabs
					tabs={[
						{ value: '1m', label: '1M' },
						{ value: '3m', label: '3M' },
						{ value: '6m', label: '6M' },
						{ value: 'ytd', label: 'YTD' },
						{ value: 'custom', label: 'Custom' }
					]}
					bind:value={range}
					size="sm"
					ariaLabel="Period"
				/>
			</div>
			<Text as="p" size="xs" tone="subtle" class="mt-2">
				{period.label} · applies to charts and tables on every page
			</Text>
		</div>
	{/if}

	<div class="flex items-center justify-between gap-3 border-t border-slate-300 p-4">
		<div class="flex min-w-0 items-center gap-2">
			<Avatar name={account.name} size="sm" />
			<div class="min-w-0">
				<p class="truncate text-sm font-medium text-slate-900">{account.name}</p>
				<p class="truncate text-xs text-slate-500">{account.email}</p>
			</div>
		</div>
		<button
			type="button"
			class="inline-flex shrink-0 items-center gap-1 border-b border-slate-300 pb-0.5 text-sm text-slate-700 transition-colors hover:border-slate-700 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700"
		>
			<Icon icon={icons.signOut} size="sm" />Sign out
		</button>
	</div>
</Panel>
