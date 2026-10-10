<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import CountStat from './CountStat.svelte';
	import { countStatSizes } from './count-stat.types';

	const { Story } = defineMeta({
		title: 'Molecules/CountStat',
		component: CountStat,
		tags: ['autodocs'],
		argTypes: {
			size: { control: 'select', options: countStatSizes }
		}
	});
</script>

<Story
	name="Playground"
	args={{ label: 'Auto-ignored', value: 37, note: 'kept out of your totals' }}
/>

<Story name="New" args={{ label: 'New', value: 128, note: 'added to your ledger' }} />

<Story name="Duplicates" args={{ label: 'Duplicates', value: 42, note: 'already imported earlier' }} />

<!-- Nothing matched is a real outcome, and zero is the honest figure for it. -->
<Story name="Nothing ignored" args={{ label: 'Auto-ignored', value: 0, note: 'no rule matched' }} />

<!-- While the import is still running there is no number yet, so none is claimed. -->
<Story
	name="Still running"
	args={{ label: 'Auto-ignored', value: 0, note: 'kept out of your totals', loading: true }}
/>

<Story name="Small" args={{ label: 'Matches', value: 23, size: 'sm' }} />

<Story
	name="Mobile"
	args={{ label: 'Auto-ignored', value: 37, note: 'kept out of your totals' }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>
