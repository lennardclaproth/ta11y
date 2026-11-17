<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import type { BackupDataset, BackupResult } from '$lib/api/types';
	import { backupPreviews } from '$lib/data/fixtures/backup';
	import ExportImportDialog from './ExportImportDialog.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/ExportImportDialog',
		component: ExportImportDialog,
		tags: ['autodocs'],
		argTypes: {
			open: { control: 'boolean' },
			admin: { control: 'boolean' }
		}
	});

	const wait = () => new Promise((resolve) => setTimeout(resolve, 600));

	const exported = async () => {
		await wait();
	};

	/** Reports the fixture counts for whichever dataset is chosen. */
	const merges = async (dataset: BackupDataset, _file: File, dryRun: boolean) => {
		await wait();
		return { ...backupPreviews[dataset], dry_run: dryRun } as BackupResult;
	};

	const alreadyHere = async (dataset: BackupDataset, _file: File, dryRun: boolean) => {
		await wait();
		const result = structuredClone(backupPreviews[dataset]) as unknown as Record<string, unknown>;
		for (const [key, value] of Object.entries(result)) {
			if (value && typeof value === 'object' && 'new' in value) {
				const count = value as { new: number; existing: number };
				result[key] = { new: 0, existing: count.new + count.existing };
			}
		}
		return { ...result, dry_run: dryRun } as unknown as BackupResult;
	};

	const missingVendor = async (dataset: BackupDataset, _file: File, dryRun: boolean) => {
		await wait();
		return {
			...backupPreviews[dataset],
			dry_run: dryRun,
			missing: { vendors: ['Bunq'], listings: [{ symbol: 'IWDA.AS', source: 'market_stack' }] }
		} as BackupResult;
	};

	const wrongKind = async () => {
		await wait();
		throw Object.assign(new Error('Request failed with status 422'), {
			status: 422,
			body: { file: 'backup: wrong kind of file: holds "portfolio", expected "cashflow"' }
		});
	};

	const importFails = async (dataset: BackupDataset, file: File, dryRun: boolean) => {
		if (dryRun) return merges(dataset, file, dryRun);
		await wait();
		throw Object.assign(new Error('Request failed with status 500'), { status: 500 });
	};

	const exportFails = async () => {
		await wait();
		throw new Error('Request failed with status 500');
	};
</script>

<Story name="Playground" args={{ open: true, onExport: exported, onImport: merges }} />

<Story
	name="As an administrator"
	args={{ open: true, admin: true, onExport: exported, onImport: merges }}
/>

<!-- On the Import tab, drop any .json file: the file is checked, its counts are shown, and
     nothing is written until "Import … into …" is pressed. -->
<Story
	name="Import, checked then imported"
	args={{ open: true, onExport: exported, onImport: merges }}
/>

<Story
	name="Import, everything already here"
	args={{ open: true, onExport: exported, onImport: alreadyHere }}
/>

<Story
	name="Import, vendor and listing missing"
	args={{ open: true, onExport: exported, onImport: missingVendor }}
/>

<Story
	name="Import, another dataset's file"
	args={{ open: true, onExport: exported, onImport: wrongKind }}
/>

<Story name="Import fails" args={{ open: true, onExport: exported, onImport: importFails }} />

<Story name="Export fails" args={{ open: true, onExport: exportFails, onImport: merges }} />
