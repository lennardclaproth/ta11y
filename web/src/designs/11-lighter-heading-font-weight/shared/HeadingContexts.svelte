<script lang="ts">
	// Design prototype for issue #11. Shows every place an EB Garamond heading appears —
	// page title, chart-card titles, modal title, drawer title, login title — on one surface,
	// next to the body text and buttons the weight has to be judged against.
	//
	// The modal and drawer are reproduced as their header markup (same atoms, same spacing as
	// Dialog.svelte / Drawer.svelte) instead of as real overlays, so all five contexts can be
	// compared in a single screenshot. Nothing outside this folder is touched.
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';

	import {
		drawerListings,
		ledgerRows,
		netCashflowSeries,
		netCashflowTotal,
		spendingSeries,
		spendingTotal
	} from './heading-weight.fixtures';

	type Props = {
		/** Suffix keeping field ids unique when two variants render on one page. */
		idPrefix?: string;
	};

	let { idPrefix = 'd11' }: Props = $props();
</script>

<div class="flex flex-col gap-8 pb-8">
	<!-- 1. Masthead + page title (TopNavbar). The wordmark is not a Heading and stays untouched. -->
	<header class="px-4 pt-3 lg:px-8">
		<div class="flex items-center justify-between gap-4 border-t-2 border-b border-slate-800 py-2">
			<span
				class="font-heading text-5xl leading-none tracking-tight text-slate-900"
				aria-hidden="true">ta11y</span
			>
			<nav aria-label="Primary navigation" class="hidden items-center gap-6 md:flex">
				<span class="border-b-2 border-amber-500 py-3 text-sm font-semibold text-slate-950"
					>Cashflow</span
				>
				<span class="border-b-2 border-transparent py-3 text-sm text-slate-700">Assets</span>
				<span class="border-b-2 border-transparent py-3 text-sm text-slate-700">Portfolio</span>
			</nav>
			<span class="text-sm text-slate-700 md:hidden">Menu</span>
		</div>

		<div class="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 pt-4">
			<Heading level="h1" size="2xl" class="leading-none">Cashflow</Heading>

			<!-- Matching TopNavbar: search, period and account live here; the create action sits in
			     the ruled ledger header below, as DESIGN.md prescribes. -->
			<div class="flex max-w-full flex-wrap items-center gap-2">
				<div class="w-full md:w-72">
					<SearchInput placeholder="Search transactions…" size="md" shape="default" />
				</div>
				<Button intent="secondary" variant="outline" size="md">
					<Icon icon="heroicons:calendar-days" size="sm" />
					1 – 31 Mar
				</Button>
				<Button intent="secondary" variant="ghost" size="md">
					<Icon icon="heroicons:user-circle" size="sm" />
					Account
				</Button>
			</div>
		</div>
	</header>

	<!-- 2. Chart-card titles (AnalyticsCard). -->
	<section class="grid gap-4 px-4 sm:grid-cols-2 lg:px-8">
		<AnalyticsCard title="Net cashflow">
			<div class="flex items-end justify-between gap-4">
				<div>
					<Money amount={netCashflowTotal} size="xl" weight="semibold" colored signDisplay="always" />
					<Text size="sm" tone="muted" class="mt-1 block">Last 30 days · EUR</Text>
				</div>
				<Sparkline
					data={netCashflowSeries}
					width={140}
					height={44}
					fill
					ariaLabel="Net cashflow trend over the last 30 days"
				/>
			</div>
		</AnalyticsCard>

		<AnalyticsCard title="Spending by category">
			<div class="flex items-end justify-between gap-4">
				<div>
					<Money amount={spendingTotal} size="xl" weight="semibold" />
					<Text size="sm" tone="muted" class="mt-1 block">Last 30 days · EUR</Text>
				</div>
				<Sparkline
					data={spendingSeries}
					width={140}
					height={44}
					fill
					ariaLabel="Spending trend over the last 30 days"
				/>
			</div>
		</AnalyticsCard>
	</section>

	<!-- 3. Ledger panel title (LedgerToolbar). This one is a bare <h2>, not the Heading atom, so
	     today it sits at the browser's default bold — the heaviest Garamond heading in the app. -->
	<section class="px-4 lg:px-8">
		<Text size="xs" tone="muted" class="mb-2 block tracking-wider uppercase">Ledger panel</Text>
		<Panel variant="muted" padding="none" shape="square" bordered={false}>
			<div
				class="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
			>
				<h2 class="text-2xl">Transactions</h2>
				<Button shape="default">
					<Icon icon="heroicons:plus" size="sm" />
					Add transaction
				</Button>
			</div>
			<ul class="divide-y divide-slate-200">
				{#each ledgerRows as row (row.description)}
					<li class="flex items-center justify-between gap-4 px-4 py-3">
						<div class="min-w-0">
							<Text size="sm" tone="strong" class="block">{row.description}</Text>
							<Text size="xs" tone="muted" class="block">{row.date} · {row.category}</Text>
						</div>
						<Money amount={row.amount} colored signDisplay="always" />
					</li>
				{/each}
			</ul>
		</Panel>
	</section>

	<!-- 4 + 5. Modal title and drawer title, reproduced from Dialog/Drawer header markup. -->
	<section class="grid gap-6 px-4 lg:grid-cols-2 lg:px-8">
		<div>
			<Text size="xs" tone="muted" class="mb-2 block tracking-wider uppercase">Modal</Text>
			<div class="rounded-xl bg-white shadow-md">
				<header class="flex items-start justify-between gap-4 px-5 pt-5">
					<div class="min-w-0 flex-1">
						<Heading level="h2" size="lg" class="text-slate-900">Add transaction</Heading>
					</div>
					<IconButton icon="heroicons:x-mark" ariaLabel="Close" size="sm" />
				</header>
				<div class="flex flex-col gap-4 px-5 py-4">
					<FormField label="Description" id="{idPrefix}-description">
						{#snippet children(field)}
							<Input id={field.id} value="Monthly rent" />
						{/snippet}
					</FormField>
					<FormField label="Amount" id="{idPrefix}-amount" hint="Negative for money going out">
						{#snippet children(field)}
							<Input id={field.id} value="-1200.00" />
						{/snippet}
					</FormField>
				</div>
				<footer class="flex items-center justify-end gap-2 px-5 pb-5">
					<Button intent="secondary" variant="ghost">Cancel</Button>
					<Button intent="primary">Save changes</Button>
				</footer>
			</div>
		</div>

		<div>
			<Text size="xs" tone="muted" class="mb-2 block tracking-wider uppercase">Drawer</Text>
			<div class="border border-slate-200 bg-white shadow-md">
				<header
					class="flex items-start justify-between gap-4 border-b border-slate-200 px-5 py-4"
				>
					<div class="min-w-0 flex-1">
						<Heading level="h2" size="lg" class="text-slate-900">Browse listings</Heading>
					</div>
					<IconButton icon="heroicons:x-mark" ariaLabel="Close" size="sm" />
				</header>
				<ul class="divide-y divide-slate-100 px-5 py-2">
					{#each drawerListings as listing (listing.symbol)}
						<li class="flex items-center justify-between gap-3 py-3">
							<div class="min-w-0">
								<Text size="sm" tone="strong" class="block">{listing.name}</Text>
								<Text size="xs" tone="muted" class="block">{listing.exchange}</Text>
							</div>
							<Badge intent="neutral" size="sm">{listing.symbol}</Badge>
						</li>
					{/each}
				</ul>
			</div>
		</div>
	</section>

	<!-- 5. Login title. -->
	<section class="px-4 lg:px-8">
		<Text size="xs" tone="muted" class="mb-2 block tracking-wider uppercase">Sign-in page</Text>
		<Panel class="w-full max-w-md" padding="lg">
			<Heading level="h2" size="xl">Sign in</Heading>
			<Text class="mt-2 block text-sm leading-relaxed text-slate-600">
				ta11y uses your existing account with an identity provider. It never sees or stores a
				password.
			</Text>
			<div class="mt-8">
				<Button intent="secondary" variant="outline" class="w-full">Continue with Google</Button>
			</div>
		</Panel>
	</section>
</div>
