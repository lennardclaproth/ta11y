<script lang="ts">
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import NavMenu from '$lib/components/molecules/nav-menu/NavMenu.svelte';
	import { zClasses } from '$lib/styles/z-index';
	import OverviewCardC from './OverviewCardC.svelte';
	import { account, navItems } from '../shared/mock-data';

	type Props = {
		activeHref: string;
		overviewOpen?: boolean;
		showPeriod?: boolean;
	};

	let { activeHref, overviewOpen = false, showPeriod = true }: Props = $props();

	const isActive = (href: string) => activeHref === href || activeHref.startsWith(`${href}/`);
</script>

<header class="px-4 pt-3 lg:px-8">
	<!-- Navigation only: one thin rule, the wordmark, the destinations, and the account. -->
	<div class="relative flex items-center gap-6 border-b border-slate-800 pb-1">
		<a
			href="/cashflow"
			aria-label="ta11y home"
			class="font-heading text-3xl leading-none tracking-tight text-slate-900 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700"
			>ta11y</a
		>

		<nav
			aria-label="Primary navigation"
			class="hidden flex-1 items-center gap-6 border-l border-slate-300 pl-6 md:flex"
		>
			{#each navItems as item (item.href)}
				<a
					href={item.href}
					aria-current={isActive(item.href) ? 'page' : undefined}
					class="border-b-2 border-transparent py-2 text-sm text-slate-700 transition-colors hover:border-slate-400 hover:text-slate-950 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-slate-700 aria-[current=page]:border-amber-500 aria-[current=page]:font-semibold aria-[current=page]:text-slate-950"
					>{item.label}</a
				>
			{/each}
		</nav>

		<div class="ml-auto flex items-center gap-1 md:ml-0">
			<div class="md:hidden">
				<NavMenu items={navItems} currentPath={activeHref} triggerLabel="Menu" />
			</div>
			<button
				type="button"
				aria-expanded={overviewOpen}
				aria-controls="overview-panel-c"
				aria-label="Your overview"
				class="inline-flex size-9 items-center justify-center rounded-full transition-colors hover:bg-slate-100 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
			>
				<Avatar name={account.name} size="sm" />
			</button>
		</div>

		{#if overviewOpen}
			<div
				id="overview-panel-c"
				class={['absolute top-full right-0 mt-2', zClasses.popover].join(' ')}
			>
				<OverviewCardC {showPeriod} />
			</div>
		{/if}
	</div>
</header>
