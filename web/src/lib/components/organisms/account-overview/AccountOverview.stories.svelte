<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import AccountOverview from './AccountOverview.svelte';
	import { accountOverviewLayouts } from './account-overview.types';
	import type { DateRangePreset } from '$lib/components/molecules/date-range-picker/date-range-picker.types';

	const { Story } = defineMeta({
		title: 'Organisms/AccountOverview',
		component: AccountOverview,
		tags: ['autodocs'],
		argTypes: {
			layout: { control: 'select', options: accountOverviewLayouts },
			showPeriod: { control: 'boolean' },
			loading: { control: 'boolean' }
		}
	});

	// Fictional figures: this story documents the layout, not an account.
	const presets: DateRangePreset[] = [
		{ value: '1m', label: '1M', from: '2026-06-01', to: '2026-06-30' },
		{ value: '3m', label: '3M', from: '2026-04-01', to: '2026-06-30' },
		{ value: '6m', label: '6M', from: '2026-01-01', to: '2026-06-30' },
		{ value: 'ytd', label: 'YTD', from: '2026-01-01', to: '2026-06-30' },
		{ value: '1y', label: '1Y', from: '2025-07-01', to: '2026-06-30' },
		{ value: '3y', label: '3Y', from: '2023-07-01', to: '2026-06-30' },
		{ value: '5y', label: '5Y', from: '2021-07-01', to: '2026-06-30' },
		{ value: 'max', label: 'Max', from: '2021-01-01', to: '2026-06-30' }
	];

	const base = {
		accountName: 'Sam Rivers',
		accountEmail: 'sam@example.com',
		netWorth: 152781.15,
		changePct: 7.36,
		periodFrom: '2026-01-01',
		periodTo: '2026-06-30',
		periodPreset: 'ytd',
		periodPresets: presets
	};
</script>

<!--
  The overview is inline in the masthead from `lg` and behind one entry below it, with the same
  content either way. The inline story is hidden under 1024px by the component itself.
-->

<Story name="Inline" asChild>
	<div class="flex justify-end bg-taupe-100 p-4">
		<AccountOverview {...base} />
	</div>
</Story>

<Story name="Panel" asChild>
	<div class="flex justify-end bg-taupe-100 p-4">
		<AccountOverview {...base} layout="panel" />
	</div>
</Story>

<Story name="Loading" asChild>
	<div class="flex justify-end bg-taupe-100 p-4">
		<AccountOverview {...base} layout="panel" loading />
	</div>
</Story>

<!-- No snapshots at all: a dash, never a zero standing in for an unknown amount. -->
<Story name="No snapshots" asChild>
	<div class="flex justify-end bg-taupe-100 p-4">
		<AccountOverview {...base} layout="panel" netWorth={null} changePct={null} />
	</div>
</Story>

<!-- A period the snapshots do not cover: the worth is known, the change is not. -->
<Story name="No change in period" asChild>
	<div class="flex justify-end bg-taupe-100 p-4">
		<AccountOverview {...base} layout="panel" changePct={null} />
	</div>
</Story>

<!-- Admin pages have no charts and no period, so the worth is labelled as the latest snapshot. -->
<Story name="Without a period" asChild>
	<div class="flex justify-end bg-taupe-100 p-4">
		<AccountOverview {...base} layout="panel" showPeriod={false} />
	</div>
</Story>
