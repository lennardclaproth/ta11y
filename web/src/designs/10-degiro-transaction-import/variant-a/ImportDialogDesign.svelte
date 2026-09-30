<script lang="ts">
	// Variant A — "Eén dialoog, twee stappen".
	// Prototype for #10: composes existing atoms/molecules only; no real service calls.
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import Dropzone from '$lib/components/molecules/dropzone/Dropzone.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import ProgressBar from '$lib/components/atoms/progress-bar/ProgressBar.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import DesignBackdrop from '../DesignBackdrop.svelte';
	import {
		accountOptions,
		countRows,
		destinationResults,
		importFileName,
		importFileSize,
		importFinishedAt,
		unlinkedProducts,
		wrongFileMessage,
		type ImportDestinationResult
	} from '../import-fixtures';

	type Phase = 'form' | 'processing' | 'result' | 'error';

	let { phase = 'form' as Phase, entry = 'Portfolio' }: { phase?: Phase; entry?: string } =
		$props();

	let accountId = $state('vnd-degiro');

	const title = $derived(phase === 'result' ? 'Import complete' : 'Import DEGIRO export');

	const statusIntent = {
		completed: 'success',
		partial: 'warning',
		failed: 'error'
	} as const;

	const statusLabel = {
		completed: 'Imported',
		partial: 'Partly imported',
		failed: 'Failed'
	} as const;

	function count(result: ImportDestinationResult, key: string): number {
		return result[key as 'total' | 'imported' | 'duplicates' | 'failed'];
	}

	// One colour class per cell: two competing text-* utilities would be resolved by
	// stylesheet order, not by the order they are listed here.
	function countTone(result: ImportDestinationResult, key: string): string {
		if (key === 'failed') return count(result, key) > 0 ? 'font-medium text-red-700' : 'text-slate-500';
		if (key === 'imported') return 'font-semibold text-slate-950';
		return 'text-slate-800';
	}
</script>

{#snippet destination(result: ImportDestinationResult)}
	<section class="border-t border-slate-400 pt-3" aria-label="{result.destination} result">
		<div class="mb-2 flex items-center justify-between gap-2">
			<Heading level="h3" size="sm" class="text-slate-900">{result.destination}</Heading>
			<Badge intent={statusIntent[result.status]} variant="soft" size="sm">
				{statusLabel[result.status]}
			</Badge>
		</div>

		<dl class="text-sm">
			{#each countRows as row (row.key)}
				<div
					class="flex items-baseline justify-between gap-3 border-b border-slate-200 py-1.5 last:border-b-0"
				>
					<dt class="text-slate-600">{row.label}</dt>
					<dd class={['tabular-nums', countTone(result, row.key)].join(' ')}>
						{count(result, row.key)}
					</dd>
				</div>
			{/each}
		</dl>

		{#if result.message}
			<Text as="p" size="sm" tone="muted" class="mt-2">{result.message}</Text>
		{/if}
	</section>
{/snippet}

<DesignBackdrop label={entry} />

<Dialog open title={title} size="xl" dismissible>
	{#if phase === 'result'}
		<div class="space-y-4">
			<Text as="p" size="sm" tone="muted">
				{importFileName} · {importFinishedAt}
			</Text>

			<div class="grid gap-4 sm:grid-cols-2">
				{#each destinationResults as result (result.destination)}
					{@render destination(result)}
				{/each}
			</div>

			<section class="border-t border-slate-400 pt-3">
				<Heading level="h3" size="sm" class="mb-1 text-slate-900">
					Products without a listing ({unlinkedProducts.length})
				</Heading>
				<Text as="p" size="sm" tone="muted" class="mb-2">
					These are imported, but they do not count towards performance yet. Add them in Listings,
					then run Rebuild portfolio.
				</Text>

				<ul class="text-sm">
					{#each unlinkedProducts as product (product.id)}
						<li
							class="flex items-baseline justify-between gap-3 border-b border-slate-200 py-1.5 last:border-b-0"
						>
							<span class="min-w-0">
								<span class="text-slate-800">{product.name}</span>
								<span class="text-slate-500">
									— {product.isin ?? product.symbol ?? 'no identifier'}
								</span>
							</span>
							<span class="shrink-0 tabular-nums text-slate-600">
								{product.transactions} tx
							</span>
						</li>
					{/each}
				</ul>
			</section>
		</div>
	{:else}
		<div class="space-y-4">
			<Text as="p" size="sm" tone="muted">
				Upload your monthly DEGIRO Account statement. One upload fills Cashflow and Portfolio, and
				rows you already imported are recognised and skipped.
			</Text>

			{#if phase === 'error'}
				<Alert intent="error" title="That file was not imported">{wrongFileMessage}</Alert>
			{/if}

			<FormField label="Account" id="imp-account" hint="The brokerage account this export belongs to">
				{#snippet children(ctx)}
					<Select
						id={ctx.id}
						bind:value={accountId}
						options={accountOptions}
						ariaLabel="Account"
						disabled={phase === 'processing'}
					/>
				{/snippet}
			</FormField>

			<FormField label="DEGIRO export" id="imp-file">
				<Dropzone
					accept=".csv"
					disabled={phase === 'processing'}
					label={phase === 'form' ? 'Drag the CSV here, or click to browse' : importFileName}
					hint={phase === 'form' ? 'One file, .csv, up to 10 MB' : importFileSize}
				/>
			</FormField>

			{#if phase === 'processing'}
				<section class="border-t border-slate-400 pt-3" aria-live="polite">
					<div class="mb-2 flex items-center justify-between gap-3">
						<Text as="span" size="sm" weight="medium">Processing…</Text>
						<Text as="span" size="sm" tone="muted">This keeps running if you close this</Text>
					</div>
					<ProgressBar value={40} ariaLabel="Import progress" />
					<ul class="mt-3 space-y-1 text-sm text-slate-600">
						<li class="flex items-center gap-2">
							<Icon icon="heroicons:check-circle" size="md" class="text-emerald-700" />
							Cashflow
						</li>
						<li class="flex items-center gap-2">
							<Icon icon="heroicons:ellipsis-horizontal-circle" size="md" class="text-slate-400" />
							Portfolio
						</li>
					</ul>
				</section>
			{:else}
				<p class="flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-slate-400 pt-3 text-sm text-slate-600">
					<span class="flex items-center gap-1.5">
						<Icon icon="heroicons:arrow-right-circle" size="md" class="text-slate-400" /> Cashflow
					</span>
					<span class="flex items-center gap-1.5">
						<Icon icon="heroicons:arrow-right-circle" size="md" class="text-slate-400" /> Portfolio
					</span>
					<span class="flex items-center gap-1.5">
						<Icon icon="heroicons:arrow-path" size="md" class="text-slate-400" /> Performance is
						recalculated
					</span>
				</p>
			{/if}
		</div>
	{/if}

	{#snippet footer()}
		{#if phase === 'result'}
			<Button variant="ghost" intent="secondary">Close</Button>
			<Button>Go to portfolio</Button>
		{:else}
			<Button variant="ghost" intent="secondary" disabled={phase === 'processing'}>Cancel</Button>
			<Button loading={phase === 'processing'} disabled={phase === 'processing'}>
				{phase === 'processing' ? 'Importing…' : 'Import'}
			</Button>
		{/if}
	{/snippet}
</Dialog>
