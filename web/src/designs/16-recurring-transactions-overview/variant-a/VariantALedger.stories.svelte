<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RecurringLedgerPage from './RecurringLedgerPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#16 Recurring overview/Variant A: Ledger',
		component: RecurringLedgerPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<RecurringLedgerPage />
</Story>

<Story name="Mobile" asChild parameters={{ viewport: { defaultViewport: 'mobile1' } }}>
	<RecurringLedgerPage />
</Story>

<Story name="Empty" asChild>
	<RecurringLedgerPage situation="empty" />
</Story>
