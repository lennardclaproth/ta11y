<script lang="ts">
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import AccountMenu from '$lib/components/molecules/account-menu/AccountMenu.svelte';
	import NetWorthBlock from './NetWorthBlock.svelte';
	import PeriodControl from './PeriodControl.svelte';
	import { icons } from '../shared/icons';
	import { account } from '../shared/mock-data';

	/**
	 * "Your overview": account, net worth with its change over the period, and the one period
	 * choice. Proposed as a new organism. `inline` is the wide-screen region in the masthead;
	 * `panel` is the same content behind the single entry on a narrow screen.
	 *
	 * Admin pages pass `showPeriod={false}`: they have no charts and no period.
	 */
	type Props = {
		layout?: 'inline' | 'panel';
		showPeriod?: boolean;
		loading?: boolean;
		change?: 'value' | 'no-snapshot' | 'no-period' | 'none';
	};

	let { layout = 'inline', showPeriod = true, loading = false, change = 'value' }: Props =
		$props();
</script>

{#if layout === 'inline'}
	<div
		aria-label="Your overview"
		class="hidden items-center gap-6 border-l border-slate-300 pl-6 lg:flex"
	>
		<NetWorthBlock {loading} {change} />
		{#if showPeriod}
			<PeriodControl />
		{/if}
		<AccountMenu
			name={account.name}
			email={account.email}
			items={[
				{ label: 'Sign out', icon: 'heroicons:arrow-right-start-on-rectangle' }
			]}
		/>
	</div>
{:else}
	<Panel variant="floating" shape="sm" shadow="md" padding="none" class="w-80 sm:w-96">
		<div class="flex items-center gap-3 p-4">
			<Avatar name={account.name} size="md" />
			<div class="min-w-0">
				<p class="truncate text-sm font-medium text-slate-900">{account.name}</p>
				<p class="truncate text-xs text-slate-500">{account.email}</p>
			</div>
		</div>

		<div class="border-t border-slate-300 p-4">
			<NetWorthBlock align="start" {loading} {change} />
		</div>

		{#if showPeriod}
			<div class="border-t border-slate-300 p-4">
				<Text as="span" size="xs" tone="muted" class="tracking-wide uppercase">Period</Text>
				<div class="mt-2">
					<PeriodControl layout="block" showCaption />
				</div>
			</div>
		{/if}

		<div class="border-t border-slate-300 p-4">
			<Button variant="ghost" intent="error" size="lg" shape="default" class="w-full">
				<Icon icon={icons.signOut} />Sign out
			</Button>
		</div>
	</Panel>
{/if}
