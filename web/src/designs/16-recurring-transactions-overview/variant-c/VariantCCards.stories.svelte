<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RecurringCardsPage from './RecurringCardsPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#16 Recurring overview/Variant C: Cards',
		component: RecurringCardsPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<RecurringCardsPage />
</Story>

<Story name="Mobile" asChild parameters={{ viewport: { defaultViewport: 'mobile1' } }}>
	<RecurringCardsPage />
</Story>
