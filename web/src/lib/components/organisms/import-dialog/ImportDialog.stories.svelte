<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ImportDialog from './ImportDialog.svelte';
	import type { Vendor } from '$lib/api/types';

	const vendors: Vendor[] = [
		{
			id: 'vnd-degiro',
			name: 'DEGIRO',
			type: 'portfolio',
			active: true,
			import_disabled: false,
			created_at: '2025-01-04T09:00:00Z',
			updated_at: '2026-01-10T09:00:00Z'
		}
	];

	const { Story } = defineMeta({
		title: 'Organisms/ImportDialog',
		component: ImportDialog,
		tags: ['autodocs']
	});
</script>

<!--
	The dialog drives itself off the import services, so the stories cover the two states
	that exist before an upload: the form with a brokerage account to pick, and the form
	with none. Processing and result follow from a real upload; in fixture mode the mock
	import service reads the dropped CSV and reports what it found.
-->
<Story name="Playground" args={{ open: true, vendors }} />

<Story name="No brokerage account" args={{ open: true, vendors: [] }} />

<Story
	name="Mobile"
	args={{ open: true, vendors }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>
