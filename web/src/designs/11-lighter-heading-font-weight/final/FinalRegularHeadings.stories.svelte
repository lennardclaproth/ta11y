<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';

	import FinalRegularHeadings from './FinalRegularHeadings.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#11 Lighter heading weight/Final',
		component: FinalRegularHeadings,
		parameters: { layout: 'fullscreen' }
	});
</script>

<!-- Reference: the comparison band plus every Garamond heading in the app. -->
<Story name="Default" />

<!-- Same content at 390px; a thin serif has to survive the small viewport too. -->
<Story name="Mobile" />

<!-- The data states exist to show the headings never change weight between them: only the
     content underneath does. Error also brings in the bare <h1> "Can't reach the server". -->
<Story name="Loading" args={{ state: 'loading', showBand: false }} />

<Story name="Empty" args={{ state: 'empty', showBand: false }} />

<Story name="Error" args={{ state: 'error', showBand: false }} />

<Story name="No matches" args={{ state: 'no-matches', showBand: false }} />
