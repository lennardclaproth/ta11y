<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import TopNavbar from './TopNavbar.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/TopNavbar',
		component: TopNavbar,
		tags: ['autodocs'],
		argTypes: {
			showPeriod: { control: 'boolean' }
		}
	});
</script>

<!--
  The bar is navigation plus the account overview. Searching, the period and the page's actions
  live with the table and the charts they act on, so nothing here changes the page below it.
  Without a session the overview shows the signed-out fallbacks; see Organisms/AccountOverview
  for its own states.
-->

<Story name="Default" asChild>
	<div class="bg-taupe-100">
		<TopNavbar />
	</div>
</Story>

<!-- Admin pages have no charts, so no period is offered. -->
<Story name="Without a period" asChild>
	<div class="bg-taupe-100">
		<TopNavbar showPeriod={false} />
	</div>
</Story>
