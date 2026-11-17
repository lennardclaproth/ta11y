<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import IgnoredByRuleGroup from './IgnoredByRuleGroup.svelte';
	import { cashflowTransactions } from '$lib/data/fixtures/cashflow';
	import { ignoreRules } from '$lib/data/fixtures/ignoreRules';
	import type { IgnoredRuleGroup } from '$lib/api/types';

	const rows = cashflowTransactions.slice(0, 3).map((tx) => ({ ...tx, ignored: true }));

	const caught: IgnoredRuleGroup = {
		rule: ignoreRules[1],
		total: 3,
		transactions: rows
	};

	const truncated: IgnoredRuleGroup = {
		rule: ignoreRules[0],
		total: 12,
		transactions: rows
	};

	const withRestored: IgnoredRuleGroup = {
		rule: ignoreRules[1],
		total: 3,
		transactions: [rows[0], { ...rows[1], ignored: false }, rows[2]]
	};

	const switchedOff: IgnoredRuleGroup = {
		rule: ignoreRules[3],
		total: 2,
		transactions: rows.slice(0, 2)
	};

	const { Story } = defineMeta({
		title: 'Organisms/IgnoredByRuleGroup',
		component: IgnoredByRuleGroup,
		tags: ['autodocs']
	});
</script>

<Story name="Playground" args={{ group: caught }} />

<!-- The rest of a wide group is a click away rather than a page of rows nobody asked for. -->
<Story name="More in the group" args={{ group: truncated }} />

<!-- A row put back by hand says so: rules leave it alone from then on. -->
<Story name="One restored by hand" args={{ group: withRestored }} />

<!-- A rule already switched off keeps its harvest, so restoring is still offered. -->
<Story name="Rule switched off" args={{ group: switchedOff }} />

<Story name="Busy" args={{ group: caught, busy: true }} />

<Story
	name="Mobile"
	args={{ group: truncated }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>
