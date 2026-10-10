<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import IgnoreRuleListItem from './IgnoreRuleListItem.svelte';
	import { ignoreRules } from '$lib/data/fixtures/ignoreRules';

	const [savings, creditCard, incoming, roundUp] = ignoreRules;

	const { Story } = defineMeta({
		title: 'Molecules/IgnoreRuleListItem',
		component: IgnoreRuleListItem,
		tags: ['autodocs']
	});
</script>

<Story name="Playground" args={{ rule: creditCard }} />

<!-- The one open on the right carries the selection, and says so to a screen reader. -->
<Story name="Open" args={{ rule: creditCard, active: true }} />

<Story name="All banks" args={{ rule: savings }} />

<Story name="Matches on the note" args={{ rule: incoming }} />

<!-- Switched off is a word, not only a lighter shade. -->
<Story name="Switched off" args={{ rule: roundUp }} />

<Story
	name="Mobile"
	args={{ rule: creditCard }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>
