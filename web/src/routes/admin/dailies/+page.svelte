<script lang="ts">
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import ListingSearchSelect from '$lib/components/molecules/listing-search-select/ListingSearchSelect.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import EodUploadForm from '$lib/components/organisms/eod-upload-form/EodUploadForm.svelte';
	import { getEOD } from '$lib/services/marketdata';
	import { importEOD } from '$lib/services/importer';
	import { toast } from '$lib/stores/toast.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { EOD, ListingSearchRow } from '$lib/api/types';

	let selected = $state<ListingSearchRow | null>(null);
	let rows = $state<EOD[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);

	/**
	 * Identifies the load the table is currently waiting on. A listing's history is one
	 * unbounded request, so picking a second listing while the first is still in flight can
	 * see them resolve out of order and leave one listing's prices sitting under the other's
	 * name. Every write below is gated on still being the newest request; clearing bumps it
	 * too, so a response that arrives after the table is emptied has nothing to land in.
	 */
	let requestId = 0;

	let uploadOpen = $state(false);
	let uploading = $state(false);

	const currency = $derived(selected?.currency ?? 'EUR');

	const price = (value: number) =>
		scaledToNumber(value).toLocaleString('en', { style: 'currency', currency });

	async function selectListing(listing: ListingSearchRow) {
		const id = ++requestId;
		selected = listing;
		loading = true;
		error = null;
		try {
			const response = await getEOD({ listing_id: listing.id, sort_order: 'desc' });
			if (id !== requestId) return;
			rows = response.Data;
		} catch {
			if (id !== requestId) return;
			error = 'Failed to load daily data';
		} finally {
			// A superseded load must not clear the spinner the newer one turned on.
			if (id === requestId) loading = false;
		}
	}

	/**
	 * Uploads a price file and then watches for the rows to land.
	 *
	 * The import is processed asynchronously and the API has no status endpoint for it --
	 * EOD imports are not account-scoped, so they raise no websocket event either. The only
	 * honest signal available is the data itself, so the table is re-read a few times and
	 * the row count is compared. A run that has not landed by the deadline is reported as
	 * still processing rather than as a failure, because it very likely is.
	 */
	async function uploadPrices(file: File) {
		const target = selected;
		if (!target) return;
		const before = rows.length;
		await importEOD({ file, listing_id: target.id });
		uploadOpen = false;
		toast.success(`Upload accepted for ${target.symbol}`);
		void awaitNewRows(target, before);
	}

	/** Re-reads the listing's prices until the count grows or the deadline passes. */
	async function awaitNewRows(listing: ListingSearchRow, before: number) {
		const deadline = Date.now() + 15_000;
		while (Date.now() < deadline) {
			await new Promise((resolve) => setTimeout(resolve, 1500));
			// Abandon quietly if the user moved on: they are looking at something else now.
			if (selected?.id !== listing.id) return;
			await selectListing(listing);
			if (rows.length > before) {
				const added = rows.length - before;
				toast.success(`${added} price row${added === 1 ? '' : 's'} added to ${listing.symbol}`);
				return;
			}
		}
		toast.info(`${listing.symbol} is still processing. Search it again shortly to see the rows.`);
	}

	// The search box owns only its own text; the table is this page's state. Without this the
	// cleared box would sit above a full table of prices with nothing left naming the listing
	// they belong to. `loading` is reset here rather than by the abandoned request, whose
	// `finally` now declines to touch it.
	function clearListing() {
		requestId++;
		selected = null;
		rows = [];
		error = null;
		loading = false;
	}
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar title="Dailies" />
	{/snippet}

	<PageContentTemplate>
		<div class="flex min-h-0 flex-1 flex-col">
			<div class="border-b border-slate-200 p-3">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div class="w-full max-w-sm">
						<ListingSearchSelect
							value={selected}
							placeholder="Search a listing…"
							onSelect={selectListing}
							onClear={clearListing}
						/>
					</div>
					<!-- Only listings on a manual provider accept uploads; the API is the authority on
					     which those are, so the action stays available and reports its refusal. -->
					<Button variant="outline" disabled={!selected} onclick={() => (uploadOpen = true)}>
						Upload prices
					</Button>
				</div>
			</div>

			{#if !selected}
				<div class="flex flex-1 items-center justify-center p-8 text-sm text-slate-500">
					Search for a listing to view its daily OHLCV data.
				</div>
			{:else}
				<DataTable
					{rows}
					{loading}
					{error}
					getRowId={(r: EOD) => r.ID}
					emptyText="No daily data for this listing"
					columns={[
						{
							key: 'date',
							header: 'Date',
							value: (r: EOD) => formatDisplayDate(r.Date.slice(0, 10))
						},
						{ key: 'open', header: 'Open', align: 'right', value: (r: EOD) => price(r.Open) },
						{ key: 'high', header: 'High', align: 'right', value: (r: EOD) => price(r.High) },
						{ key: 'low', header: 'Low', align: 'right', value: (r: EOD) => price(r.Low) },
						{ key: 'close', header: 'Close', align: 'right', value: (r: EOD) => price(r.Close) },
						{
							key: 'volume',
							header: 'Volume',
							align: 'right',
							value: (r: EOD) => r.Volume.toLocaleString('en')
						}
					]}
				/>
			{/if}
		</div>
	</PageContentTemplate>
</AppShellTemplate>

<Dialog
	bind:open={uploadOpen}
	title="Upload prices"
	size="lg"
	dismissible={!uploading}
	closeOnBackdrop={!uploading}
	closeOnEscape={!uploading}
>
	{#if uploadOpen && selected}
		<EodUploadForm
			listing={selected}
			bind:uploading
			onUpload={uploadPrices}
			onCancel={() => (uploadOpen = false)}
		/>
	{/if}
</Dialog>
