<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ProviderCatalogueDrawer from './ProviderCatalogueDrawer.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/ProviderCatalogueDrawer',
		component: ProviderCatalogueDrawer,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	let open = $state(false);
	let openAlphaVantage = $state(false);
</script>

<!--
	The drawer reads through the marketdata service, so on mock fixtures searching
	`asm` shows the interesting rows: one already tracked, one with no name, and one
	with no price history — all three unselectable. `vwrl` then demonstrates the
	metered provider search, because some VWRL listings exist only upstream.

	Adopting a row fills the outcome box above the table; searching again clears it, since it
	reports one batch and not the rows now on screen. The band stays until it is dismissed.
-->
<Story name="Marketstack catalogue" asChild>
	<div>
		<Button onclick={() => (open = true)}>Browse catalogue</Button>
		<ProviderCatalogueDrawer bind:open />
	</div>
</Story>

<Story name="Unsupported source" asChild>
	<div>
		<Button onclick={() => (openAlphaVantage = true)}>Browse Alpha Vantage</Button>
		<ProviderCatalogueDrawer bind:open={openAlphaVantage} source="alpha_vantage" />
	</div>
</Story>
