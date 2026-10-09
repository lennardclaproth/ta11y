<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import FinalPage from './FinalPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#15 Monthly wealth goal/Final',
		component: FinalPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<!-- The rail scrolled to the wealth-goal group: the running month, the streak and the standing. -->
<Story name="Default" args={{ startAt: 'month' }} />

<Story name="Mobile" args={{ startAt: 'month' }} />

<!-- Where the page actually opens: the three Cashflow charts, with the goal one scroll away. -->
<Story name="Cashflow cards" args={{ startAt: 'trend' }} />

<!-- Pointing transactions at income or wealth, in bulk. -->
<Story name="Marking transactions" args={{ startAt: 'trend', selection: true }} />

<Story name="Marking transactions Mobile" args={{ startAt: 'trend', selection: true }} />

<!-- Arriving from "5 transactions in July 2026 are not assigned yet". -->
<Story name="Open transactions of a month" args={{ startAt: 'month', focusMonth: 'July 2026' }} />

<Story name="Setting the goal" args={{ startAt: 'month', dialog: true }} />

<Story name="Loading" args={{ startAt: 'month', state: 'loading' }} />

<!-- No goal set yet: the goal card invites you to set one instead of scoring a zero. -->
<Story name="Empty" args={{ startAt: 'month', state: 'no-goal' }} />

<Story name="Error" args={{ startAt: 'month', state: 'error' }} />

<Story name="No matches" args={{ startAt: 'trend', noMatches: true }} />
