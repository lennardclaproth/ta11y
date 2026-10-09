<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ClassPage from './ClassPage.svelte';
	import AssetsWithDrawer from './AssetsWithDrawer.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#14 Daily-priced asset holdings/Final',
		component: ClassPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<!-- The class page: item ledger on the left, the selected item as an article on the right. -->
<Story name="Default" asChild>
	<ClassPage />
</Story>

<!-- Narrow screens land on the ledger, where every item opens its performance in place. -->
<Story name="Mobile" asChild>
	<ClassPage view="ledger" />
</Story>

<Story name="One item on Mobile" asChild>
	<ClassPage view="item" />
</Story>

<!-- Assets itself keeps its shape; the class drawer is the quick performance overview. -->
<Story name="Class drawer on Assets" asChild>
	<AssetsWithDrawer />
</Story>

<Story name="Class drawer on Assets, Mobile" asChild>
	<AssetsWithDrawer />
</Story>

<Story name="Add item, pick what you own" asChild>
	<ClassPage addOpen addStep="quote" />
</Story>

<Story name="Add item, no matches" asChild>
	<ClassPage addOpen addStep="quote" quoteQuery="tesla" />
</Story>

<Story name="Add item, first purchase" asChild>
	<ClassPage addOpen addStep="purchase" />
</Story>

<Story name="Add item, save failed" asChild>
	<ClassPage addOpen addStep="purchase" saveError />
</Story>

<Story name="Loading" asChild>
	<ClassPage loading />
</Story>

<Story name="Empty" asChild>
	<ClassPage empty />
</Story>

<Story name="Error" asChild>
	<ClassPage error />
</Story>

<Story name="Stale prices" asChild>
	<ClassPage stale />
</Story>
