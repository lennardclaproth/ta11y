<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Final from './Final.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#22 Alerts and toasts/Final',
		component: Final,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<Final />
</Story>

<Story name="Mobile" asChild>
	<Final />
</Story>

<Story name="Loading" asChild>
	<Final scenario="loading" />
</Story>

<Story name="Empty" asChild>
	<Final scenario="empty" />
</Story>

<Story name="Error" asChild>
	<Final scenario="error" />
</Story>

<Story name="No matches" asChild>
	<Final scenario="no-matches" />
</Story>
