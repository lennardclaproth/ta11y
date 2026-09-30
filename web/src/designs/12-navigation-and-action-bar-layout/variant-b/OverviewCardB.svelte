<script lang="ts">
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import DateRangePicker from '$lib/components/molecules/date-range-picker/DateRangePicker.svelte';
	import { icons } from '../shared/icons';
	import { account, netWorth, netWorthChange, netWorthChangePct, period } from '../shared/mock-data';

	/** Same content on every screen width; only the anchor differs. */
	let { showPeriod = true }: { showPeriod?: boolean } = $props();
</script>

<Panel variant="floating" shape="md" shadow="md" padding="none" class="w-80">
	<div class="flex items-center gap-3 p-4">
		<Avatar name={account.name} size="md" />
		<div class="min-w-0">
			<p class="truncate text-sm font-medium text-slate-900">{account.name}</p>
			<p class="truncate text-xs text-slate-500">{account.email}</p>
		</div>
	</div>

	<div class="border-t border-slate-200 p-4">
		<Text as="span" size="xs" tone="muted">Net worth</Text>
		<div class="mt-1">
			<Money amount={netWorth} currency="EUR" size="xl" weight="semibold" />
		</div>
		<div class="mt-1 flex items-center gap-2">
			<TrendIndicator value={netWorthChangePct} size="sm" />
			<Text as="span" size="xs" tone="muted">
				{netWorthChange >= 0 ? '+' : '−'}€{Math.abs(netWorthChange).toLocaleString('en', {
					maximumFractionDigits: 0
				})} since {period.shortLabel.split(' – ')[0]}
			</Text>
		</div>
	</div>

	{#if showPeriod}
		<div class="border-t border-slate-200 p-4">
			<Text as="span" size="xs" tone="muted">Period</Text>
			<div class="mt-1">
				<DateRangePicker from={period.from} to={period.to} size="md" class="w-full" />
			</div>
			<Text as="span" size="xs" tone="subtle">Applies to charts and tables on every page.</Text>
		</div>
	{/if}

	<div class="border-t border-slate-200 p-2">
		<Button variant="ghost" intent="error" size="md" shape="default" class="w-full">
			<Icon icon={icons.signOut} />Sign out
		</Button>
	</div>
</Panel>
