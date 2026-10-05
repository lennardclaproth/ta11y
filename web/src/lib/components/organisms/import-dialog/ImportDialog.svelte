<script lang="ts">
	// One upload, two destinations. The dialog posts the same DEGIRO export to the
	// cashflow and the portfolio import, polls both until they finish, and then shows
	// what landed where. The two halves are independent by design: one can succeed
	// while the other is refused, so the result reports them side by side rather than
	// collapsing them into a single verdict.
	import { ApiError } from '$lib/api/client';
	import type { ImportResult, Vendor } from '$lib/api/types';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import Dropzone from '$lib/components/molecules/dropzone/Dropzone.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import ActivityTrail from '$lib/components/molecules/activity-trail/ActivityTrail.svelte';
	import type { TrailState, TrailStep } from '$lib/components/molecules/activity-trail/activity-trail.types';
	import { getImport, importCashflow, importPortfolio } from '$lib/services/importer';
	import {
		destinationStatusLabels,
		importCountRows,
		wrongFileMessage,
		type DestinationOutcome,
		type DestinationStatus,
		type ImportCountKey,
		type ImportPhase
	} from './import-dialog.types';
	import { countToneClasses, destinationBadgeIntents } from './import-dialog.variants';

	type Props = {
		/** Two-way bindable open state. */
		open?: boolean;
		/** Brokerage vendors the export can belong to. */
		vendors?: Vendor[];
		/** Called once both destinations have finished, so the caller can refresh. */
		onFinished?: () => void;
		/** Called when the user leaves for the portfolio from the result. */
		onGoToPortfolio?: () => void;
	};

	let { open = $bindable(false), vendors = [], onFinished, onGoToPortfolio }: Props = $props();

	/** How often an import in flight is re-read. Processing is detached, so it only arrives by asking. */
	const POLL_MS = 1000;
	/** Stop polling rather than spin forever if an import never reaches a terminal state. */
	const POLL_TIMEOUT_MS = 120_000;

	let phase = $state<ImportPhase>('form');
	let vendorId = $state('');
	let files = $state<FileList | null>(null);
	let failure = $state('');
	let cashflowResult = $state<ImportResult | null>(null);
	let portfolioResult = $state<ImportResult | null>(null);

	const file = $derived(files?.[0] ?? null);
	const vendorOptions = $derived(vendors.map((vendor) => ({ value: vendor.id, label: vendor.name })));
	const selectedVendor = $derived(vendors.find((vendor) => vendor.id === vendorId) ?? null);

	// The first brokerage vendor is preselected so the common case is one file away,
	// but the choice stays visible: the account an export belongs to is the one thing
	// the file itself does not say.
	$effect(() => {
		if (vendorId === '' && vendors.length > 0) vendorId = vendors[0].id;
	});

	const outcomes = $derived<DestinationOutcome[]>([
		{ destination: 'Cashflow', result: cashflowResult },
		{ destination: 'Portfolio', result: portfolioResult }
	]);

	const anyImported = $derived(
		(cashflowResult?.imported ?? 0) + (portfolioResult?.imported ?? 0) > 0
	);
	const everythingKnown = $derived(
		phase === 'result' &&
			!anyImported &&
			outcomes.every((outcome) => outcome.result?.status === 'completed')
	);

	function destinationStatus(result: ImportResult | null): DestinationStatus {
		if (!result || result.status === 'failed') return 'failed';
		return result.failed > 0 ? 'partial' : 'completed';
	}

	function count(result: ImportResult | null, key: ImportCountKey): number {
		return result ? result[key] : 0;
	}

	function countTone(result: ImportResult | null, key: ImportCountKey): string {
		const value = count(result, key);
		if (value === 0) return countToneClasses.zero;
		if (key === 'failed') return countToneClasses.failed;
		if (key === 'imported') return countToneClasses.imported;
		return countToneClasses.plain;
	}

	/**
	 * What a destination has to say beyond its counts — only when it did not fully
	 * succeed. The classified `reason` is all the server returns about a failure; its
	 * own error message stays in the server log.
	 */
	function destinationMessage(result: ImportResult | null): string {
		if (!result) return 'This destination did not report back.';
		if (result.status === 'failed') {
			return result.reason === 'file_not_recognised'
				? wrongFileMessage
				: 'This destination could not process the file. Nothing was added here.';
		}
		if (result.failed > 0) {
			return `${result.failed} rows could not be read and were skipped.`;
		}
		return '';
	}

	const unlinkedProducts = $derived(portfolioResult?.unlinked_products ?? []);

	function trailState(result: ImportResult | null): TrailState {
		if (phase === 'form') return 'pending';
		if (!result) return 'running';
		if (result.status === 'failed') return 'failed';
		if (result.status !== 'completed') return 'running';
		return result.failed > 0 ? 'warning' : 'done';
	}

	function trailDetail(result: ImportResult | null): string | undefined {
		if (phase === 'form') return undefined;
		if (!result) return 'Waiting';
		if (result.status === 'failed') return 'Not recognised';
		if (result.status !== 'completed') return 'Reading…';
		return destinationStatusLabels[destinationStatus(result)];
	}

	const fileStepState = $derived<TrailState>(
		phase === 'form'
			? 'pending'
			: outcomes.every((outcome) => outcome.result?.status === 'failed')
				? 'failed'
				: 'done'
	);

	// The rebuild is kicked off by the completed portfolio import and runs on its own,
	// so the trail says it started rather than claiming a finish it cannot observe.
	const performanceState = $derived<TrailState>(
		phase !== 'result'
			? 'pending'
			: portfolioResult?.status === 'completed'
				? 'done'
				: 'pending'
	);

	const steps = $derived<TrailStep[]>([
		{
			key: 'file',
			title: 'File',
			detail: phase === 'form' ? undefined : fileStepState === 'failed' ? 'Not recognised' : 'Accepted',
			state: fileStepState
		},
		{
			key: 'cashflow',
			title: 'Cashflow',
			detail: trailDetail(cashflowResult),
			state: trailState(cashflowResult)
		},
		{
			key: 'portfolio',
			title: 'Portfolio',
			detail: trailDetail(portfolioResult),
			state: trailState(portfolioResult)
		},
		{
			key: 'performance',
			title: 'Performance',
			detail:
				phase === 'form'
					? undefined
					: performanceState === 'done'
						? 'Rebuild started'
						: 'Waiting',
			state: performanceState
		}
	]);

	const trailCaption = $derived(
		phase === 'processing'
			? 'This keeps running if you close this'
			: phase === 'result'
				? 'Finished'
				: failure
					? 'Nothing was changed'
					: 'One upload, in this order'
	);

	const title = $derived(phase === 'result' ? 'Import complete' : 'Import a brokerage export');

	/**
	 * Identifies the upload in flight. Closing the dialog abandons it — processing keeps
	 * running on the server, but a run that finishes after the user walked away must not
	 * reopen the dialog on a result they did not ask for.
	 */
	let runId = $state(0);

	function reset() {
		phase = 'form';
		files = null;
		failure = '';
		cashflowResult = null;
		portfolioResult = null;
		runId += 1;
	}

	function close() {
		open = false;
		reset();
	}

	const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

	/** Polls one import until it reaches a terminal state, or gives up after the timeout. */
	async function poll(importId: string): Promise<ImportResult | null> {
		const deadline = Date.now() + POLL_TIMEOUT_MS;
		while (Date.now() < deadline) {
			const result = await getImport(importId);
			if (result.status === 'completed' || result.status === 'failed') return result;
			await sleep(POLL_MS);
		}
		return null;
	}

	/**
	 * Uploads to one destination and follows it to the end. A destination that never
	 * gets off the ground resolves to null rather than throwing, so the other half
	 * still gets reported — partial success is the expected outcome here, not an error.
	 */
	async function run(
		upload: () => Promise<{ import_id: string }>,
		assign: (result: ImportResult | null) => void
	): Promise<void> {
		try {
			const accepted = await upload();
			assign(await poll(accepted.import_id));
		} catch (cause) {
			assign(null);
			if (!failure) failure = describe(cause);
		}
	}

	function describe(cause: unknown): string {
		if (!(cause instanceof ApiError)) {
			return 'Could not confirm the upload was accepted. Check your connection and try again.';
		}
		if (cause.status === 404) return 'That account no longer exists. Close this and try again.';
		if (cause.status === 422) {
			return 'This vendor does not accept imports. Pick another brokerage account.';
		}
		if (cause.status === 400) return 'The file was rejected. Upload the CSV DEGIRO exported.';
		return 'The upload could not be accepted. Try again.';
	}

	async function submit() {
		if (phase === 'processing' || !file || !vendorId) return;
		failure = '';
		cashflowResult = null;
		portfolioResult = null;
		phase = 'processing';

		const current = file;
		const run_ = runId;
		await Promise.all([
			run(
				() => importCashflow({ file: current, vendor_id: vendorId }),
				(result) => (cashflowResult = result)
			),
			run(
				() => importPortfolio({ file: current, vendor_id: vendorId }),
				(result) => (portfolioResult = result)
			)
		]);

		if (run_ !== runId) return;

		// Two outcomes have nothing to report and belong on the form, where they can be
		// acted on: a file neither destination recognised is the wrong export, and an
		// upload neither destination accepted never became an import at all.
		if (outcomes.every((outcome) => outcome.result?.reason === 'file_not_recognised')) {
			failure = wrongFileMessage;
			phase = 'form';
			return;
		}
		if (outcomes.every((outcome) => outcome.result === null)) {
			if (!failure) failure = 'The upload could not be accepted. Try again.';
			phase = 'form';
			return;
		}

		phase = 'result';
		onFinished?.();
	}
