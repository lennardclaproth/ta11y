<script lang="ts">
	// Design prototype for issue #11. Shows every place an EB Garamond heading appears — page
	// title, chart-card titles, ledger title, admin title, modal title, drawer title, drawer
	// labels and the login title — on one surface, next to the body text and buttons the weight
	// has to be judged against.
	//
	// The modal and drawer are reproduced as their header markup (same atoms, same spacing as
	// Dialog.svelte / Drawer.svelte) instead of as real overlays, so every context fits in one
	// screenshot. Nothing outside this folder is touched.
	//
	// `state` drives only the *content* under the headings. The headings themselves render
	// identically in every state — that is the point: the weight is a theme decision, not a
	// per-state one.
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
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
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';

	import {
		drawerAssets,
		drawerListings,
		drawerMutations,
		ledgerRows,
		netCashflowSeries,
		netCashflowTotal,
		spendingSeries,
		spendingTotal
	} from './heading-weight.fixtures';

	type ContextState = 'default' | 'loading' | 'empty' | 'error' | 'no-matches';

	type Props = {
		/** Which content state sits under the headings. */
		state?: ContextState;
		/** Suffix keeping field ids unique when two surfaces render on one page. */
		idPrefix?: string;
	};

	let { state = 'default', idPrefix = 'd11' }: Props = $props();

	const showsRecords = $derived(state === 'default');
	const searchTerm = $derived(state === 'no-matches' ? 'zzz-unknown' : '');

	// DESIGN.md: keep zero, unknown and loading visibly distinct. In the no-matches state the
	// total really is zero, so it stays a zero — the caption says so rather than implying a period.
	const kpiCaption = $derived(
		state === 'no-matches' ? '0 matching transactions · EUR' : 'Last 30 days · EUR'
	);
</script>

