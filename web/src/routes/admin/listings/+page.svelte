<script lang="ts">
	import { onMount } from 'svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import ListingForm from '$lib/components/organisms/listing-form/ListingForm.svelte';
	import ProviderCatalogueDrawer from '$lib/components/organisms/provider-catalogue-drawer/ProviderCatalogueDrawer.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import { listListings, createListing as createListingService } from '$lib/services/marketdata';
	import { toast } from '$lib/stores/toast.svelte';
	import type { CreateListingRequest, Listing } from '$lib/api/types';

	let listings = $state<Listing[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let createOpen = $state(false);
	let creating = $state(false);
	let catalogueOpen = $state(false);
	let filter = $state('');

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

	async function createListing(body: CreateListingRequest) {
		const listing = await createListingService(body);
		listings = [...listings.filter((item) => item.id !== listing.id), listing].sort((a, b) =>
			a.symbol.localeCompare(b.symbol)
		);
		createOpen = false;
		toast.success(`Listing ${listing.symbol} created`);
	}
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar title="Listings" accountName="Admin User" />
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
				{ key: 'type', header: 'Type', value: (r: Listing) => r.type ?? '—' }
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

<ProviderCatalogueDrawer
	bind:open={catalogueOpen}
	onAdopted={(created) => {
		toast.success(`Added ${created.length} listing${created.length === 1 ? '' : 's'}`);
		void load();
	}}
/>
