<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import type { Pathname } from '$app/types';
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import NavMenu from '$lib/components/molecules/nav-menu/NavMenu.svelte';
	import AccountOverview from '$lib/components/organisms/account-overview/AccountOverview.svelte';
	import type { NavItem } from '$lib/components/molecules/nav-menu/nav-menu.types';
	import { zClasses } from '$lib/styles/z-index';
	import { accountStore } from '$lib/stores/account.svelte';
	import { netWorthStore } from '$lib/stores/net-worth.svelte';
	import { periodStore, type PeriodPreset } from '$lib/stores/period.svelte';

	/**
	 * The bar carries navigation only: the wordmark and the destinations. Searching, the period
	 * and the page's own actions have moved to the table and the charts they act on, so nothing
	 * in here changes what is on the page. Everything personal — who is signed in, what the
	 * account is worth, which period you are looking at — sits to the right of a rule as its own
	 * region, and behind a single entry on a narrow screen.
	 */
	type Props = {
		/** Admin pages have no charts and no period, so they hide the period control. */
		showPeriod?: boolean;
		class?: string;
	};

	let { showPeriod = true, class: className = '' }: Props = $props();

	let overviewOpen = $state(false);
	let overviewTrigger = $state<HTMLButtonElement | null>(null);

	$effect(() => {
		void netWorthStore.ensureLoaded();
	});

	/**
	 * Escape closes the narrow-screen overview and hands focus back to the entry that opened it.
	 * It is deliberately not a `Popover`: the period picker inside it is one, and it portals its
	 * panel outside this one, so an outside-click rule here would close the overview the moment
	 * you reached for a date.
	 */
	$effect(() => {
		if (!overviewOpen) return;
		function onKeydown(event: KeyboardEvent) {
			if (event.key !== 'Escape') return;
			overviewOpen = false;
			overviewTrigger?.focus();
		}
		document.addEventListener('keydown', onKeydown, true);
		return () => document.removeEventListener('keydown', onKeydown, true);
	});

	const accountName = $derived(accountStore.email || 'Account');
	const accountEmail = $derived(accountStore.email || undefined);

	// Shared destinations for the desktop masthead and the narrow-screen menu. Admin
	// destinations are omitted entirely for non-admin accounts rather than shown and
	// then refused, so the navigation only ever offers reachable pages.
	const primaryItems: (NavItem & { href: Pathname })[] = [
		{ label: 'Cashflow', href: '/cashflow' as Pathname, icon: 'heroicons:banknotes' },
		{ label: 'Assets', href: '/assets' as Pathname, icon: 'heroicons:building-library' },
		{ label: 'Portfolio', href: '/portfolio' as Pathname, icon: 'heroicons:chart-pie' }
	];

	const adminItems: (NavItem & { href: Pathname })[] = [
		{ label: 'Listings', href: '/admin/listings' as Pathname, icon: 'heroicons:cog-6-tooth' },
		{ label: 'Dailies', href: '/admin/dailies' as Pathname, icon: 'heroicons:calendar-days' },
		{ label: 'Credentials', href: '/admin/credentials' as Pathname, icon: 'heroicons:key' }
	];

	const visibleAdminItems = $derived(accountStore.isAdmin ? adminItems : []);

	const navItems: (NavItem & { href: Pathname })[] = $derived([
		...primaryItems,
		...visibleAdminItems.map((item, index) => (index === 0 ? { ...item, divider: true } : item))
	]);

	const isActive = (href: string) =>
		page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);

	function changePeriod(range: { from: string; to: string; preset: string }) {
		periodStore.set({ ...range, preset: range.preset as PeriodPreset });
	}

	function signOut() {
		void accountStore.signOut();
	}
</script>

{#snippet overview(layout: 'inline' | 'panel')}
	<AccountOverview
		{layout}
		{accountName}
		{accountEmail}
		netWorth={netWorthStore.worth}
		changePct={netWorthStore.changePct}
		loading={netWorthStore.loading}
		failed={netWorthStore.failed}
		{showPeriod}
		periodFrom={periodStore.from}
		periodTo={periodStore.to}
		periodPreset={periodStore.preset}
		periodPresets={periodStore.options}
		onPeriodChange={changePeriod}
		onSignOut={signOut}
	/>
{/snippet}

<header class={['px-4 pt-3 lg:px-8', className].filter(Boolean).join(' ')}>
	<div class="relative flex items-center justify-between gap-6 border-b border-slate-800 pb-2">
		<!-- Navigation column: wordmark above, destinations below. -->
		<div class="flex min-w-0 flex-col gap-2">
			<a
				href={resolve('/cashflow')}
				aria-label="ta11y home"
				class="font-heading text-4xl leading-none tracking-tight text-slate-900 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700"
				>ta11y</a
			>
			<nav aria-label="Primary navigation" class="hidden items-center gap-5 md:flex">
				{#each primaryItems as item (item.href)}
					<a
						href={resolve(item.href)}
						aria-current={isActive(item.href) ? 'page' : undefined}
						class="border-b-2 border-transparent pb-0.5 text-sm text-slate-700 transition-colors hover:border-slate-400 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700 aria-[current=page]:border-amber-500 aria-[current=page]:font-semibold aria-[current=page]:text-slate-950"
						>{item.label}</a
					>
				{/each}
				{#if visibleAdminItems.length > 0}
					<span class="h-4 w-px bg-slate-300" aria-hidden="true"></span>
					{#each visibleAdminItems as item (item.href)}
						<a
							href={resolve(item.href)}
							aria-current={isActive(item.href) ? 'page' : undefined}
							class="border-b-2 border-transparent pb-0.5 text-sm text-slate-500 transition-colors hover:border-slate-400 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700 aria-[current=page]:border-amber-500 aria-[current=page]:font-semibold aria-[current=page]:text-slate-950"
							>{item.label}</a
						>
					{/each}
				{/if}
			</nav>
		</div>

		<!-- Your overview, open in the masthead from `lg`. -->
		{@render overview('inline')}

		<!-- Narrow screens: destinations behind the menu, the same overview behind one entry. -->
		<div class="flex items-center gap-1 lg:hidden">
			<div class="md:hidden">
				<NavMenu items={navItems} currentPath={page.url.pathname} triggerLabel="Menu" />
			</div>
			<button
				bind:this={overviewTrigger}
				type="button"
				aria-expanded={overviewOpen}
				aria-controls="account-overview-panel"
				aria-label="Your overview"
				onclick={() => (overviewOpen = !overviewOpen)}
				class="inline-flex h-11 items-center gap-1 rounded-lg px-2 text-slate-700 transition-colors hover:bg-slate-100 focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none"
			>
				<Avatar name={accountName} size="sm" />
				<Icon icon="heroicons:chevron-down" size="sm" />
			</button>
		</div>

		{#if overviewOpen}
			<div
				id="account-overview-panel"
				class={['absolute top-full right-0 mt-2 lg:hidden', zClasses.popover].join(' ')}
			>
				{@render overview('panel')}
			</div>
		{/if}
	</div>
</header>
