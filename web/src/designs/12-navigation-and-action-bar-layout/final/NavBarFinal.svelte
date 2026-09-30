<script lang="ts">
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import NavMenu from '$lib/components/molecules/nav-menu/NavMenu.svelte';
	import { zClasses } from '$lib/styles/z-index';
	import AccountOverview from './AccountOverview.svelte';
	import { icons } from '../shared/icons';
	import { account, navItems } from '../shared/mock-data';

	/**
	 * The bar carries navigation only: the wordmark and the destinations. Everything personal
	 * sits to the right of a vertical rule, as its own region, so nothing that acts on the page
	 * lives in the bar any more.
	 */
	type Props = {
		activeHref: string;
		showPeriod?: boolean;
		loading?: boolean;
		change?: 'value' | 'no-snapshot' | 'no-period' | 'none';
		/** Opens the narrow-screen overview so the prototype can show it. */
		overviewOpen?: boolean;
	};

	let {
		activeHref,
		showPeriod = true,
		loading = false,
		change = 'value',
		overviewOpen = false
	}: Props = $props();

	const isActive = (href: string) => activeHref === href || activeHref.startsWith(`${href}/`);
	const primary = navItems.filter((item) => !item.href.startsWith('/admin'));
	const admin = navItems.filter((item) => item.href.startsWith('/admin'));
</script>

<header class="px-4 pt-3 lg:px-8">
	<div class="relative flex items-center justify-between gap-6 border-b border-slate-800 pb-2">
		<!-- Navigation column: wordmark above, destinations below. -->
		<div class="flex min-w-0 flex-col gap-2">
			<a
				href="/cashflow"
				aria-label="ta11y home"
				class="font-heading text-4xl leading-none tracking-tight text-slate-900 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700"
				>ta11y</a
			>
			<nav aria-label="Primary navigation" class="hidden items-center gap-5 md:flex">
				{#each primary as item (item.href)}
					<a
						href={item.href}
						aria-current={isActive(item.href) ? 'page' : undefined}
						class="border-b-2 border-transparent pb-0.5 text-sm text-slate-700 transition-colors hover:border-slate-400 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700 aria-[current=page]:border-amber-500 aria-[current=page]:font-semibold aria-[current=page]:text-slate-950"
						>{item.label}</a
					>
				{/each}
				<span class="h-4 w-px bg-slate-300" aria-hidden="true"></span>
				{#each admin as item (item.href)}
					<a
						href={item.href}
						aria-current={isActive(item.href) ? 'page' : undefined}
						class="border-b-2 border-transparent pb-0.5 text-sm text-slate-500 transition-colors hover:border-slate-400 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700 aria-[current=page]:border-amber-500 aria-[current=page]:font-semibold aria-[current=page]:text-slate-950"
						>{item.label}</a
					>
				{/each}
			</nav>
		</div>

		<!-- Your overview, always visible from lg. -->
		<AccountOverview {showPeriod} {loading} {change} />

		<!-- Narrow screens: destinations behind the menu, the same overview behind one entry. -->
		<div class="flex items-center gap-1 lg:hidden">
			<div class="md:hidden">
				<NavMenu items={navItems} currentPath={activeHref} triggerLabel="Menu" />
			</div>
			<button
				type="button"
				aria-expanded={overviewOpen}
				aria-controls="overview-panel"
				aria-label="Your overview"
				class="inline-flex h-11 items-center gap-1 rounded-lg px-2 text-slate-700 transition-colors hover:bg-slate-100 focus-visible:ring-2 focus-visible:ring-amber-300 focus-visible:outline-none"
			>
				<Avatar name={account.name} size="sm" />
				<Icon icon={icons.chevronDown} size="sm" />
			</button>
		</div>

		{#if overviewOpen}
			<div
				id="overview-panel"
				class={['absolute top-full right-0 mt-2 lg:hidden', zClasses.popover].join(' ')}
			>
				<AccountOverview layout="panel" {showPeriod} {loading} {change} />
			</div>
		{/if}
	</div>
</header>
