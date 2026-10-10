<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RuleMatchPreview from './RuleMatchPreview.svelte';
	import { cashflowTransactions } from '$lib/data/fixtures/cashflow';

	const sample = cashflowTransactions.slice(0, 4);

	const { Story } = defineMeta({
		title: 'Molecules/RuleMatchPreview',
		component: RuleMatchPreview,
		tags: ['autodocs']
	});
</script>

<Story
	name="Playground"
	args={{ matching: 23, scanned: 1284, sample, contains: 'Credit card' }}
/>

<!-- First read of a rule being typed: a placeholder, not a zero that would be a claim. -->
<Story
	name="Checking"
	args={{ matching: 0, scanned: 0, sample: [], contains: 'Credit card', loading: true }}
/>

<!-- A rule written for next month's import catches nothing today, and that is fine. -->
<Story
	name="No matches"
	args={{ matching: 0, scanned: 1284, sample: [], contains: 'Credit card settlement 2026' }}
/>

<Story
	name="Error"
	args={{
		matching: 0,
		scanned: 0,
		sample: [],
		contains: 'Credit card',
		error: 'The check did not come back. Try again.'
	}}
/>

<Story
	name="Mobile"
	args={{ matching: 23, scanned: 1284, sample, contains: 'Credit card' }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>