</script>

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

		<ActivityTrail {steps} direction="horizontal" ariaLabel="Import activity" />
	</section>
{/snippet}

{#snippet destination(outcome: DestinationOutcome)}
	{@const status = destinationStatus(outcome.result)}
	{@const message = destinationMessage(outcome.result)}
	<section class="border-t border-slate-400 pt-3" aria-label="{outcome.destination} result">
		<div class="mb-2 flex items-center justify-between gap-2">
			<Heading level="h3" size="sm" class="text-slate-900">{outcome.destination}</Heading>
			<Badge intent={destinationBadgeIntents[status]} variant="soft" size="sm">
				{destinationStatusLabels[status]}
			</Badge>
		</div>

		<dl class="text-sm">
			{#each importCountRows as row (row.key)}
				<div
					class="flex items-baseline justify-between gap-3 border-b border-slate-200 py-1.5 last:border-b-0"
				>
					<dt class="text-slate-600">{row.label}</dt>
					<dd class={['tabular-nums', countTone(outcome.result, row.key)].join(' ')}>
						{count(outcome.result, row.key)}
					</dd>
				</div>
			{/each}
		</dl>

		{#if message}
			<Text as="p" size="sm" tone="muted" class="mt-2">{message}</Text>
		{/if}
	</section>
{/snippet}

<Dialog bind:open title={title} size="xl" closeOnBackdrop={phase !== 'processing'} onClose={reset}>
	{#if phase === 'result'}
		<div class="space-y-4">
			<Text as="p" size="sm" tone="muted">
				{file?.name ?? 'Uploaded export'}{selectedVendor ? ` · ${selectedVendor.name}` : ''}
			</Text>

			{@render trail()}

			{#if everythingKnown}
				<Alert intent="info" title="No new transactions">
					Every row in this export was already imported earlier. Nothing was added, and nothing was
					duplicated.
				</Alert>
			{/if}

			<div class="grid gap-4 sm:grid-cols-2">
				{#each outcomes as outcome (outcome.destination)}
					{@render destination(outcome)}
				{/each}
			</div>

			{#if unlinkedProducts.length > 0}
				<section class="border-t border-slate-400 pt-3">
					<Heading level="h3" size="sm" class="mb-1 text-slate-900">
						Products without a listing ({unlinkedProducts.length})
					</Heading>
					<Text as="p" size="sm" tone="muted" class="mb-2">
						These are imported, but they do not count towards performance yet. Add them in Listings,
						then run Rebuild portfolio.
					</Text>

					<ul class="text-sm">
						{#each unlinkedProducts as product (product.isin ?? product.symbol ?? product.name)}
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
		<div class="space-y-4" aria-busy={phase === 'processing'}>
			<Text as="p" size="sm" tone="muted">
				Upload your monthly DEGIRO Account statement. One upload fills Cashflow and Portfolio, and
				rows you already imported are recognised and skipped.
			</Text>

			{#if failure}
				<Alert intent="error" title="That file was not imported">{failure}</Alert>
			{/if}

			{#if vendorOptions.length === 0}
				<Alert intent="warning" title="No brokerage account">
					Add an active brokerage vendor before importing a broker export.
				</Alert>
			{/if}

			<FormField label="Account" id="imp-account" hint="The brokerage account this export belongs to">
				{#snippet children(ctx)}
					<Select
						id={ctx.id}
						bind:value={vendorId}
						options={vendorOptions}
						placeholder="Select an account"
						ariaLabel="Account"
						disabled={phase === 'processing'}
					/>
				{/snippet}
			</FormField>

			<FormField label="Brokerage export" id="imp-file">
				{#snippet children(ctx)}
					<Dropzone
						id={ctx.id}
						bind:files
						accept=".csv"
						disabled={phase === 'processing'}
						label={file ? file.name : 'Drag the CSV here, or click to browse'}
						hint={file ? `${(file.size / 1024).toFixed(1)} KB` : 'One file, .csv, up to 10 MB'}
					/>
				{/snippet}
			</FormField>

			{@render trail()}
		</div>
	{/if}

	{#snippet footer()}
		{#if phase === 'result'}
			<Button variant="ghost" intent="secondary" onclick={close}>Close</Button>
			<Button
				onclick={() => {
					close();
					onGoToPortfolio?.();
				}}
			>
				Go to portfolio
			</Button>
		{:else}
			<Button variant="ghost" intent="secondary" disabled={phase === 'processing'} onclick={close}>
				Cancel
			</Button>
			<Button
				onclick={() => void submit()}
				loading={phase === 'processing'}
				disabled={phase === 'processing' || !file || !vendorId}
			>
				{phase === 'processing' ? 'Importing…' : 'Import'}
			</Button>
		{/if}
	{/snippet}
</Dialog>
