<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import AssetClassDrawer from './AssetClassDrawer.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { assetClassDetails } from '$lib/data/fixtures/assets';
	import type { NoticeMessage } from '$lib/components/molecules/notice-band/notice-band.types';

	const stocks = assetClassDetails['cls-stocks'];
	const crypto = assetClassDetails['cls-crypto'];

	const { Story } = defineMeta({
		title: 'Organisms/AssetClassDrawer',
		component: AssetClassDrawer,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	let open = $state(false);
	let openLoading = $state(false);
	let openMixed = $state(false);
	let openNotice = $state(false);
	let notice = $state<NoticeMessage | null>({
		intent: 'error',
		message: 'Saved, but the class could not be refreshed'
	});
</script>

<Story name="With details" asChild>
	<div>
		<Button onclick={() => (open = true)}>Open asset class</Button>
		<AssetClassDrawer bind:open details={stocks} />
	</div>
</Story>

<Story name="Daily-priced items" asChild>
	<div>
		<Button onclick={() => (openMixed = true)}>Open mixed class</Button>
		<AssetClassDrawer
			bind:open={openMixed}
			details={crypto}
			onOpenClass={() => {}}
			onAddItem={() => {}}
			onOpenItem={() => {}}
		/>
	</div>
</Story>

<!-- Adding an item happens over this drawer, which stays open, so a failed refresh afterwards is
     reported in the drawer instead of as a toast behind its scrim. -->
<Story name="Refresh failed after adding an item" asChild>
	<div>
		<Button onclick={() => (openNotice = true)}>Open with a notice</Button>
		<AssetClassDrawer
			bind:open={openNotice}
			details={stocks}
			{notice}
			onDismissNotice={() => (notice = null)}
			onOpenClass={() => {}}
			onAddItem={() => {}}
		/>
	</div>
</Story>

<Story name="Loading" asChild>
	<div>
		<Button onclick={() => (openLoading = true)}>Open (loading)</Button>
		<AssetClassDrawer bind:open={openLoading} loading />
	</div>
</Story>
