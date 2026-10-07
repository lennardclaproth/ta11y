<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import type { ComponentProps } from 'svelte';
	import TransactionDateHeader from './TransactionDateHeader.svelte';

	type HeaderProps = ComponentProps<typeof TransactionDateHeader>;

	const { Story } = defineMeta({
		title: 'Molecules/TransactionDateHeader',
		component: TransactionDateHeader,
		tags: ['autodocs'],
		argTypes: {
			disabled: { control: 'boolean' }
		}
	});

	// Fixed days so the stories read the same whenever they are opened.
	const today = '2026-09-30';
	const monthsBack = '2026-07-14';
</script>

<script lang="ts">
	let value = $state(monthsBack);
</script>

{#snippet playground(args: HeaderProps)}
	<div class="min-h-80 max-w-lg rounded-2xl bg-white px-5 pt-4 pb-5">
		<TransactionDateHeader {...args} />
	</div>
{/snippet}

<Story name="Playground" args={{ value: today, today }} template={playground} />

<Story name="Today" asChild>
	<div class="max-w-lg rounded-2xl bg-white px-5 pt-4 pb-5">
		<TransactionDateHeader value={today} {today} />
	</div>
</Story>

<!-- Backdated: the distance is the plain-language check on a date months back. -->
<Story name="Months ago" asChild>
	<div class="max-w-lg rounded-2xl bg-white px-5 pt-4 pb-5">
		<TransactionDateHeader bind:value {today} />
	</div>
</Story>

<Story name="Saving" asChild>
	<div class="max-w-lg rounded-2xl bg-white px-5 pt-4 pb-5">
		<TransactionDateHeader value={monthsBack} {today} disabled />
	</div>
</Story>

<Story name="Calendar open" asChild>
	<div class="min-h-96 max-w-lg rounded-2xl bg-white px-5 pt-4 pb-5">
		<TransactionDateHeader value={monthsBack} {today} open />
	</div>
</Story>
