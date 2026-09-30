<script lang="ts">
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import TrendIndicator from '$lib/components/molecules/trend-indicator/TrendIndicator.svelte';
	import DateRangePicker from '$lib/components/molecules/date-range-picker/DateRangePicker.svelte';
	import AccountMenu from '$lib/components/molecules/account-menu/AccountMenu.svelte';
	import NavMenu from '$lib/components/molecules/nav-menu/NavMenu.svelte';
	import { icons } from '../shared/icons';
	import { account, navItems, netWorth, netWorthChangePct, period } from '../shared/mock-data';

	type Props = {
		activeHref: string;
		/** Narrow screens keep the same overview behind one entry; open it for the prototype. */
		overviewOpen?: boolean;
	};

	let { activeHref, overviewOpen = false }: Props = $props();

	const isActive = (href: string) => activeHref === href || activeHref.startsWith(`${href}/`);
</script>

<header class="px-4 pt-3 pb-4 lg:px-8">
	<div
		class="flex items-center justify-between gap-6 border-t-2 border-b border-slate-800 py-2 lg:gap-8"
	>
		<a
			href="/cashflow"
			aria-label="ta11y home"
			class="font-heading text-5xl leading-none tracking-tight text-slate-900 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700"
			>ta11y</a
		>

		<nav aria-label="Primary navigation" class="hidden flex-1 items-center gap-6 md:flex">
			{#each navItems as item (item.href)}
				<a
					href={item.href}
					aria-current={isActive(item.href) ? 'page' : undefined}
					class="border-b-2 border-transparent py-3 text-sm text-slate-700 transition-colors hover:border-slate-400 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700 aria-[current=page]:border-amber-500 aria-[current=page]:font-semibold aria-[current=page]:text-slate-950"
					>{item.label}</a
				>
			{/each}
		</nav>

		<!-- Your overview: fully visible on wide screens, no menu to open. -->
		<div class="hidden items-center gap-4 lg:flex">
			<div class="flex flex-col items-end border-l border-slate-300 pl-4">
				<Text as="span" size="xs" tone="muted">Net worth</Text>
				<div class="flex items-baseline gap-2">
					<Money amount={netWorth} currency="EUR" size="lg" weight="semibold" />
					<TrendIndicator value={netWorthChangePct} size="sm" />
				</div>
			</div>
			<DateRangePicker from={period.from} to={period.to} size="md" showPresets />
			<AccountMenu
				name={account.name}
				email={account.email}
				items={[{ label: 'Sign out', icon: 'heroicons:arrow-right-start-on-rectangle' }]}
			/>
		</div>

		<div class="flex items-center gap-1 lg:hidden">
			<div class="md:hidden">
				<NavMenu items={navItems} currentPath={activeHref} triggerLabel="Menu" />
			</div>
			<button
				type="button"
				aria-expanded={overviewOpen}
				aria-controls="overview-panel-a"
				class="inline-flex h-9 items-center gap-2 rounded-lg px-2 text-slate-700 transition-colors hover:bg-slate-100 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
			>
				<Avatar name={account.name} size="sm" />
				<Icon icon={icons.chevronDown} size="sm" />
			</button>
		</div>
	</div>

	{#if overviewOpen}
		<div
			id="overview-panel-a"
			class="mt-2 border-b border-slate-300 bg-white p-4 lg:hidden"
			aria-label="Your overview"
		>
			<div class="flex items-center gap-3">
				<Avatar name={account.name} size="md" />
				<div class="min-w-0">
					<p class="truncate text-sm font-medium text-slate-900">{account.name}</p>
					<p class="truncate text-xs text-slate-500">{account.email}</p>
				</div>
			</div>

			<div class="mt-3 flex items-end justify-between gap-3 border-t border-slate-200 pt-3">
				<Text as="span" size="sm" tone="muted">Net worth</Text>
				<div class="flex flex-col items-end">
					<Money amount={netWorth} currency="EUR" size="lg" weight="semibold" />
					<TrendIndicator value={netWorthChangePct} size="sm" />
				</div>
			</div>

			<div class="mt-3 border-t border-slate-200 pt-3">
				<Text as="span" size="sm" tone="muted">Period</Text>
				<div class="mt-1">
					<DateRangePicker from={period.from} to={period.to} size="lg" class="w-full" />
				</div>
			</div>

			<div class="mt-3 border-t border-slate-200 pt-3">
				<Button variant="ghost" intent="error" size="md" shape="default" class="w-full">
					<Icon icon={icons.signOut} />Sign out
				</Button>
			</div>
		</div>
	{/if}
</header>
