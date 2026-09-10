<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { ApiError } from '$lib/api/client';
	import type { Listing } from '$lib/api/types';
	import ListingEditForm from './ListingEditForm.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/ListingEditForm',
		component: ListingEditForm
	});

	const listing: Listing = {
		id: 'lst-asml',
		symbol: 'ASML',
		name: 'ASML Holding NV',
		source: 'marketstack',
		description: null,
		exchange: 'AEX',
		region: 'Netherlands',
		currency: 'EUR',
		isin: 'NL0010273215',
		ticker: 'ASML.AS',
		type: 'stock',
		created_at: '2026-06-01T09:00:00Z',
		updated_at: '2026-06-01T09:00:00Z'
	};

	/** A listing the provider gave almost nothing for — the fields it can fill in. */
	const sparse: Listing = {
		...listing,
		id: 'lst-sparse',
		symbol: 'ASMB',
		name: 'Assembly Biosciences Inc',
		exchange: null,
		region: null,
		currency: null,
		isin: null,
		ticker: null,
		type: null
	};

	const noop = { onSave: async () => {}, onCancel: () => {} };
</script>

<!-- Save starts disabled: the endpoint rejects an empty patch, so an untouched form
     has nothing to send. -->
<Story name="Populated" args={{ listing, ...noop }} />

<Story name="Sparse metadata" args={{ listing: sparse, ...noop }} />

<Story
	name="Rejected"
	args={{
		listing,
		onSave: async () => {
			throw new ApiError(400, 'Invalid', { currency: 'currency is invalid' });
		},
		onCancel: () => {}
	}}
/>

<Story name="Saving" args={{ listing, saving: true, ...noop }} />
