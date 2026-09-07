<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ListingSearchSelect from './ListingSearchSelect.svelte';
	import { listings } from '$lib/data/fixtures/marketdata';
	import type { ListingSearchRow } from '$lib/api/types';

	// Deterministic, offline search fns injected so stories don't depend on the service/network.
	// Fixtures are listings, so they stand in as tracked rows.
	const asRow = (listing: (typeof listings)[number]): ListingSearchRow => ({
		...listing,
		exchange_mic: null,
		tracked: true,
		has_eod: true,
		adoptable: false,
		adoptable_reason: 'this instrument is already in your listings'
	});
	const stubSearch = (q: string): Promise<ListingSearchRow[]> =>
		Promise.resolve(
			listings
				.filter(
					(l) =>
						l.symbol.toLowerCase().includes(q.toLowerCase()) ||
						l.name.toLowerCase().includes(q.toLowerCase())
				)
				.map(asRow)
		);
	const neverResolves = (): Promise<ListingSearchRow[]> => new Promise(() => {});
	const empty = (): Promise<ListingSearchRow[]> => Promise.resolve([]);
	const fails = (): Promise<ListingSearchRow[]> => Promise.reject(new Error('boom'));

	const { Story } = defineMeta({
		title: 'Molecules/ListingSearchSelect',
		component: ListingSearchSelect,
		tags: ['autodocs']
	});
</script>

<Story name="Results" asChild>
	<div class="min-h-72 w-80">
		<ListingSearchSelect search={stubSearch} query="a" debounceMs={0} />
	</div>
</Story>

<Story name="Loading" asChild>
	<div class="min-h-48 w-80">
		<ListingSearchSelect search={neverResolves} query="aapl" debounceMs={0} />
	</div>
</Story>

<Story name="Empty" asChild>
	<div class="min-h-48 w-80">
		<ListingSearchSelect search={empty} query="zzz" debounceMs={0} />
	</div>
</Story>

<Story name="Error" asChild>
	<div class="min-h-48 w-80">
		<ListingSearchSelect search={fails} query="err" debounceMs={0} />
	</div>
</Story>
