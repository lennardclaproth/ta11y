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
	import { accountStore } from '$lib/stores/account.svelte';

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
		/** Defaults to the signed-in account's email. */
		accountName?: string;
		accountEmail?: string;

		accountItems?: MenuItem[];
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
		accountItems = [],
		class: className = ''
	}: Props = $props();

	// The signed-in account labels the menu, so no screen has to be told who is using it.
	const resolvedAccountName = $derived(accountName ?? accountStore.email ?? 'Account');
	const resolvedAccountEmail = $derived(accountEmail ?? accountStore.email ?? undefined);

	// Sign out is always available, and always last. A caller supplying its own item of
	// the same label is dropped rather than duplicated: the menu keys its rows by label,
	// and a duplicate key throws and renders nothing at all.
	const resolvedAccountItems: MenuItem[] = $derived([
		...accountItems.filter((item) => item.label !== 'Sign out'),
		{
			label: 'Sign out',
			icon: 'heroicons:arrow-right-start-on-rectangle',
			intent: 'danger' as const,
			divider: accountItems.length > 0,
			onSelect: () => {
				void accountStore.signOut();
			}
		}
	]);

	// Shared destinations for the desktop masthead and the narrow-screen menu. Admin
	// destinations are omitted entirely for non-admin accounts rather than shown and
	// then refused, so the navigation only ever offers reachable pages.
	const navItems: (NavItem & { href: Pathname })[] = $derived([
		{ label: 'Cashflow', href: '/cashflow' as Pathname, icon: 'heroicons:banknotes' },
		{ label: 'Assets', href: '/assets' as Pathname, icon: 'heroicons:building-library' },
		{ label: 'Portfolio', href: '/portfolio' as Pathname, icon: 'heroicons:chart-pie' },
		...(accountStore.isAdmin
			? [
					{
						label: 'Listings',
						href: '/admin/listings' as Pathname,
						icon: 'heroicons:cog-6-tooth',
						divider: true
					},
					{
						label: 'Dailies',
						href: '/admin/dailies' as Pathname,
						icon: 'heroicons:calendar-days'
					},
					{
						label: 'Credentials',
						href: '/admin/credentials' as Pathname,
						icon: 'heroicons:key'
					}
				]
			: [])
	]);
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
			<AccountMenu name={resolvedAccountName} email={resolvedAccountEmail} items={resolvedAccountItems} />
		</div>
	</div>
</header>