<div class="flex flex-col gap-8 pb-8">
	<!-- 1. Masthead + page title (TopNavbar). The wordmark is not a heading and stays untouched. -->
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
					<SearchInput
						value={searchTerm}
						placeholder="Search transactions…"
						size="md"
						shape="default"
					/>
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

	<!-- The server-unreachable heading from routes/+layout.svelte: a bare <h1>, so today it is the
	     heaviest heading in the app. Only shown in the error state, where it actually appears. -->
	{#if state === 'error'}
		<section class="px-4 lg:px-8">
			<Text size="xs" tone="muted" class="mb-2 block tracking-wider uppercase">
				Session bootstrap failure
			</Text>
			<Panel padding="lg" class="max-w-xl">
				<h1 class="text-2xl">Can't reach the server</h1>
				<Text size="sm" tone="muted" class="mt-2 block">
					The server did not respond. Check your connection and try again.
				</Text>
				<div class="mt-4">
					<Button intent="primary">Try again</Button>
				</div>
			</Panel>
		</section>
	{/if}

	<!-- 2. Chart-card titles (AnalyticsCard). -->
	<section class="grid gap-4 px-4 sm:grid-cols-2 lg:px-8">
		<AnalyticsCard title="Net cashflow">
			{#if state === 'loading'}
				<div class="flex flex-col gap-2">
					<Skeleton variant="rect" width="8rem" height="1.75rem" />
					<Skeleton variant="rect" width="100%" height="2.75rem" />
				</div>
			{:else if state === 'error'}
				<Alert intent="error" title="Couldn't load net cashflow">Try again in a moment.</Alert>
			{:else if state === 'empty'}
				<Text size="sm" tone="muted" class="block">
					No transactions yet — add one to see your net cashflow.
				</Text>
			{:else}
				<div class="flex items-end justify-between gap-4">
					<div>
						<Money
							amount={state === 'no-matches' ? 0 : netCashflowTotal}
							size="xl"
							weight="semibold"
							colored={state !== 'no-matches'}
							signDisplay={state === 'no-matches' ? 'auto' : 'always'}
						/>
						<Text size="sm" tone="muted" class="mt-1 block">{kpiCaption}</Text>
					</div>
					{#if state !== 'no-matches'}
						<Sparkline
							data={netCashflowSeries}
							width={140}
							height={44}
							fill
							ariaLabel="Net cashflow trend over the last 30 days"
						/>
					{/if}
				</div>
			{/if}
		</AnalyticsCard>

		<AnalyticsCard title="Spending by category">
			{#if state === 'loading'}
				<div class="flex flex-col gap-2">
					<Skeleton variant="rect" width="8rem" height="1.75rem" />
					<Skeleton variant="rect" width="100%" height="2.75rem" />
				</div>
			{:else if state === 'error'}
				<Alert intent="error" title="Couldn't load spending">Try again in a moment.</Alert>
			{:else if state === 'empty'}
				<Text size="sm" tone="muted" class="block">Nothing to break down yet.</Text>
			{:else}
				<div class="flex items-end justify-between gap-4">
					<div>
						<Money amount={state === 'no-matches' ? 0 : spendingTotal} size="xl" weight="semibold" />
						<Text size="sm" tone="muted" class="mt-1 block">{kpiCaption}</Text>
					</div>
					{#if state !== 'no-matches'}
						<Sparkline
							data={spendingSeries}
							width={140}
							height={44}
							fill
							ariaLabel="Spending trend over the last 30 days"
						/>
					{/if}
				</div>
			{/if}
		</AnalyticsCard>
	</section>

	<!-- 3. Ledger panel title (LedgerToolbar): a bare <h2>, not the Heading atom, so today it sits
	     at the browser's default bold — heavier still than the page title. -->
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

			{#if state === 'loading'}
				<div class="flex flex-col gap-3 px-4 py-4">
					{#each [0, 1, 2] as row (row)}
						<Skeleton variant="rect" width="100%" height="2.5rem" />
					{/each}
				</div>
			{:else if state === 'error'}
				<div class="px-4 py-4">
					<Alert intent="error" title="Couldn't load transactions">
						The request failed. Your filters are still in place.
					</Alert>
					<div class="mt-3">
						<Button intent="secondary" variant="outline" size="sm">Try again</Button>
					</div>
				</div>
			{:else if state === 'empty'}
				<div class="px-4 py-10 text-center">
					<Text size="sm" tone="strong" class="block">No transactions yet</Text>
					<Text size="sm" tone="muted" class="mt-1 block">
						Add your first transaction to start your ledger.
					</Text>
					<div class="mt-4">
						<Button intent="primary" size="sm">
							<Icon icon="heroicons:plus" size="sm" />
							Add transaction
						</Button>
					</div>
				</div>
			{:else if state === 'no-matches'}
				<div class="px-4 py-10 text-center">
					<Text size="sm" tone="strong" class="block">No transactions match your search</Text>
					<Text size="sm" tone="muted" class="mt-1 block">
						Nothing found for “{searchTerm}” between 1 and 31 Mar.
					</Text>
					<div class="mt-4">
						<Button intent="secondary" variant="outline" size="sm">Clear filters</Button>
					</div>
				</div>
			{:else}
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
			{/if}
		</Panel>
	</section>

	<!-- 4 – 9. Modal, both drawers, the admin page and the login page. Modal and drawer headers are
	     reproduced from Dialog.svelte / Drawer.svelte markup rather than opened as real overlays,
	     so every remaining heading fits in one screenshot. -->
	<section class="grid items-start gap-6 px-4 lg:grid-cols-2 lg:px-8">
		<!-- Left column: modal, admin page, sign-in page. -->
		<div class="flex flex-col gap-6">
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

			<!-- Admin pages use another bare <h2>, at the same size as the ledger title. -->
			<div>
				<Text size="xs" tone="muted" class="mb-2 block tracking-wider uppercase">Admin page</Text>
				<Panel padding="lg">
					<h2 class="text-2xl">Market listings</h2>
					<Text size="sm" tone="muted" class="mt-2 block">
						Listings available to every account in this workspace.
					</Text>
				</Panel>
			</div>

			<div>
				<Text size="xs" tone="muted" class="mb-2 block tracking-wider uppercase">Sign-in page</Text>
				<Panel padding="lg">
					<Heading level="h2" size="xl">Sign in</Heading>
					<Text size="sm" tone="muted" class="mt-2 block">
						ta11y uses your existing account with an identity provider. It never sees or stores a
						password.
					</Text>
					<div class="mt-8">
						<Button intent="secondary" variant="outline" class="w-full">Continue with Google</Button>
					</div>
				</Panel>
			</div>
		</div>

		<!-- Right column: both drawers Lennard named. Same title markup, different bodies. -->
		<div class="flex flex-col gap-6">
			<div>
				<Text size="xs" tone="muted" class="mb-2 block tracking-wider uppercase">
					Drawer — asset class
				</Text>
				<div class="border border-slate-200 bg-white shadow-md">
					<header
						class="flex items-start justify-between gap-4 border-b border-slate-200 px-5 py-4"
					>
						<div class="min-w-0 flex-1">
							<Heading level="h2" size="lg" class="text-slate-900">Equities</Heading>
						</div>
						<IconButton icon="heroicons:x-mark" ariaLabel="Close" size="sm" />
					</header>

					<div class="flex flex-col gap-6 px-5 py-4">
						<!-- The two small labels Lennard asked to bring along: <h3> at 14px, semibold today. -->
						<section>
							<h3 class="mb-2 text-sm text-slate-900">Assets</h3>
							{#if state === 'loading'}
								<div class="flex flex-col gap-2">
									{#each [0, 1] as row (row)}
										<Skeleton variant="rect" width="100%" height="2rem" />
									{/each}
								</div>
							{:else if showsRecords || state === 'no-matches'}
								<ul class="divide-y divide-slate-100 rounded-xl border border-slate-200">
									{#each drawerAssets as asset (asset.name)}
										<li class="flex items-center justify-between gap-3 px-3 py-2">
											<Text size="sm" class="min-w-0 truncate">{asset.name}</Text>
											<Money amount={asset.worth} size="sm" />
										</li>
									{/each}
								</ul>
							{:else}
								<Text size="sm" tone="muted" class="block">No assets in this class</Text>
							{/if}
						</section>

						<section>
							<h3 class="mb-2 text-sm text-slate-900">Recent changes</h3>
							{#if state === 'loading'}
								<Skeleton variant="rect" width="100%" height="2rem" />
							{:else if showsRecords || state === 'no-matches'}
								<ul class="flex flex-col gap-2">
									{#each drawerMutations as mutation (mutation.date)}
										<li class="flex items-center justify-between gap-3">
											<Text size="sm" class="min-w-0">
												{mutation.changeType} · <span class="text-xs text-slate-500"
													>{mutation.date}</span
												>
											</Text>
											<Money amount={mutation.worth} size="sm" />
										</li>
									{/each}
								</ul>
							{:else}
								<Text size="sm" tone="muted" class="block">No recorded changes</Text>
							{/if}
						</section>
					</div>
				</div>
			</div>

			<div>
				<Text size="xs" tone="muted" class="mb-2 block tracking-wider uppercase">
					Drawer — browse listings
				</Text>
				<div class="border border-slate-200 bg-white shadow-md">
					<header
						class="flex items-start justify-between gap-4 border-b border-slate-200 px-5 py-4"
					>
						<div class="min-w-0 flex-1">
							<Heading level="h2" size="lg" class="text-slate-900">Browse listings</Heading>
						</div>
						<IconButton icon="heroicons:x-mark" ariaLabel="Close" size="sm" />
					</header>

					<div class="px-5 py-4">
						{#if state === 'loading'}
							<div class="flex flex-col gap-2">
								{#each [0, 1, 2] as row (row)}
									<Skeleton variant="rect" width="100%" height="2.5rem" />
								{/each}
							</div>
						{:else if state === 'no-matches'}
							<Text size="sm" tone="strong" class="block">No listings match “{searchTerm}”</Text>
							<Text size="sm" tone="muted" class="mt-1 block">
								Try a different symbol or name.
							</Text>
						{:else if state === 'empty'}
							<Text size="sm" tone="muted" class="block">No listings available yet.</Text>
						{:else if state === 'error'}
							<Alert intent="error" title="Couldn't load listings">Try again in a moment.</Alert>
						{:else}
							<ul class="divide-y divide-slate-100">
								{#each drawerListings as listing (listing.symbol)}
									<li class="flex items-center justify-between gap-3 py-2">
										<div class="min-w-0">
											<Text size="sm" tone="strong" class="block">{listing.name}</Text>
											<Text size="xs" tone="muted" class="block">{listing.exchange}</Text>
										</div>
										<Badge intent="neutral" size="sm">{listing.symbol}</Badge>
									</li>
								{/each}
							</ul>
						{/if}
					</div>
				</div>
			</div>
		</div>
	</section>
</div>
