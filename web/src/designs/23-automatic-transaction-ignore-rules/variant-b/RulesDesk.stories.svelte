<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RulesDeskScreen from './RulesDeskScreen.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#23 Ignore rules/Variant B: Rules desk',
		component: RulesDeskScreen
	});
</script>

<Story name="Default" asChild>
	<RulesDeskScreen />
</Story>

<Story name="Mobile" asChild>
	<RulesDeskScreen mobileView="detail" flow />
</Story>
