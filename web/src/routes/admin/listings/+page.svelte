<script lang="ts">
	import { onMount } from 'svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import ListingForm from '$lib/components/organisms/listing-form/ListingForm.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { listListings, createListing as createListingService } from '$lib/services/marketdata';
	import { toast } from '$lib/stores/toast.svelte';
	import { adminMode } from '$lib/stores/admin.svelte';
	import type { CreateListingRequest, Listing } from '$lib/api/types';

	let listings = $state<Listing[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let createOpen = $state(false);
	let creating = $state(false);

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
		<TopNavbar
			title="Listings"
			accountName="Admin User"
			adminMode={adminMode.enabled}
			onAdminToggle={(v) => adminMode.set(v)}
		/>
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
			<Button disabled={loading} onclick={() => (createOpen = true)}>Add listing</Button>
		</div>
		{#if error}
			<div role="alert" class="flex flex-wrap items-center gap-3 p-4">
				<p class="text-sm text-red-700">{error}. Check your connection and try again.</p>
				<Button variant="outline" onclick={load}>Retry loading</Button>
			</div>
		{/if}
		<DataTable
			rows={listings}
			{loading}
			emptyText={error
				? 'Listings are unavailable. Retry loading above.'
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
