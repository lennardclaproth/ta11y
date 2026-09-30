<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Final from './Final.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#13 Transaction date/Final',
		component: Final,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<Final scene="create" />
</Story>

<Story name="Mobile" asChild>
	<Final scene="create" />
</Story>

<Story name="Backdated" asChild>
	<Final scene="create-backdated" />
</Story>

<Story name="Change date" asChild>
	<Final scene="edit" />
</Story>

<Story name="Change date Mobile" asChild>
	<Final scene="edit" />
</Story>

<Story name="Imported transaction" asChild>
	<Final scene="edit-imported" />
</Story>

<Story name="Saving" asChild>
	<Final scene="edit-saving" />
</Story>

<Story name="Error" asChild>
	<Final scene="edit-error" />
</Story>

<Story name="Portfolio rebuild busy" asChild>
	<Final scene="portfolio-busy" />
</Story>

<Story name="Loading" asChild>
	<Final scene="loading" />
</Story>

<Story name="Empty" asChild>
	<Final scene="empty" />
</Story>

<Story name="No matches" asChild>
	<Final scene="no-matches" />
</Story>
