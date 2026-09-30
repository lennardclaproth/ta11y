<script lang="ts">
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import NavMenu from '$lib/components/molecules/nav-menu/NavMenu.svelte';
	import { zClasses } from '$lib/styles/z-index';
	import OverviewCardB from './OverviewCardB.svelte';
	import { icons } from '../shared/icons';
	import { account, navItems } from '../shared/mock-data';

	type Props = {
		activeHref: string;
		/** One entry point on every width; open it to show the card in the prototype. */
		overviewOpen?: boolean;
		/** Admin pages have no period, so the card drops that block. */
		showPeriod?: boolean;
	};

	let { activeHref, overviewOpen = false, showPeriod = true }: Props = $props();

	const isActive = (href: string) => activeHref === href || activeHref.startsWith(`${href}/`);
</script>

<header class="px-4 pt-3 pb-4 lg:px-8">
	<div class="relative flex items-center gap-6 border-t-2 border-b border-slate-800 py-2">
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

		<div class="ml-auto flex items-center gap-1 md:ml-0">
			<div class="md:hidden">
				<NavMenu items={navItems} currentPath={activeHref} triggerLabel="Menu" />
			</div>
			<button
				type="button"
				aria-expanded={overviewOpen}
				aria-controls="overview-panel-b"
				aria-label="Your overview"
				class="inline-flex h-10 items-center gap-2 rounded-lg px-2 text-slate-700 transition-colors hover:bg-slate-100 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
			>
				<Avatar name={account.name} size="sm" />
				<Icon icon={icons.chevronDown} size="sm" />
			</button>
		</div>

		{#if overviewOpen}
			<div
				id="overview-panel-b"
				class={['absolute top-full right-0 mt-2', zClasses.popover].join(' ')}
			>
				<OverviewCardB {showPeriod} />
			</div>
		{/if}
	</div>
</header>
