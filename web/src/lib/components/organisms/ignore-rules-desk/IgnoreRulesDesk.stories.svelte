<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import IgnoreRulesDesk from './IgnoreRulesDesk.svelte';
	import { cashflowTransactions } from '$lib/data/fixtures/cashflow';
	import { ignoreRules } from '$lib/data/fixtures/ignoreRules';
	import { deskPanes, draftFromRule, emptyDraft } from './ignore-rules-desk.types';
	import type { IgnoreRulePreview } from '$lib/api/types';

	const open = draftFromRule(ignoreRules[1]);

	const preview: IgnoreRulePreview = {
		matching: 23,
		not_yet_ignored: 23,
		scanned: 1284,
		sample: cashflowTransactions.slice(0, 4)
	};

	const nothing: IgnoreRulePreview = {
		matching: 0,
		not_yet_ignored: 0,
		scanned: 1284,
		sample: []
	};

	const banks = [
		{ value: 'ing', label: 'ING' },
		{ value: 'n26', label: 'N26' }
	];

	const { Story } = defineMeta({
		title: 'Organisms/IgnoreRulesDesk',
		component: IgnoreRulesDesk,
		tags: ['autodocs'],
		argTypes: {
			pane: { control: 'select', options: deskPanes }
		}
	});
</script>

<Story
	name="Playground"
	args={{ rules: ignoreRules, draft: open, preview, bankOptions: banks, pane: 'detail' }}
/>

<!-- Wide screens show both halves; `pane` only decides what a narrow one shows. -->
<Story
	name="Index"
	args={{ rules: ignoreRules, draft: open, preview, bankOptions: banks, pane: 'list' }}
/>

<Story name="Loading" args={{ rules: [], draft: null, loading: true, bankOptions: banks }} />

<!-- No rules yet: the right half explains what a rule is rather than showing a dead form. -->
<Story name="Empty" args={{ rules: [], draft: null, bankOptions: banks }} />

<Story
	name="Error"
	args={{
		rules: [],
		draft: null,
		error: 'The rules did not load. Check your connection.',
		bankOptions: banks
	}}
/>

<!-- A rule written for next month's import catches nothing today; applying is offered as 0. -->
<Story
	name="No matches"
	args={{
		rules: ignoreRules,
		draft: { ...open, contains: 'Credit card settlement 2026' },
		preview: nothing,
		bankOptions: banks,
		pane: 'detail'
	}}
/>

<Story
	name="New rule"
	args={{
		rules: ignoreRules,
		draft: emptyDraft(),
		preview: nothing,
		bankOptions: banks,
		pane: 'detail'
	}}
/>

<Story
	name="Refused by the API"
	args={{
		rules: ignoreRules,
		draft: { ...open, contains: 'to' },
		preview: nothing,
		bankOptions: banks,
		pane: 'detail',
		fieldErrors: { contains: 'contains must be at least 3 characters' }
	}}
/>

<Story
	name="Mobile"
	args={{ rules: ignoreRules, draft: open, preview, bankOptions: banks, pane: 'detail' }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>
