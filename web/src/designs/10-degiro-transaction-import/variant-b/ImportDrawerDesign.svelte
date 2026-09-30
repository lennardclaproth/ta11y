<script lang="ts">
	// Variant B — "Zijpaneel met importlogboek".
	// Prototype for #10: the page stays visible while the import reports step by step.
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import Dropzone from '$lib/components/molecules/dropzone/Dropzone.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Spinner from '$lib/components/atoms/spinner/Spinner.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import DesignBackdrop from '../DesignBackdrop.svelte';
	import {
		accountOptions,
		destinationResults,
		importFileName,
		importFileSize,
		unlinkedProducts,
		wrongFileMessage
	} from '../import-fixtures';

	type Phase = 'form' | 'processing' | 'result' | 'error';

	let { phase = 'form' as Phase }: { phase?: Phase } = $props();

	let accountId = $state('vnd-degiro');

	type StepState = 'done' | 'warning' | 'running' | 'pending' | 'failed';

	interface Step {
		title: string;
		detail: string;
		state: StepState;
	}

	const cashflow = destinationResults[0];
	const portfolio = destinationResults[1];

	const summary = (r: typeof cashflow) =>
		`${r.imported} new · ${r.duplicates} already imported · ${r.failed} failed`;

	const resultSteps: Step[] = [
		{ title: 'File accepted', detail: `${importFileName} · ${cashflow.total} rows`, state: 'done' },
		{ title: 'Cashflow', detail: summary(cashflow), state: 'done' },
		{ title: 'Portfolio', detail: summary(portfolio), state: 'warning' },
		{ title: 'Performance recalculated', detail: 'Value, result and cost basis are up to date', state: 'done' }
	];

	const processingSteps: Step[] = [
		{ title: 'File accepted', detail: `${importFileName} · ${cashflow.total} rows`, state: 'done' },
		{ title: 'Cashflow', detail: summary(cashflow), state: 'done' },
		{ title: 'Portfolio', detail: 'Reading transactions…', state: 'running' },
		{ title: 'Performance recalculated', detail: 'Waiting for the portfolio import', state: 'pending' }
	];

	const steps = $derived(phase === 'result' ? resultSteps : phase === 'processing' ? processingSteps : []);

	const markerClasses = {
		done: 'border-emerald-700 bg-emerald-700 text-white',
		warning: 'border-amber-500 bg-amber-500 text-slate-900',
		failed: 'border-red-700 bg-red-700 text-white',
		running: 'border-slate-400 bg-white text-slate-600',
		pending: 'border-slate-300 bg-white text-slate-300'
	} satisfies Record<StepState, string>;

	const markerIcon = {
		done: 'heroicons:check',
		warning: 'heroicons:exclamation-triangle',
		failed: 'heroicons:x-mark',
		running: 'heroicons:arrow-path',
		pending: 'heroicons:minus'
	} satisfies Record<StepState, string>;
</script>

<DesignBackdrop label="Portfolio" />

<Drawer open title="Import DEGIRO export" width="max-w-lg" dismissible>
	<div class="space-y-4">
		{#if phase === 'result'}
			<!-- The form collapses once there is a result: the log is what you came back for. -->
			<div class="flex items-baseline justify-between gap-3 border-b border-slate-200 pb-3">
				<div class="min-w-0">
					<Text as="span" size="sm" weight="medium" class="block truncate text-slate-900">
						{importFileName}
					</Text>
					<Text as="span" size="sm" tone="muted" class="block">
						{accountOptions[0].label} · {importFileSize}
					</Text>
				</div>
				<Button size="sm" variant="ghost" intent="secondary" class="shrink-0">
					Import another
				</Button>
			</div>
		{:else}
			<Text as="p" size="sm" tone="muted">
				One upload fills Cashflow and Portfolio. Rows you already imported are recognised and
				skipped.
			</Text>

			{#if phase === 'error'}
				<Alert intent="error" title="That file was not imported">{wrongFileMessage}</Alert>
			{/if}

			<FormField
				label="Account"
				id="imp-b-account"
				hint="The brokerage account this export belongs to"
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

			<FormField label="DEGIRO export" id="imp-b-file">
				<Dropzone
					accept=".csv"
					disabled={phase === 'processing'}
					label={phase === 'processing' ? importFileName : 'Drag the CSV here, or click to browse'}
					hint={phase === 'processing' ? importFileSize : 'One file, .csv, up to 10 MB'}
				/>
			</FormField>

			<Button class="w-full" loading={phase === 'processing'} disabled={phase === 'processing'}>
				{phase === 'processing' ? 'Importing…' : 'Import'}
			</Button>
		{/if}

		<section class="border-t border-slate-400 pt-3" aria-live="polite">
			<Heading level="h3" size="sm" class="mb-3 text-slate-900">Activity</Heading>

			{#if steps.length === 0}
				<Text as="p" size="sm" tone="muted">
					Nothing imported yet. After you upload, each destination reports here.
				</Text>
			{:else}
				<ol>
					{#each steps as step, index (step.title)}
						<li class="relative flex gap-3 pb-4 last:pb-0">
							{#if index < steps.length - 1}
								<span
									class="absolute top-7 bottom-0 left-3 w-px -translate-x-1/2 bg-slate-300"
									aria-hidden="true"
								></span>
							{/if}

							<span
								class={[
									'flex size-6 shrink-0 items-center justify-center rounded-full border',
									markerClasses[step.state]
								].join(' ')}
							>
								<Icon
									icon={markerIcon[step.state]}
									size="sm"
									class={step.state === 'running' ? 'animate-spin motion-reduce:animate-none' : ''}
								/>
							</span>

							<div class="min-w-0 flex-1">
								<Text as="span" size="sm" weight="medium" class="block text-slate-900">
									{step.title}
								</Text>
								<Text as="span" size="sm" tone="muted" class="block tabular-nums">
									{step.detail}
								</Text>
							</div>
						</li>
					{/each}
				</ol>

				{#if phase === 'result' && portfolio.message}
					<Text as="p" size="sm" tone="muted" class="mt-3">
						Portfolio: {portfolio.message}
					</Text>
				{/if}
			{/if}

			{#if phase === 'processing'}
				<div class="flex items-center gap-2 pt-1">
					<Spinner size="sm" />
					<Text as="span" size="sm" tone="muted">
						Keeps running if you close this panel.
					</Text>
				</div>
			{/if}
		</section>

		{#if phase === 'result'}
			<Alert intent="warning" title="{unlinkedProducts.length} products without a listing">
				<p class="mb-2">
					These do not count towards performance yet. Add them in Listings, then run Rebuild
					portfolio.
				</p>
				<ul class="space-y-1">
					{#each unlinkedProducts as product (product.id)}
						<li class="flex items-baseline justify-between gap-3 border-b border-amber-200 pb-1 last:border-b-0">
							<span class="min-w-0">
								{product.name}
								<span class="text-slate-600">— {product.isin ?? product.symbol ?? 'no identifier'}</span>
							</span>
							<span class="shrink-0 tabular-nums">{product.transactions} tx</span>
						</li>
					{/each}
				</ul>
			</Alert>
		{/if}
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary">Close</Button>
		{#if phase === 'result'}
			<Button>Go to portfolio</Button>
		{/if}
	{/snippet}
</Drawer>
