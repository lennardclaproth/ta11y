<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import type { Pathname } from '$app/types';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import DateRangePicker from '$lib/components/molecules/date-range-picker/DateRangePicker.svelte';
	import ActionMenu from '$lib/components/molecules/action-menu/ActionMenu.svelte';
	import AccountMenu from '$lib/components/molecules/account-menu/AccountMenu.svelte';
	import NavMenu from '$lib/components/molecules/nav-menu/NavMenu.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';
	import type { NavItem } from '$lib/components/molecules/nav-menu/nav-menu.types';

	type DateRange = { from: string | null; to: string | null };

	type Props = {
		title?: string;
		showSearch?: boolean;
		searchValue?: string;
		searchPlaceholder?: string;
		onSearch?: (query: string) => void;
		showDateRange?: boolean;
		dateFrom?: string | null;
		dateTo?: string | null;
		onDateChange?: (range: DateRange) => void;
		actions?: MenuItem[];
		accountName: string;
		accountEmail?: string;
		adminMode?: boolean;
		showAdminToggle?: boolean;
		accountItems?: MenuItem[];
		onAdminToggle?: (value: boolean) => void;
		class?: string;
	};

	let {
		title,
		showSearch = false,
		searchValue = $bindable(''),
		searchPlaceholder = 'Search…',
		onSearch,
		showDateRange = false,
		dateFrom = $bindable(null),
		dateTo = $bindable(null),
		onDateChange,
		actions = [],
		accountName,
		accountEmail,
		adminMode = $bindable(false),
		showAdminToggle = true,
		accountItems = [],
		onAdminToggle,
		class: className = ''
	}: Props = $props();

	// Shared destinations for the desktop masthead and the narrow-screen menu.
	const navItems: (NavItem & { href: Pathname })[] = [
		{ label: 'Cashflow', href: '/cashflow', icon: 'heroicons:banknotes' },
		{ label: 'Assets', href: '/assets', icon: 'heroicons:building-library' },
		{ label: 'Portfolio', href: '/portfolio', icon: 'heroicons:chart-pie' },
		{ label: 'Listings', href: '/admin/listings', icon: 'heroicons:cog-6-tooth', divider: true }
	];
</script>

<header class={['px-4 pt-3 pb-4 lg:px-8', className].filter(Boolean).join(' ')}>
	<div class="flex items-center justify-between gap-4 border-t-2 border-b border-slate-800 py-2">
		<a
			href={resolve('/cashflow')}
			aria-label="ta11y home"
			class="font-heading text-5xl leading-none tracking-tight text-slate-900 underline-offset-4 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700"
			>ta11y</a
		>
		<nav aria-label="Primary navigation" class="hidden items-center gap-6 md:flex">
			{#each navItems as item (item.href)}
				<a
					href={resolve(item.href)}
					aria-current={page.url.pathname === item.href ||
					page.url.pathname.startsWith(`${item.href}/`)
						? 'page'
						: undefined}
					class="border-b-2 border-transparent py-3 text-sm text-slate-700 transition-colors hover:border-slate-400 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700 aria-[current=page]:border-amber-500 aria-[current=page]:font-semibold aria-[current=page]:text-slate-950"
					>{item.label}</a
				>
			{/each}
		</nav>
		<div class="md:hidden">
			<NavMenu items={navItems} currentPath={page.url.pathname} triggerLabel="Menu" />
		</div>
	</div>

	<div class="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 pt-4">
		{#if title}<Heading level="h1" size="2xl" class="leading-none">{title}</Heading>{/if}

		<div class="order-3 w-full md:order-none md:min-w-48 md:flex-1">
			{#if showSearch}
				<div class="w-full md:ml-auto md:max-w-md">
					<SearchInput
						bind:value={searchValue}
						{onSearch}
						placeholder={searchPlaceholder}
						size="md"
						shape="default"
					/>
				</div>
			{/if}
		</div>

		<div class="flex max-w-full flex-wrap items-center gap-2 md:shrink-0">
			{#if showDateRange}
				<DateRangePicker bind:from={dateFrom} bind:to={dateTo} size="md" onChange={onDateChange} />
			{/if}
			{#if actions.length > 0}
				<ActionMenu items={actions} />
			{/if}
			<AccountMenu
				name={accountName}
				email={accountEmail}
				bind:adminMode
				{showAdminToggle}
				items={accountItems}
				{onAdminToggle}
			/>
		</div>
	</div>
</header>
