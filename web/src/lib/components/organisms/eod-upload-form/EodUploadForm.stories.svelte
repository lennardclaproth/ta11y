<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { ApiError } from '$lib/api/client';
	import type { ListingSearchRow } from '$lib/api/types';
	import EodUploadForm from './EodUploadForm.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/EodUploadForm',
		component: EodUploadForm
	});

	const manual: ListingSearchRow = {
		id: 'lst-bnd',
		symbol: 'BND-WORLD',
		name: 'Brand New Day World Index Fund',
		source: 'brandnewday',
		exchange: null,
		exchange_mic: null,
		currency: 'EUR',
		isin: 'NL0011225305',
		tracked: true,
		has_eod: true,
		adoptable: false,
		adoptable_reason: null
	};

	const noop = { onUpload: async () => {}, onCancel: () => {} };
</script>

<!-- The ordinary case: a manually ingested fund waiting for its NAV file. -->
<Story name="Ready" args={{ listing: manual, ...noop }} />

<!-- The refusal that matters: a listing priced by an API provider will never accept a
     file, so the message says so rather than inviting a retry. -->
<Story
	name="Not a manual provider"
	args={{
		listing: { ...manual, symbol: 'ASML', name: 'ASML Holding NV', source: 'market_stack' },
		onUpload: async () => {
			throw new ApiError(422, 'Not manual', { listing_id: 'import provider is not manual' });
		},
		onCancel: () => {}
	}}
/>

<Story
	name="File rejected"
	args={{
		listing: manual,
		onUpload: async () => {
			throw new ApiError(400, 'Bad file', { file: 'import file is required' });
		},
		onCancel: () => {}
	}}
/>

<Story name="Uploading" args={{ listing: manual, uploading: true, ...noop }} />
