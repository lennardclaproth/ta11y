<script lang="ts">
	import { onMount } from 'svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import ListingForm from '$lib/components/organisms/listing-form/ListingForm.svelte';
	import ListingEditForm from '$lib/components/organisms/listing-form/ListingEditForm.svelte';
	import ProviderCatalogueDrawer from '$lib/components/organisms/provider-catalogue-drawer/ProviderCatalogueDrawer.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import {
		listListings,
		createListing as createListingService,
		updateListing as updateListingService,
		deleteListing as deleteListingService
	} from '$lib/services/marketdata';
	import { ApiError } from '$lib/api/client';
	import { toast } from '$lib/stores/toast.svelte';
	import type { CreateListingRequest, Listing } from '$lib/api/types';
	import type { ListingFieldPatch } from '$lib/components/organisms/listing-form/listing-form.types';

	let listings = $state<Listing[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let createOpen = $state(false);
	let creating = $state(false);
	let catalogueOpen = $state(false);
	let filter = $state('');

	let editing = $state<Listing | null>(null);
	let saving = $state(false);

	let confirmingDelete = $state<Listing | null>(null);
	let deleting = $state(false);
	let deleteError = $state('');

	// Filtering happens in the browser: the page already holds every listing, so a
	// round trip would cost more than it saves.
	const visibleListings = $derived.by(() => {
		const needle = filter.trim().toLowerCase();
		if (!needle) return listings;
		return listings.filter(
			(listing) =>
				listing.symbol.toLowerCase().includes(needle) ||
				listing.name.toLowerCase().includes(needle) ||
				(listing.isin ?? '').toLowerCase().includes(needle)
		);
	});

	async function load() {
		loading = true;
		error = null;
		try {
			listings = await listListings();
		} catch {
			error = 'Failed to load listings';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	/** Replaces one listing in place, keeping the table's symbol ordering. */
	function upsert(listing: Listing) {
		listings = [...listings.filter((item) => item.id !== listing.id), listing].sort((a, b) =>
			a.symbol.localeCompare(b.symbol)
		);
	}

	async function createListing(body: CreateListingRequest) {
		const listing = await createListingService(body);
		upsert(listing);
		createOpen = false;
		toast.success(`Listing ${listing.symbol} created`);
	}

	async function removeListing() {
		const target = confirmingDelete;
		if (!target || deleting) return;
		deleting = true;
		deleteError = '';
		try {
			await deleteListingService(target.id);
			listings = listings.filter((item) => item.id !== target.id);
			confirmingDelete = null;
			toast.success(`Listing ${target.symbol} deleted`);
		} catch (cause) {
			if (cause instanceof ApiError && cause.status === 409) {
				deleteError =
					'This listing is used by a portfolio. Deleting it would remove the valuation history that depends on it, so it has to stay.';
			} else if (cause instanceof ApiError && cause.status === 404) {
				deleteError = 'This listing no longer exists. Close this and refresh the listings.';
			} else {
				deleteError = 'Could not delete the listing. Check your connection and try again.';
			}
		} finally {
			deleting = false;
		}
	}

	async function saveListing(patch: ListingFieldPatch) {
		if (!editing) return;
		const listing = await updateListingService({ id: editing.id, ...patch });
		upsert(listing);
		editing = null;
		toast.success(`Listing ${listing.symbol} updated`);
	}
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar title="Listings" />
	{/snippet}

	<PageContentTemplate>
		<div
			class="flex shrink-0 flex-wrap items-center justify-between gap-4 border-b border-slate-200 p-4"
		>
			<div>
				<h2 class="text-2xl">Market listings</h2>
				<p class="mt-1 text-sm text-slate-600">
					Manage instruments used in your portfolio and price history.
				</p>
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<div class="w-56">
					<SearchInput
						bind:value={filter}
						placeholder="Filter listings…"
						ariaLabel="Filter listings"
					/>
				</div>
				<Button variant="outline" intent="secondary" onclick={() => (catalogueOpen = true)}>
					Browse catalogue
				</Button>
				<Button disabled={loading} onclick={() => (createOpen = true)}>Add listing</Button>
			</div>
		</div>
		{#if error}
			<div role="alert" class="flex flex-wrap items-center gap-3 p-4">
				<p class="text-sm text-red-700">{error}. Check your connection and try again.</p>
				<Button variant="outline" onclick={load}>Retry loading</Button>
			</div>
		{/if}
		<DataTable
			rows={visibleListings}
			{loading}
			emptyText={error
				? 'Listings are unavailable. Retry loading above.'
				: filter.trim()
					? `No listings match "${filter.trim()}". Try Browse catalogue to add one.`
					: 'No listings yet. Add your first listing to get started.'}
			columns={[
				{ key: 'symbol', header: 'Symbol', value: (r: Listing) => r.symbol },
				{ key: 'name', header: 'Name', value: (r: Listing) => r.name },
				{ key: 'source', header: 'Source', value: (r: Listing) => r.source },
				{ key: 'exchange', header: 'Exchange', value: (r: Listing) => r.exchange ?? '—' },
				{ key: 'currency', header: 'Currency', value: (r: Listing) => r.currency ?? '—' },
				{ key: 'type', header: 'Type', value: (r: Listing) => r.type ?? '—' },
				{ key: 'actions', header: '', width: 'w-40', align: 'right', cell: actionsCell }
			]}
		/>
	</PageContentTemplate>
</AppShellTemplate>

<Dialog
	bind:open={createOpen}
	title="Add listing"
	size="lg"
	dismissible={!creating}
	closeOnBackdrop={!creating}
	closeOnEscape={!creating}
>
	{#if createOpen}
		<ListingForm
			bind:saving={creating}
			onSave={createListing}
			onCancel={() => (createOpen = false)}
		/>
	{/if}
</Dialog>

{#snippet actionsCell(row: Listing)}
	<span class="flex items-center justify-end gap-2">
		<Button variant="outline" size="sm" onclick={() => (editing = row)}>Edit</Button>
		<Button
			variant="ghost"
			intent="error"
			size="sm"
			onclick={() => {
				deleteError = '';
				confirmingDelete = row;
			}}
		>
			Delete
		</Button>
	</span>
{/snippet}

<!-- Deleting is irreversible and the API refuses the cases that would destroy history,
     so the confirm names the instrument and reports a refusal in place rather than as a
     toast the user would have to read before it disappeared. -->
<Dialog
	open={confirmingDelete !== null}
	title={confirmingDelete ? `Delete ${confirmingDelete.symbol}?` : 'Delete listing'}
	size="sm"
	dismissible={!deleting}
	closeOnBackdrop={!deleting}
	closeOnEscape={!deleting}
	onClose={() => (confirmingDelete = null)}
>
	{#if confirmingDelete}
		{@const target = confirmingDelete}
		<div class="space-y-4">
			<p class="text-slate-600">
				{target.name} will be removed, along with any price history stored for it. Listings held in a
				portfolio cannot be deleted.
			</p>
			{#if deleteError}<p role="alert" class="text-sm text-red-700">{deleteError}</p>{/if}
			<div class="flex justify-end gap-2 border-t border-slate-200 pt-4">
				<Button
					variant="ghost"
					intent="secondary"
					disabled={deleting}
					onclick={() => (confirmingDelete = null)}
				>
					Cancel
				</Button>
				<Button intent="error" disabled={deleting} loading={deleting} onclick={removeListing}>
					{deleting ? 'Deleting…' : 'Delete listing'}
				</Button>
			</div>
		</div>
	{/if}
</Dialog>

<Dialog
	open={editing !== null}
	title={editing ? `Edit ${editing.symbol}` : 'Edit listing'}
	size="lg"
	dismissible={!saving}
	closeOnBackdrop={!saving}
	closeOnEscape={!saving}
	onClose={() => (editing = null)}
>
	{#if editing}
		<!-- Keyed so pointing the dialog at another listing rebuilds the form rather than
		     leaving it holding the previous instrument's values. -->
		{#key editing.id}
			<ListingEditForm
				listing={editing}
				bind:saving
				onSave={saveListing}
				onCancel={() => (editing = null)}
			/>
		{/key}
	{/if}
</Dialog>

<ProviderCatalogueDrawer
	bind:open={catalogueOpen}
	onAdopted={(created) => {
		toast.success(`Added ${created.length} listing${created.length === 1 ? '' : 's'}`);
		void load();
	}}
/>
