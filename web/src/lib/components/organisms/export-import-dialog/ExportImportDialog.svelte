<script lang="ts">
	// Export and import, one dataset at a time. Both tabs share one selector, so every action
	// says exactly what it does — "Export cashflow", "Import 1,429 into Cashflow". An import is
	// checked first: the file is merged in a dry run and its counts are read before anything is
	// written. A file naming vendors or listings this ta11y lacks is refused with those named.
	//
	// The dialog is presentational: the caller downloads and uploads. A dialog that stays open
	// reports its own outcomes in a band inside it, never as a toast behind its scrim.
	import { untrack } from 'svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import Dropzone from '$lib/components/molecules/dropzone/Dropzone.svelte';
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Radio from '$lib/components/atoms/radio/Radio.svelte';
	import Spinner from '$lib/components/atoms/spinner/Spinner.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import type { BackupDataset, BackupResult } from '$lib/api/types';
	import { listingSources } from '$lib/components/organisms/listing-form/listing-form.types';
	import { countLines, datasetOptions, type TransferPhase } from './export-import-dialog.types';

	type Props = {
		open?: boolean;
		/** Administrators also move the shared end-of-day prices. */
		admin?: boolean;
		/** Downloads the dataset's file; a thrown error is reported in the dialog. */
		onExport: (dataset: BackupDataset) => Promise<void>;
		/**
		 * Merges the file into the dataset, or on a dry run reports what it would do. A file
		 * naming what this ta11y lacks resolves with `missing` filled; a file the server will
		 * not read throws.
		 */
		onImport: (dataset: BackupDataset, file: File, dryRun: boolean) => Promise<BackupResult>;
	};

	let { open = $bindable(false), admin = false, onExport, onImport }: Props = $props();

	let tab = $state('export');
	let dataset = $state<BackupDataset>('cashflow');
	let files = $state<FileList | null>(null);
	let file = $state<File | null>(null);
	let phase = $state<TransferPhase>('idle');
	let result = $state<BackupResult | null>(null);
	let exporting = $state(false);
	let exported = $state<BackupDataset | null>(null);
	let failure = $state<string | null>(null);

	// Every open starts clean: a half-finished import from last time is not this one.
	$effect(() => {
		if (open) {
			untrack(() => {
				tab = 'export';
				dataset = 'cashflow';
				resetImport();
				exported = null;
			});
		}
	});

	const options = $derived(datasetOptions.filter((o) => admin || !o.adminOnly));
	const selected = $derived(options.find((o) => o.value === dataset) ?? options[0]);
	const lines = $derived(result ? countLines(dataset, result) : []);
	const newTotal = $derived(lines.reduce((n, l) => n + l.count.new, 0));
	const existingTotal = $derived(lines.reduce((n, l) => n + l.count.existing, 0));
	const missingNames = $derived(
		result
			? [
					...result.missing.vendors.map((v) => `${v} (bank or broker)`),
					...result.missing.listings.map(
						(l) =>
							`${l.symbol} (${listingSources.find((s) => s.value === l.source)?.label ?? l.source} listing)`
					)
				]
			: []
	);
	// Once a file is chosen the dataset is settled; switching it would leave the file behind.
	// Only the chosen one stays, which keeps the counts in view rather than below the list.
	const locked = $derived(tab === 'import' && file !== null);
	const shown = $derived(locked ? options.filter((o) => o.value === dataset) : options);
	const busy = $derived(phase === 'checking' || phase === 'importing' || exporting);
	// The other tab stays shut while something runs, so its outcome lands where it started.
	const tabs = $derived([
		{ value: 'export', label: 'Export', disabled: busy && tab !== 'export' },
		{ value: 'import', label: 'Import', disabled: busy && tab !== 'import' }
	]);

	const fmt = (n: number) => n.toLocaleString('en');

	function resetImport() {
		files = null;
		file = null;
		phase = 'idle';
		result = null;
		failure = null;
	}

	// A dropped or chosen file is checked straight away; nothing is written until confirmed.
	$effect(() => {
		const chosen = files?.[0] ?? null;
		if (chosen && chosen !== untrack(() => file)) {
			untrack(() => {
				file = chosen;
				void check();
			});
		}
	});

	function refusal(error: unknown): string {
		const body = error && typeof error === 'object' && 'body' in error ? error.body : null;
		const reason =
			body && typeof body === 'object' ? (body as Record<string, unknown>).file : undefined;
		if (typeof reason === 'string' && reason.includes('wrong kind')) {
			return `This isn't a ${selected.title.toLowerCase()} export. Choose a file made with Export ${selected.title.toLowerCase()}.`;
		}
		return 'This file can’t be read as a ta11y export. Choose a file made with Export.';
	}

	async function check() {
		if (!file) return;
		phase = 'checking';
		failure = null;
		try {
			result = await onImport(dataset, file, true);
			phase = missingNames.length > 0 ? 'missing' : 'preview';
		} catch (error) {
			const status = error && typeof error === 'object' && 'status' in error ? error.status : 0;
			if (status === 422 || status === 400) {
				failure = refusal(error);
				phase = 'refused';
			} else {
				failure = 'The file could not be checked. Try again.';
				phase = 'idle';
				file = null;
				files = null;
			}
		}
	}

	async function runImport() {
		if (!file) return;
		phase = 'importing';
		failure = null;
		try {
			result = await onImport(dataset, file, false);
			phase = missingNames.length > 0 ? 'missing' : 'done';
		} catch {
			// The server merges in one transaction, so a failed import wrote nothing.
			failure = 'The import failed and nothing was written. Try again.';
			phase = 'preview';
		}
	}

	async function runExport() {
		exporting = true;
		failure = null;
		exported = null;
		try {
			await onExport(dataset);
			exported = dataset;
		} catch {
			failure = `The ${selected.title.toLowerCase()} export failed. Try again.`;
		} finally {
			exporting = false;
		}
	}

	function selectTab(value: string) {
		failure = null;
		exported = null;
		if (value === 'export') resetImport();
	}
