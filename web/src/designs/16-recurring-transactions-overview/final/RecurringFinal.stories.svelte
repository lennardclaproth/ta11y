<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RecurringOverviewPage from './RecurringOverviewPage.svelte';
	import CashflowMarkPage from './CashflowMarkPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#16 Recurring overview/Final',
		component: RecurringOverviewPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<RecurringOverviewPage />
</Story>

<Story name="Mobile" asChild parameters={{ viewport: { defaultViewport: 'mobile1' } }}>
	<RecurringOverviewPage />
</Story>

<Story name="Loading" asChild>
	<RecurringOverviewPage situation="loading" />
</Story>

<Story name="Empty" asChild>
	<RecurringOverviewPage situation="empty" />
</Story>

<Story name="Error" asChild>
	<RecurringOverviewPage situation="error" />
</Story>

<Story name="No matches" asChild>
	<RecurringOverviewPage situation="no-matches" />
</Story>

<Story name="Income" asChild>
	<RecurringOverviewPage group="income" />
</Story>

<Story name="Suggestions" asChild>
	<RecurringOverviewPage suggestionsOpen />
</Story>

<Story name="Item" asChild>
	<RecurringOverviewPage itemOpen />
</Story>

<Story name="End from month" asChild>
	<RecurringOverviewPage itemOpen endOpen />
</Story>

<Story name="Mark as recurring" asChild>
	<CashflowMarkPage />
</Story>
