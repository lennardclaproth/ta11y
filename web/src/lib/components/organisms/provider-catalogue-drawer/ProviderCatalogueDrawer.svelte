<script lang="ts">
	import { ApiError } from '$lib/api/client';
	import type { CatalogueStatus, CatalogueSync, Listing, ListingSearchRow } from '$lib/api/types';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import FooterBar from '$lib/components/organisms/footer-bar/FooterBar.svelte';
	import {
		createListing,
		getCatalogueStatus,
		searchListings,
		searchProviderCatalogue,
		startCatalogueSync
	} from '$lib/services/marketdata';
	import { toast } from '$lib/stores/toast.svelte';
	import type { AdoptOutcome } from './provider-catalogue-drawer.types';

	type Props = {
		/** Two-way bindable open state. */
		open?: boolean;
		/** Provider whose catalogue is searched. */
		source?: string;
		/** Called with the listings that were created, so the caller can refresh. */
		onAdopted?: (listings: Listing[]) => void;
	};

	let { open = $bindable(false), source = 'market_stack', onAdopted }: Props = $props();

	let query = $state('');
	let rows = $state<ListingSearchRow[]>([]);
	let total = $state(0);
	let limit = $state(25);
	let offset = $state(0);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let searched = $state(false);

	let selectedIds = $state<string[]>([]);
	let adopting = $state(false);
	let outcomes = $state<AdoptOutcome[]>([]);

	let askingProvider = $state(false);
	/** Provider match count for the last metered search, when it exceeded one page. */
	let upstreamTotal = $state<number | null>(null);
	let truncated = $state(false);

	let status = $state<CatalogueStatus | null>(null);

	/**
	 * Page budget asked for on a resync. Sent explicitly rather than left to the backend
	 * default so the button can name the cost it is about to spend -- one provider request
	 * per page -- the same way the metered search does.
	 */
	const SEED_PAGES = 20;
	/** How often a run in flight is re-read. The run is detached, so progress only arrives by asking. */
	const SYNC_POLL_MS = 2000;

	let startingSync = $state(false);
	let syncError = $state<string | null>(null);
	/** The run this drawer is following, so its outcome is announced exactly once. */
	let watchedRunId = $state<string | null>(null);

	const selectedRows = $derived(rows.filter((row) => selectedIds.includes(row.id)));

	/** Days since the catalogue was last seeded, used for the staleness warning. */
	const daysSinceSync = $derived.by(() => {
		const finished = status?.latest_sync?.finished_at;
		if (!finished) return null;
		const elapsed = Date.now() - new Date(finished).getTime();
		return Math.floor(elapsed / 86_400_000);
	});

	const latestRun = $derived(status?.latest_sync ?? null);
	const syncRunning = $derived(latestRun?.status === 'running');

	// Reset per-session state each time the drawer opens, and refresh the cache
	// summary so the staleness line reflects reality rather than a stale snapshot.
	$effect(() => {
		if (!open) return;
		outcomes = [];
		selectedIds = [];
		clearProviderNotice();
		syncError = null;
		void loadStatus();
	});

	// A seed run is detached from the request that started it, so progress only exists on the
	// run row: poll while one is in flight. `syncRunning` is a boolean, so this re-runs on the
	// transition rather than on every refreshed status, and tears the interval down when the
	// run reaches a terminal state or the drawer closes.
	$effect(() => {
		if (!open || !syncRunning) return;
		const timer = setInterval(() => void loadStatus(), SYNC_POLL_MS);
		return () => clearInterval(timer);
	});

	/**
	 * Drops the "the provider matched more than it returned" notice. It describes one provider
	 * response and one query, so anything that replaces the rows -- a local search, a page, a
	 * reopen -- makes it a claim about results no longer on screen.
	 */
	function clearProviderNotice() {
		upstreamTotal = null;
		truncated = false;
	}

	async function loadStatus() {
		try {
			const next = await getCatalogueStatus(source);
			status = next;
			announceIfFinished(next.latest_sync ?? null);
		} catch {
			// The summary is contextual; failing to load it must not block searching. The last
			// known value is kept rather than cleared: a poll that blips would otherwise hide the
			// line and, because the effect watches `syncRunning`, stop the polling with it.
		}
	}

	/** Reports the followed run once it reaches a terminal state, then stops following it. */
	function announceIfFinished(run: CatalogueSync | null) {
		if (!run || run.id !== watchedRunId || run.status === 'running') return;
		watchedRunId = null;
		if (run.status === 'completed') {
			toast.success(`Catalogue resynced — ${run.rows_upserted.toLocaleString()} entries cached`);
		} else {
			syncError = run.last_error ?? 'The catalogue sync failed. Try again.';
		}
	}

	/**
	 * Starts a bounded seed run. The API answers immediately and works in the background, so
	 * the only thing to do here is start following it.
	 */
	async function startSync() {
		startingSync = true;
		syncError = null;
		try {
			const run = await startCatalogueSync({ source, pages: SEED_PAGES });
			watchedRunId = run.id;
			// Show the run straight away rather than waiting a poll for it to appear.
			if (status) status = { ...status, latest_sync: run };
		} catch (cause) {
			if (cause instanceof ApiError && cause.status === 409) {
				// One is already running -- follow that one instead of reporting a failure.
				await loadStatus();
				watchedRunId = status?.latest_sync?.id ?? null;
			} else {
				syncError = 'Could not start the catalogue sync. Check your connection and try again.';
			}
		} finally {
			startingSync = false;
		}
	}

	/**
	 * Searches what is cached locally. This is free and instant, which is why it is
	 * the default: the provider is only consulted when the user asks.
	 */
	async function searchLocal(nextOffset = 0) {
		if (!query.trim()) return;
		loading = true;
		error = null;
		clearProviderNotice();
		try {
			const response = await searchListings({
				q: query.trim(),
				limit,
				offset: nextOffset,
				scope: 'all'
			});
			rows = response.data;
			total = response.pagination.total;
			offset = nextOffset;
			searched = true;
		} catch {
			error = 'Could not search the catalogue. Check your connection and try again.';
		} finally {
			loading = false;
		}
	}

	/**
	 * Asks the provider directly, which costs one request. The local cache can never
	 * be known to be complete for a query it has not seen, so this is a deliberate
	 * action rather than an automatic fallback.
	 */
	async function askProvider() {
		if (!query.trim()) return;
		askingProvider = true;
		error = null;
		try {
			const response = await searchProviderCatalogue({ source, q: query.trim(), limit });
			rows = response.data;
			total = response.pagination.total;
			offset = 0;
			searched = true;
			upstreamTotal = response.upstream_total;
			truncated = response.truncated;
			await loadStatus();
		} catch (cause) {
			error =
				cause instanceof ApiError && cause.status === 400
					? 'The provider rejected that search. Try a different term.'
					: 'Could not reach the market data provider. Check your connection and try again.';
		} finally {
			askingProvider = false;
		}
	}

	/**
	 * Adopts the selected catalogue rows as listings, one request each, reporting a
	 * per-row outcome so one rejection does not discard the rest of the batch.
	 *
	 * Price history is deliberately not fetched here: the backfill is synchronous and
	 * one per listing would stall the request and burn through the provider's request
	 * budget. It happens the first time the listing's prices are viewed.
	 */
	async function adoptSelection() {
		const targets = selectedRows.filter((row) => row.adoptable);
		if (targets.length === 0) return;

		adopting = true;
		error = null;
		const results: AdoptOutcome[] = [];
		const created: Listing[] = [];

		for (const row of targets) {
			try {
				const listing = await createListing({
					symbol: row.symbol,
					name: row.name ?? '',
					source: row.source,
					exchange: row.exchange ?? undefined,
					sync_prices: false
				});
				created.push(listing);
				results.push({ symbol: row.symbol, status: 'added' });
			} catch (cause) {
				if (cause instanceof ApiError && cause.status === 409) {
					results.push({ symbol: row.symbol, status: 'exists' });
				} else {
					results.push({
						symbol: row.symbol,
						status: 'failed',
						message: cause instanceof ApiError ? cause.message : 'unexpected error'
					});
				}
			}
		}

		outcomes = results;
		selectedIds = [];
		adopting = false;
		if (created.length > 0) onAdopted?.(created);
		// Re-run the search so adopted rows come back marked as tracked.
		await searchLocal(offset);
	}

	function statusBadge(row: ListingSearchRow) {
		if (row.tracked) return { intent: 'success' as const, label: 'In your listings' };
		if (!row.name || !row.name.trim()) return { intent: 'warning' as const, label: 'No name' };
		if (!row.has_eod) return { intent: 'warning' as const, label: 'No price history' };
		return { intent: 'neutral' as const, label: 'Available' };
	}

	const addedCount = $derived(outcomes.filter((outcome) => outcome.status === 'added').length);
	const failedOutcomes = $derived(outcomes.filter((outcome) => outcome.status === 'failed'));
	const existingCount = $derived(outcomes.filter((outcome) => outcome.status === 'exists').length);
