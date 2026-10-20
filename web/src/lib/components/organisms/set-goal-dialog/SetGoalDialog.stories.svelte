<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import SetGoalDialog from './SetGoalDialog.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/SetGoalDialog',
		component: SetGoalDialog,
		tags: ['autodocs'],
		argTypes: {
			percent: { control: { type: 'number', min: 0, max: 100 } },
			saving: { control: 'boolean' }
		}
	});
</script>

<Story
	name="Playground"
	args={{
		open: true,
		percent: 30,
		exampleIncome: 5000,
		effectiveFrom: '2026-07-01'
	}}
/>

<!-- First run: there is no income marked yet, so the worked example is left out rather
     than shown as €0. -->
<Story name="Setting the first goal">
	<SetGoalDialog open percent={30} effectiveFrom="2026-07-01" title="Monthly goal" />
</Story>

<Story name="Saving">
	<SetGoalDialog open percent={35} exampleIncome={5000} effectiveFrom="2026-07-01" saving />
</Story>

<Story name="Refused">
	<SetGoalDialog
		open
		percent={120}
		exampleIncome={5000}
		effectiveFrom="2026-07-01"
		error="A goal is a whole percentage between 0 and 100."
	/>
</Story>
