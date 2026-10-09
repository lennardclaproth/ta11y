<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ImportReviewScreen from './ImportReviewScreen.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#23 Ignore rules/Variant C: Import review',
		component: ImportReviewScreen
	});
</script>

<Story name="Default" asChild>
	<ImportReviewScreen />
</Story>

<Story name="Mobile" asChild>
	<ImportReviewScreen flow />
</Story>