</script>

<Drawer bind:open title="Browse provider catalogue" width="max-w-3xl" dismissible={!adopting}>
	<div class="flex min-h-0 flex-col gap-4">
		<div class="space-y-2">
			<p class="text-sm text-slate-600">
				Search instruments cached from {source.replace('_', ' ')} and add them to your listings. Cached
				results are free; asking the provider directly costs one API request.
			</p>
			<div class="flex flex-wrap items-center justify-between gap-2">
				{#if status}
					<p class="text-xs text-slate-500">
						<span>{status.entries.toLocaleString()} entries cached</span>
						{#if syncRunning && latestRun}
							<!-- A run in flight replaces the staleness line: it has no finished_at yet, so
							     "last synced" would have nothing to say. -->
							<span>
								· syncing… {latestRun.pages_fetched} of {SEED_PAGES} pages, {latestRun.rows_upserted.toLocaleString()}
								entries
							</span>
						{:else}
							{#if daysSinceSync !== null}
								<span
									>· last synced {daysSinceSync === 0 ? 'today' : `${daysSinceSync} days ago`}</span
								>
							{/if}
							{#if daysSinceSync !== null && daysSinceSync > 90}
								<span class="text-amber-700">— this may be out of date.</span>
							{/if}
						{/if}
					</p>
				{/if}
				<Button
					variant="outline"
					intent="secondary"
					size="sm"
					disabled={startingSync || syncRunning}
					loading={startingSync || syncRunning}
					onclick={startSync}
				>
					{syncRunning ? 'Syncing…' : `Resync catalogue (${SEED_PAGES} requests)`}
				</Button>
			</div>
			{#if syncError}
				<p role="alert" class="text-sm text-red-700">{syncError}</p>
			{/if}
		</div>

		<div class="flex flex-wrap items-center gap-2">
			<div class="min-w-56 flex-1">
				<SearchInput
					bind:value={query}
					placeholder="Symbol, name or ISIN…"
					ariaLabel="Search the provider catalogue"
					onSearch={() => searchLocal(0)}
				/>
			</div>
			<Button disabled={!query.trim() || loading} onclick={() => searchLocal(0)}>Search</Button>
			<Button
				variant="outline"
				intent="secondary"
				disabled={!query.trim() || askingProvider}
				loading={askingProvider}
				onclick={askProvider}
			>
				{askingProvider ? 'Asking provider…' : 'Search provider (1 request)'}
			</Button>
		</div>

		{#if truncated && upstreamTotal !== null}
			<p role="status" class="text-sm text-amber-800">
				The provider matched {upstreamTotal.toLocaleString()} instruments and returned the top 100. Narrow
				your search if you don't see what you need.
			</p>
		{/if}

		{#if outcomes.length > 0}
			<div
				role="status"
				class="space-y-1 rounded-lg border border-slate-200 bg-taupe-50 p-3 text-sm"
			>
				{#if addedCount > 0}
					<p class="text-slate-800">Added {addedCount} listing{addedCount === 1 ? '' : 's'}.</p>
				{/if}
				{#if existingCount > 0}
					<p class="text-slate-600">
						{existingCount} already existed and {existingCount === 1 ? 'was' : 'were'} left unchanged.
					</p>
				{/if}
				{#each failedOutcomes as outcome (outcome.symbol)}
					<p class="text-red-700">{outcome.symbol} could not be added: {outcome.message}</p>
				{/each}
				<p class="text-xs text-slate-500">
					Price history is fetched the first time you open each listing.
				</p>
			</div>
		{/if}

		<DataTable
			{rows}
			{loading}
			{error}
			selectable
			bind:selectedIds
			getRowId={(row: ListingSearchRow) => row.id}
			isRowSelectable={(row: ListingSearchRow) => row.adoptable}
			emptyText={searched
				? 'Nothing cached for that search. Try asking the provider directly.'
				: 'Search for a symbol or company name to begin.'}
			columns={[
				{
					key: 'symbol',
					header: 'Symbol',
					width: 'w-40',
					value: (row: ListingSearchRow) => row.symbol
				},
				{ key: 'name', header: 'Name', cell: nameCell },
				{ key: 'exchange', header: 'Exchange', cell: exchangeCell },
				{ key: 'status', header: 'Status', width: 'w-40', cell: statusCell }
			]}
		>
			{#snippet footer()}
				<FooterBar
					{total}
					bind:limit
					bind:offset
					selectedCount={selectedRows.length}
					onPageChange={(next) => searchLocal(next)}
					onLimitChange={() => searchLocal(0)}
				>
					{#snippet actions()}
						<Button
							intent="success"
							disabled={adopting || selectedRows.length === 0}
							loading={adopting}
							onclick={adoptSelection}
						>
							{adopting
								? 'Adding…'
								: `Add ${selectedRows.length} listing${selectedRows.length === 1 ? '' : 's'}`}
						</Button>
					{/snippet}
				</FooterBar>
			{/snippet}
		</DataTable>
	</div>
</Drawer>

{#snippet nameCell(row: ListingSearchRow)}
	{#if row.name && row.name.trim()}
		<span class="text-slate-800">{row.name}</span>
	{:else}
		<span class="text-slate-400 italic">no name supplied</span>
	{/if}
{/snippet}

{#snippet exchangeCell(row: ListingSearchRow)}
	<span class="text-slate-700">{row.exchange ?? '—'}</span>
	{#if row.exchange_mic}
		<span class="ml-2 text-xs text-slate-500">{row.exchange_mic}</span>
	{/if}
{/snippet}

{#snippet statusCell(row: ListingSearchRow)}
	{@const badge = statusBadge(row)}
	<span title={row.adoptable_reason ?? undefined}>
		<Badge intent={badge.intent} variant="soft" size="sm">{badge.label}</Badge>
	</span>
{/snippet}
