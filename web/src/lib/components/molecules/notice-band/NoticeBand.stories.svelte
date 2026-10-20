<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import type { ComponentProps } from 'svelte';

	import NoticeBand from './NoticeBand.svelte';
	import { noticeBandGutters, noticeBandSurfaces, noticeIntents } from './notice-band.types';

	type NoticeBandProps = ComponentProps<typeof NoticeBand>;

	const { Story } = defineMeta({
		title: 'Molecules/NoticeBand',
		component: NoticeBand,
		tags: ['autodocs'],
		argTypes: {
			intent: { control: 'select', options: noticeIntents },
			surface: { control: 'select', options: noticeBandSurfaces },
			gutter: { control: 'select', options: noticeBandGutters },
			title: { control: 'text' },
			dismissible: { control: 'boolean' }
		}
	});
</script>

<script lang="ts">
	import Button from '$lib/components/atoms/button/Button.svelte';
</script>

{#snippet playground(args: NoticeBandProps)}
	<div class="bg-taupe-100 py-6">
		<NoticeBand {...args}>Transaction created</NoticeBand>
	</div>
{/snippet}

<Story
	name="Playground"
	args={{ intent: 'success', gutter: 'page', dismissible: true }}
	template={playground}
/>

<Story name="Intents" asChild>
	<div class="flex flex-col border border-slate-300 bg-taupe-50 p-4">
		{#each noticeIntents as intent, index (intent)}
			<NoticeBand {intent} gutter="panel" class={index > 0 ? '-mt-px' : ''}>
				This is a {intent} message describing what happened.
			</NoticeBand>
		{/each}
	</div>
</Story>

<Story name="With title" asChild>
	<div class="bg-taupe-100 py-6">
		<NoticeBand intent="error" title="This transaction was not saved">
			A transaction with the same amount and description already exists on Mar 2, 2026.
		</NoticeBand>
	</div>
</Story>

{#snippet retry()}
	<Button size="sm" variant="outline" intent="secondary">Try again</Button>
{/snippet}

<Story name="Error that stays" asChild>
	<div class="border border-slate-300 bg-taupe-50 p-4">
		<NoticeBand intent="error" gutter="panel" action={retry}>Failed to load transactions</NoticeBand>
	</div>
</Story>

<Story name="Inset surface" asChild>
	<!-- Inside a white dialog or drawer, where a white band would vanish. -->
	<div class="rounded-2xl bg-white p-5 shadow-md">
		<NoticeBand intent="warning" surface="inset" gutter="dialog" class="-mx-5">
			Some rows were skipped during the import
		</NoticeBand>
	</div>
</Story>

<Story name="Stacked" asChild>
	<div class="bg-taupe-100 py-6">
		<NoticeBand intent="success" dismissible onDismiss={() => {}}>Transaction created</NoticeBand>
		<NoticeBand intent="info" dismissible onDismiss={() => {}} class="-mt-px">
			Portfolio rebuild started
		</NoticeBand>
	</div>
</Story>