</script>

{#snippet counts()}
	<table class="w-full border-collapse text-sm">
		<thead>
			<tr class="border-b border-slate-400 text-left text-xs text-slate-500">
				<th class="py-1.5 font-normal">Record</th>
				<th class="w-24 py-1.5 text-right font-normal">New</th>
				<th class="w-28 py-1.5 text-right font-normal">Already here</th>
			</tr>
		</thead>
		<tbody>
			{#each lines as line (line.key)}
				<tr class="border-b border-slate-200 last:border-b-0">
					<td class="py-1.5 text-slate-700">{line.label}</td>
					<td class="py-1.5 text-right text-slate-900 tabular-nums">{fmt(line.count.new)}</td>
					<td class="py-1.5 text-right text-slate-500 tabular-nums">
						{fmt(line.count.existing)}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/snippet}

<!-- While a file is being checked or imported every way out is inert: closing would hide the
     outcome of a write that is still on its way. -->
<Dialog
	bind:open
	size="lg"
	title="Export and import data"
	closeOnBackdrop={!busy}
	closeOnEscape={!busy}
	dismissible={!busy}
>
	<div class="space-y-4">
		<Tabs {tabs} bind:value={tab} size="sm" ariaLabel="Export or import" onChange={selectTab} />

		{#if failure && phase !== 'refused'}
			<NoticeBand intent="error" surface="inset" gutter="dialog" class="-mx-5">{failure}</NoticeBand
			>
		{/if}

		<fieldset>
			<legend class="mb-1 text-xs text-slate-500">
				{tab === 'export' ? 'What to export' : 'What to import'}
			</legend>
			<div class="divide-y divide-slate-200 border-y border-slate-200">
				{#each shown as option (option.value)}
					<label
						class={[
							'flex items-start gap-3 py-2.5',
							locked ? 'cursor-default' : 'cursor-pointer'
						].join(' ')}
					>
						<Radio
							name="export-import-dataset"
							value={option.value}
							bind:group={dataset}
							disabled={locked || busy}
							class="mt-0.5"
						/>
						<span class="min-w-0 flex-1">
							<span class="flex items-center gap-2">
								<span class="font-heading text-base text-slate-900">{option.title}</span>
								{#if option.adminOnly}
									<Badge intent="neutral" variant="soft" size="sm">Admin</Badge>
								{/if}
							</span>
							<span class="block text-sm text-slate-600">{option.holds}</span>
						</span>
					</label>
				{/each}
			</div>
		</fieldset>

		{#if tab === 'export'}
			{#if exported === dataset}
				<NoticeBand intent="success" surface="inset" gutter="dialog" class="-mx-5">
					{selected.title} exported. Your browser saved the file.
				</NoticeBand>
			{/if}
		{:else if phase === 'idle'}
			<Dropzone
				accept=".json,application/json"
				bind:files
				label={`Drop a ${selected.title.toLowerCase()} export here, or click to choose one`}
				hint="What is already here is left alone. Nothing is written until you confirm."
			/>
		{:else}
			<div class="flex flex-wrap items-baseline justify-between gap-2">
				<p class="flex min-w-0 items-baseline gap-2 text-sm">
					<Icon
						icon="heroicons:document-text"
						size="sm"
						class="shrink-0 self-center text-slate-500"
					/>
					<span class="truncate text-slate-900">{file?.name}</span>
				</p>
				{#if phase !== 'checking' && phase !== 'importing' && phase !== 'done'}
					<Button variant="ruled" size="sm" intent="secondary" onclick={resetImport}>
						Choose another file
					</Button>
				{/if}
			</div>

			{#if phase === 'refused'}
				<NoticeBand
					intent="error"
					surface="inset"
					gutter="dialog"
					title="This file wasn't imported"
					class="-mx-5"
				>
					{failure}
				</NoticeBand>
			{:else if phase === 'missing'}
				<NoticeBand
					intent="error"
					surface="inset"
					gutter="dialog"
					title="This file can't be imported here yet"
					class="-mx-5"
				>
					It names {missingNames.join(' and ')}, which this ta11y doesn't have. Nothing was written.
					Once an administrator adds {missingNames.length === 1 ? 'it' : 'them'}, check the file
					again.
				</NoticeBand>
			{:else if phase === 'done'}
				<NoticeBand
					intent="success"
					surface="inset"
					gutter="dialog"
					title="Imported {fmt(newTotal)} records into {selected.title}"
					class="-mx-5"
				>
					{selected.rebuilds ?? 'What was already here was left alone.'}
				</NoticeBand>
				{@render counts()}
			{:else if phase === 'checking'}
				<Text as="p" size="sm" tone="muted" class="flex items-center gap-2">
					<Spinner size="sm" label="Checking the file" />Checking it against what is here…
				</Text>
			{:else}
				{#if newTotal === 0}
					<NoticeBand
						intent="info"
						surface="inset"
						gutter="dialog"
						title="Everything in this file is already here"
						class="-mx-5"
					>
						Importing it would add nothing.
					</NoticeBand>
				{:else}
					<Text as="p" size="sm" class="text-slate-800">
						Checked, nothing written yet: <span class="tabular-nums">{fmt(newTotal)}</span> new,
						<span class="tabular-nums">{fmt(existingTotal)}</span> already here.
					</Text>
				{/if}
				{@render counts()}
				{#if selected.rebuilds}
					<Text as="p" size="xs" tone="muted">{selected.rebuilds}</Text>
				{/if}
			{/if}
		{/if}
	</div>

	{#snippet footer()}
		<div class="flex flex-wrap justify-end gap-3">
			{#if tab === 'export'}
				<Button variant="ghost" intent="secondary" shape="default" onclick={() => (open = false)}>
					Close
				</Button>
				<Button shape="default" loading={exporting} onclick={() => void runExport()}>
					<Icon icon="heroicons:arrow-down-tray" />Export {selected.title.toLowerCase()}
				</Button>
			{:else if phase === 'done'}
				<Button shape="default" onclick={() => (open = false)}>Done</Button>
			{:else if phase === 'missing'}
				<Button variant="ghost" intent="secondary" shape="default" onclick={() => (open = false)}>
					Close
				</Button>
				<Button variant="outline" intent="secondary" shape="default" onclick={() => void check()}>
					Check again
				</Button>
			{:else}
				<Button
					variant="ghost"
					intent="secondary"
					shape="default"
					disabled={phase === 'importing'}
					onclick={() => (open = false)}
				>
					Cancel
				</Button>
				<Button
					shape="default"
					loading={phase === 'importing'}
					disabled={(phase !== 'preview' && phase !== 'importing') || newTotal === 0}
					onclick={() => void runImport()}
				>
					{phase === 'preview' || phase === 'importing'
						? newTotal === 0
							? 'Nothing to import'
							: `Import ${fmt(newTotal)} into ${selected.title}`
						: `Import into ${selected.title}`}
				</Button>
			{/if}
		</div>
	{/snippet}
</Dialog>
