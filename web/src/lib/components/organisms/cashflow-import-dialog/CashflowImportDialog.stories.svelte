<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import CashflowImportDialog from './CashflowImportDialog.svelte';
	import type { Vendor } from '$lib/api/types';

	const vendors: Vendor[] = [
		{
			id: 'vnd-ing',
			name: 'ING',
			type: 'bank',
			active: true,
			import_disabled: false,
			created_at: '2025-01-04T09:00:00Z',
			updated_at: '2026-01-10T09:00:00Z'
		},
		{
			id: 'vnd-n26',
			name: 'N26',
			type: 'bank',
			active: true,
			import_disabled: false,
			created_at: '2025-01-04T09:00:00Z',
			updated_at: '2026-01-10T09:00:00Z'
		}
	];

	const { Story } = defineMeta({
		title: 'Organisms/CashflowImportDialog',
		component: CashflowImportDialog,
		tags: ['autodocs']
	});
</script>

<!--
	The dialog drives itself off the import services, so the stories cover the states that
	exist before an upload. Processing and result follow from a real upload; in fixture mode
	the mock import service reads the dropped CSV and reports what it found.
-->
<Story name="Playground" args={{ open: true, vendors, onBrokerageImport: () => {} }} />

<Story name="No bank account" args={{ open: true, vendors: [] }} />

<Story
	name="Mobile"
	args={{ open: true, vendors, onBrokerageImport: () => {} }}
	parameters={{ viewport: { defaultViewport: 'mobile1' } }}
/>
