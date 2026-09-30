<script lang="ts">
	// Variant C — "Importstrook op de pagina".
	// Prototype for #10: no overlay. The import opens as a ruled section above the ledger
	// and turns into a full-width receipt that stays until it is dismissed.
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Dropzone from '$lib/components/molecules/dropzone/Dropzone.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import {
		accountOptions,
		countRows,
		destinationResults,
		importFileName,
		importFileSize,
		importFinishedAt,
		unlinkedProducts,
		wrongFileMessage,
		type ImportDestinationResult,
		type UnlinkedProduct
	} from '../import-fixtures';

	type Phase = 'form' | 'processing' | 'result' | 'error';

	let { phase = 'form' as Phase }: { phase?: Phase } = $props();

	let accountId = $state('vnd-degiro');

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

{#snippet isinCell(row: UnlinkedProduct)}
	<span class="tabular-nums text-slate-700">{row.isin ?? '—'}</span>
{/snippet}

{#snippet actionCell(row: UnlinkedProduct)}
	<Button size="sm" variant="outline" intent="secondary" ariaLabel="Add listing for {row.name}">
		Add listing
	</Button>
{/snippet}

<div class="min-h-[40rem] bg-taupe-100 px-4 py-4 lg:px-8">
	<header class="mb-4 border-b-2 border-slate-900 pb-2">
		<Heading level="h1" size="xl" class="text-slate-900">Portfolio</Heading>
	</header>

	<!-- The import section: ruled, in normal document flow, above the ledger. -->
	<section class="mb-5 border-t-2 border-slate-900 bg-taupe-50 px-4 py-4" aria-label="Import">
		<div class="mb-3 flex items-start justify-between gap-3">
			<div class="min-w-0">
				<Heading level="h2" size="md" class="text-slate-900">
					{phase === 'result' ? 'Import complete' : 'Import DEGIRO export'}
				</Heading>
				<Text as="p" size="sm" tone="muted">
					{#if phase === 'result'}
						{importFileName} · {importFinishedAt}
					{:else}
						One upload fills Cashflow and Portfolio. Rows you already imported are skipped.
					{/if}
				</Text>
			</div>
			<IconButton
				icon="heroicons:x-mark"
				ariaLabel="Close import"
				size="sm"
				variant="ghost"
				intent="secondary"
			/>
		</div>

		{#if phase === 'result'}
			<!-- The receipt: one row per figure, one column per destination. -->
			<div class="max-w-2xl overflow-x-auto">
				<table class="w-full min-w-[26rem] border-collapse text-sm">
					<thead>
						<tr class="border-b border-slate-400">
							<th scope="col" class="py-2 pr-3 text-left text-xs font-semibold text-slate-700">
								Rows
							</th>
							{#each destinationResults as result (result.destination)}
								<th scope="col" class="py-2 pl-3 text-right text-xs font-semibold text-slate-700">
									<span class="flex items-center justify-end gap-2">
										{result.destination}
										<Badge intent={statusIntent[result.status]} variant="soft" size="sm">
											{statusLabel[result.status]}
										</Badge>
									</span>
								</th>
							{/each}
						</tr>
					</thead>
					<tbody>
						{#each countRows as row (row.key)}
							<tr class="border-b border-slate-200">
								<th scope="row" class="py-2 pr-3 text-left font-normal text-slate-600">
									{row.label}
								</th>
								{#each destinationResults as result (result.destination)}
									<td
										class={[
											'py-2 pl-3 text-right tabular-nums',
											countTone(result, row.key)
										].join(' ')}
									>
										{count(result, row.key)}
									</td>
								{/each}
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			{#each destinationResults as result (result.destination)}
				{#if result.message}
					<Text as="p" size="sm" tone="muted" class="mt-2">
						{result.destination}: {result.message}
					</Text>
				{/if}
			{/each}

			<div class="mt-5 border-t border-slate-400 pt-3">
				<Heading level="h3" size="sm" class="text-slate-900">
					Products without a listing ({unlinkedProducts.length})
				</Heading>
				<Text as="p" size="sm" tone="muted" class="mb-2">
					These are imported, but they do not count towards performance until a listing with the
					same ISIN or symbol exists. Add one, then run Rebuild portfolio.
				</Text>

				<DataTable
					rows={unlinkedProducts}
					emptyText="Every product matched a listing"
					class="max-h-64"
					columns={[
						{ key: 'name', header: 'Product', value: (r: UnlinkedProduct) => r.name },
						{ key: 'isin', header: 'ISIN', cell: isinCell },
						{ key: 'symbol', header: 'Symbol', value: (r: UnlinkedProduct) => r.symbol ?? '—' },
						{
							key: 'tx',
							header: 'Transactions',
							align: 'right',
							value: (r: UnlinkedProduct) => r.transactions
						},
						{ key: 'action', header: '', align: 'right', cell: actionCell }
					]}
				/>
			</div>

			<div class="mt-4 flex flex-wrap items-center justify-end gap-2 border-t border-slate-400 pt-3">
				<Button variant="ghost" intent="secondary">Import another file</Button>
				<Button>Go to portfolio</Button>
			</div>
		{:else}
			{#if phase === 'error'}
				<Alert intent="error" title="That file was not imported" class="mb-3">
					{wrongFileMessage}
				</Alert>
			{/if}

			<div class="flex max-w-4xl flex-col gap-3 lg:flex-row lg:items-end">
				<FormField label="DEGIRO export" id="imp-c-file" class="min-w-0 flex-1">
					<Dropzone
						accept=".csv"
						disabled={phase === 'processing'}
						label={phase === 'processing'
							? importFileName
							: 'Drag the CSV here, or click to browse'}
						hint={phase === 'processing' ? importFileSize : 'One file, .csv, up to 10 MB'}
						class="p-4"
					/>
				</FormField>

				<FormField
					label="Account"
					id="imp-c-account"
					hint="Brokerage account"
					class="lg:w-64 lg:shrink-0"
				>
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

				<Button
					class="lg:mb-6 lg:shrink-0"
					loading={phase === 'processing'}
					disabled={phase === 'processing'}
				>
					{phase === 'processing' ? 'Importing…' : 'Import'}
				</Button>
			</div>

			{#if phase === 'processing'}
				<p class="mt-3 border-t border-slate-400 pt-3 text-sm text-slate-600" aria-live="polite">
					Processing. Cashflow and Portfolio report here when they are done; this keeps running if
					you leave the page.
				</p>
			{/if}
		{/if}
	</section>

	<!-- Stand-in for the ledger the section sits above. -->
	<div class="border-t border-slate-400 bg-taupe-50 px-3 py-6">
		<Text as="p" size="sm" tone="muted">Positions</Text>
		<div class="mt-3 space-y-2">
			{#each [0, 1, 2, 3] as row (row)}
				<div class="h-6 border-b border-slate-200"></div>
			{/each}
		</div>
	</div>
</div>
