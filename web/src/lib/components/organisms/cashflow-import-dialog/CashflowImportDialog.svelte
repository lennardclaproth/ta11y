<script lang="ts">
	// A bank statement fills one destination: Cashflow. The dialog posts the export to the
	// cashflow import, polls it until it finishes and reports what landed. Matching the new
	// rows to confirmed recurring items runs on the server once the import completes, so the
	// trail says it started rather than claiming a finish it cannot observe.
	import { ApiError } from '$lib/api/client';
	import type { ImportResult, Vendor } from '$lib/api/types';
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import Dropzone from '$lib/components/molecules/dropzone/Dropzone.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import ActivityTrail from '$lib/components/molecules/activity-trail/ActivityTrail.svelte';
	import type {
		TrailState,
		TrailStep
	} from '$lib/components/molecules/activity-trail/activity-trail.types';
	import { getImport, importCashflow } from '$lib/services/importer';
	import {
		destinationStatusLabels,
		importCountRows,
		type DestinationStatus,
		type ImportCountKey,
		type ImportPhase
	} from '$lib/components/organisms/import-dialog/import-dialog.types';
	import {
		countToneClasses,
		destinationBadgeIntents
	} from '$lib/components/organisms/import-dialog/import-dialog.variants';
	import { wrongBankFileMessage } from './cashflow-import-dialog.types';

	type Props = {
		/** Two-way bindable open state. */
		open?: boolean;
		/** Bank vendors the statement can belong to. */
		vendors?: Vendor[];
		/** Called once the import has finished, so the caller can refresh. */
		onFinished?: () => void;
		/** Called when the user switches to the brokerage import for a DEGIRO statement. */
		onBrokerageImport?: () => void;
	};

	let { open = $bindable(false), vendors = [], onFinished, onBrokerageImport }: Props = $props();

	/** How often an import in flight is re-read. Processing is detached, so it only arrives by asking. */
	const POLL_MS = 1000;
	/** Stop polling rather than spin forever if an import never reaches a terminal state. */
	const POLL_TIMEOUT_MS = 120_000;

	let phase = $state<ImportPhase>('form');
	let vendorId = $state('');
	let files = $state<FileList | null>(null);
	let failure = $state('');
	let result = $state<ImportResult | null>(null);

	const file = $derived(files?.[0] ?? null);
	const vendorOptions = $derived(
		vendors.map((vendor) => ({ value: vendor.id, label: vendor.name }))
	);
	const selectedVendor = $derived(vendors.find((vendor) => vendor.id === vendorId) ?? null);
	const bankName = $derived(selectedVendor?.name ?? 'your bank');

	// The first bank is preselected so the common case is one file away, but the choice
	// stays visible: the account a statement belongs to is the one thing the file does not say.
	// A selection that is no longer offered falls back the same way, so it is never posted.
	$effect(() => {
		if (!vendors.some((vendor) => vendor.id === vendorId)) vendorId = vendors[0]?.id ?? '';
	});

	const status = $derived<DestinationStatus>(
		!result || result.status === 'failed' ? 'failed' : result.failed > 0 ? 'partial' : 'completed'
	);
	const everythingKnown = $derived(
		result?.status === 'completed' &&
			result.total_rows > 0 &&
			result.imported === 0 &&
			result.failed === 0
	);

	/**
	 * What the import has to say beyond its counts — only when it did not fully succeed.
	 * The classified `reason` is all the server returns about a failure; its own error
	 * message stays in the server log.
	 */
	const message = $derived.by(() => {
		if (!result) return 'The import did not report back.';
		if (result.status === 'failed')
			return 'The statement could not be processed. Nothing was added.';
		if (result.total_rows === 0) return 'The statement has no transaction rows. Nothing was added.';
		if (result.failed === 1) return '1 row could not be read and was skipped.';
		if (result.failed > 0) return `${result.failed} rows could not be read and were skipped.`;
		return '';
	});

	function count(key: ImportCountKey): number {
		return result ? result[key] : 0;
	}

	function countTone(key: ImportCountKey): string {
		const value = count(key);
		if (value === 0) return countToneClasses.zero;
		if (key === 'failed') return countToneClasses.failed;
		if (key === 'imported') return countToneClasses.imported;
		return countToneClasses.plain;
	}

	const cashflowState = $derived<TrailState>(
		phase === 'form'
			? 'pending'
			: !result
				? phase === 'processing'
					? 'running'
					: 'failed'
				: result.status === 'failed'
					? 'failed'
					: result.status !== 'completed'
						? 'running'
						: result.failed > 0
							? 'warning'
							: 'done'
	);

	const cashflowDetail = $derived.by(() => {
		if (phase === 'form') return undefined;
		if (!result) return phase === 'processing' ? 'Waiting' : 'No answer';
		if (result.status === 'failed') return 'Failed';
		if (result.status !== 'completed') return 'Reading…';
		return destinationStatusLabels[status];
	});

	// Matching only has something to do when the import added rows.
	const recurringStarted = $derived(
		phase === 'result' && result?.status === 'completed' && result.imported > 0
	);

	const steps = $derived<TrailStep[]>([
		{
			key: 'file',
			title: 'File',
			detail: phase === 'form' ? undefined : 'Accepted',
			state: phase === 'form' ? 'pending' : 'done'
		},
		{ key: 'cashflow', title: 'Cashflow', detail: cashflowDetail, state: cashflowState },
		{
			key: 'recurring',
			title: 'Recurring',
			detail:
				phase === 'form'
					? undefined
					: recurringStarted
						? 'Matching started'
						: phase === 'result'
							? 'Not run'
							: 'Waiting',
			state: recurringStarted ? 'done' : 'pending'
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

	const title = $derived(phase === 'result' ? 'Import complete' : 'Import a bank statement');

	/**
	 * Identifies the upload in flight. Closing the dialog abandons it — processing keeps
	 * running on the server, but a run that finishes after the user walked away must not
	 * land on a newer upload's result screen.
	 */
	let runId = $state(0);

	function reset() {
		phase = 'form';
		files = null;
		failure = '';
		result = null;
		runId += 1;
	}

	function close() {
		open = false;
		reset();
	}

	const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

	/** Polls the import until it reaches a terminal state, or gives up after the timeout. */
	async function poll(importId: string): Promise<ImportResult | null> {
		const deadline = Date.now() + POLL_TIMEOUT_MS;
		while (Date.now() < deadline) {
			const current = await getImport(importId);
			if (current.status === 'completed' || current.status === 'failed') return current;
			await sleep(POLL_MS);
		}
		return null;
	}

	function describe(cause: unknown): string {
		if (!(cause instanceof ApiError)) {
			return 'Could not confirm the upload was accepted. Check your connection and try again.';
		}
		if (cause.status === 404) return 'That account no longer exists. Close this and try again.';
		if (cause.status === 422)
			return 'This bank does not accept imports. Pick another bank account.';
		if (cause.status === 400) return `The file was rejected. Upload the CSV ${bankName} exported.`;
		return 'The upload could not be accepted. Try again.';
	}

	async function submit() {
		if (phase === 'processing' || !file || !vendorId) return;
		failure = '';
		result = null;
		phase = 'processing';

		const run = runId;
		let outcome: ImportResult | null;
		try {
			const accepted = await importCashflow({ file, vendor_id: vendorId });
			outcome = await poll(accepted.import_id);
		} catch (cause) {
			if (run !== runId) return;
			failure = describe(cause);
			phase = 'form';
			return;
		}
		if (run !== runId) return;

		// A file the parser did not recognise is the wrong export, which belongs on the
		// form where another file can be picked.
		if (outcome?.reason === 'file_not_recognised') {
			failure = wrongBankFileMessage(bankName);
			phase = 'form';
			return;
		}

		result = outcome;
		phase = 'result';
		onFinished?.();
	}

	function switchToBrokerage() {
		close();
		onBrokerageImport?.();
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

<Dialog bind:open {title} size="lg" closeOnBackdrop={phase !== 'processing'} onClose={reset}>
	{#if phase === 'result'}
		<div class="space-y-4">
			<Text as="p" size="sm" tone="muted">
				{file?.name ?? 'Uploaded statement'}{selectedVendor ? ` · ${selectedVendor.name}` : ''}
			</Text>

			{@render trail()}

			{#if everythingKnown}
				<NoticeBand
					intent="info"
					surface="inset"
					gutter="dialog"
					title="No new transactions"
					class="-mx-5"
				>
					Every row in this statement was already imported earlier. Nothing was added, and nothing
					was duplicated.
				</NoticeBand>
			{/if}

			<section class="border-t border-slate-400 pt-3" aria-label="Cashflow result">
				<div class="mb-2 flex items-center justify-between gap-2">
					<Heading level="h3" size="sm" class="text-slate-900">Cashflow</Heading>
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
							<dd class={['tabular-nums', countTone(row.key)].join(' ')}>{count(row.key)}</dd>
						</div>
					{/each}
				</dl>

				{#if message}
					<Text as="p" size="sm" tone="muted" class="mt-2">{message}</Text>
				{/if}
			</section>

			{#if recurringStarted}
				<Text as="p" size="sm" tone="muted">
					New rows that match a recurring item you confirmed are linked to it.
				</Text>
			{/if}
		</div>
	{:else}
		<div class="space-y-4" aria-busy={phase === 'processing'}>
			<Text as="p" size="sm" tone="muted">
				Upload the CSV your bank exported. Rows you already imported are recognised and skipped.
			</Text>

			{#if failure}
				<NoticeBand
					intent="error"
					surface="inset"
					gutter="dialog"
					title="That file was not imported"
					class="-mx-5"
				>
					{failure}
				</NoticeBand>
			{/if}

			{#if vendorOptions.length === 0}
				<NoticeBand
					intent="warning"
					surface="inset"
					gutter="dialog"
					title="No bank account"
					class="-mx-5"
				>
					Add an active bank vendor before importing a bank statement.
				</NoticeBand>
			{/if}

			<FormField label="Bank account" id="cfi-account" hint="The account this statement belongs to">
				{#snippet children(ctx)}
					<Select
						id={ctx.id}
						bind:value={vendorId}
						options={vendorOptions}
						placeholder="Select an account"
						ariaLabel="Bank account"
						disabled={phase === 'processing'}
					/>
				{/snippet}
			</FormField>

			<FormField label="Statement" id="cfi-file">
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

			{#if onBrokerageImport}
				<Text as="p" size="sm" tone="muted">
					Importing a DEGIRO statement?
					<button
						type="button"
						class="text-slate-900 underline underline-offset-2 hover:text-slate-700 disabled:text-slate-400 disabled:no-underline"
						disabled={phase === 'processing'}
						onclick={switchToBrokerage}
					>
						Use the brokerage import
					</button>
				</Text>
			{/if}
		</div>
	{/if}

	{#snippet footer()}
		{#if phase === 'result'}
			<Button variant="ghost" intent="secondary" onclick={reset}>Import another</Button>
			<Button onclick={close}>Done</Button>
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
