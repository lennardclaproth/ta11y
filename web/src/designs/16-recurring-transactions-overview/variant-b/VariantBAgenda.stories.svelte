<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RecurringAgendaPage from './RecurringAgendaPage.svelte';

	const { Story } = defineMeta({
		title: 'Designs/#16 Recurring overview/Variant B: Agenda',
		component: RecurringAgendaPage,
		parameters: { layout: 'fullscreen' }
	});
</script>

<Story name="Default" asChild>
	<RecurringAgendaPage />
</Story>

<Story name="Mobile" asChild parameters={{ viewport: { defaultViewport: 'mobile1' } }}>
	<RecurringAgendaPage />
</Story>
