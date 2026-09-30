<script lang="ts">
	// Final design for #10 — variant A's two-step dialog, with variant B's activity trail
	// in place of the progress bar. The trail runs horizontal or vertical (`direction`);
	// everything else is identical, so the two orientations can be compared side by side.
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import Dropzone from '$lib/components/molecules/dropzone/Dropzone.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import DesignBackdrop from '../DesignBackdrop.svelte';
	import ImportActivityTrail from './ImportActivityTrail.svelte';
	import type { TrailDirection, TrailStep } from './activity-trail.types';
	import {
		accountOptions,
		countRows,
		destinationResults,
		importFileName,
		importFileSize,
		importFinishedAt,
		noNewRowsResults,
		unlinkedProducts,
		wrongFileMessage,
		type ImportDestinationResult
	} from '../import-fixtures';

	type Phase = 'form' | 'processing' | 'result' | 'empty' | 'error';

	let {
		phase = 'form' as Phase,
		direction = 'horizontal' as TrailDirection,
		entry = 'Portfolio'
	}: { phase?: Phase; direction?: TrailDirection; entry?: string } = $props();

	let accountId = $state('vnd-degiro');

	const isResult = $derived(phase === 'result' || phase === 'empty');
	const title = $derived(isResult ? 'Import complete' : 'Import DEGIRO export');
	const results = $derived(phase === 'empty' ? noNewRowsResults : destinationResults);
	const products = $derived(phase === 'empty' ? [] : unlinkedProducts);

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
	// stylesheet order, not by the order they are listed here. A zero is always muted —
	// emphasis belongs on counts that actually happened.
	function countTone(result: ImportDestinationResult, key: string): string {
		const value = count(result, key);
		if (value === 0) return 'text-slate-500';
		if (key === 'failed') return 'font-medium text-red-700';
		if (key === 'imported') return 'font-semibold text-slate-950';
		return 'text-slate-800';
	}

	// The trail carries the path and its status; the numbers live in the columns below it,
	// so the result never states the same count twice.
	const steps = $derived.by((): TrailStep[] => {
		if (phase === 'error') {
			return [
				{ key: 'file', title: 'File', detail: 'Not recognised', state: 'failed' },
				{ key: 'cashflow', title: 'Cashflow', detail: 'Not started', state: 'pending' },
				{ key: 'portfolio', title: 'Portfolio', detail: 'Not started', state: 'pending' },
				{ key: 'performance', title: 'Performance', detail: 'Not started', state: 'pending' }
			];
		}

		if (phase === 'processing') {
			return [
				{ key: 'file', title: 'File', detail: `${results[0].total} rows read`, state: 'done' },
				{ key: 'cashflow', title: 'Cashflow', detail: 'Imported', state: 'done' },
				{ key: 'portfolio', title: 'Portfolio', detail: 'Reading…', state: 'running' },
				{ key: 'performance', title: 'Performance', detail: 'Waiting', state: 'pending' }
			];
		}

		if (isResult) {
			return [
				{ key: 'file', title: 'File', detail: 'Accepted', state: 'done' },
				{
					key: 'cashflow',
					title: 'Cashflow',
					detail: statusLabel[results[0].status],
					state: results[0].status === 'completed' ? 'done' : 'warning'
				},
				{
					key: 'portfolio',
					title: 'Portfolio',
					detail: statusLabel[results[1].status],
					state: results[1].status === 'completed' ? 'done' : 'warning'
				},
				{ key: 'performance', title: 'Performance', detail: 'Recalculated', state: 'done' }
			];
		}

		// Before the upload the trail is only a route map: titles, no detail line to wrap.
		return [
			{ key: 'file', title: 'File', state: 'pending' },
			{ key: 'cashflow', title: 'Cashflow', state: 'pending' },
			{ key: 'portfolio', title: 'Portfolio', state: 'pending' },
			{ key: 'performance', title: 'Performance', state: 'pending' }
		];
	});

	const trailCaption = $derived(
		phase === 'processing'
			? 'This keeps running if you close this'
			: isResult
				? importFinishedAt
				: phase === 'error'
					? 'Nothing was changed'
					: 'One upload, in this order'
	);
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

{#snippet trail()}
	<section
		class="border-t border-slate-400 pt-3"
		aria-live="polite"
		aria-label="Import activity"
		aria-busy={phase === 'processing'}
	>
		<div class="mb-3 flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
			<Text as="span" size="sm" weight="medium" class="text-slate-900">
				{phase === 'processing' ? 'Importing…' : 'Activity'}
			</Text>
			<Text as="span" size="sm" tone="muted">{trailCaption}</Text>
		</div>

		<ImportActivityTrail {steps} {direction} />
	</section>
{/snippet}

<DesignBackdrop label={entry} />

<Dialog open {title} size="xl" dismissible>
	{#if isResult}
		<div class="space-y-4">
			<Text as="p" size="sm" tone="muted">
				{importFileName} · {accountOptions[0].label}
			</Text>

			{@render trail()}

			{#if phase === 'empty'}
				<Alert intent="info" title="No new transactions">
					Every row in this export was already imported earlier. Nothing was added, and nothing was
					duplicated.
				</Alert>
			{/if}

			<div class="grid gap-4 sm:grid-cols-2">
				{#each results as result (result.destination)}
					{@render destination(result)}
				{/each}
			</div>

			{#if products.length > 0}
				<section class="border-t border-slate-400 pt-3">
					<Heading level="h3" size="sm" class="mb-1 text-slate-900">
						Products without a listing ({products.length})
					</Heading>
					<Text as="p" size="sm" tone="muted" class="mb-2">
						These are imported, but they do not count towards performance yet. Add them in Listings,
						then run Rebuild portfolio.
					</Text>

					<ul class="text-sm">
						{#each products as product (product.id)}
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
			{/if}
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

			{@render trail()}
		</div>
	{/if}

	{#snippet footer()}
		{#if isResult}
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
