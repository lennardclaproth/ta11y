<script lang="ts">
	// Final design, screen 1 of 3 — "Ignore rules" (chosen from variant B).
	//
	// A page of its own under Cashflow: a ruled index of rules on the left, the opened rule with
	// its live preview on the right. The preview is the safety net the pitch asks for — what a
	// rule would catch is visible before it is allowed to catch anything. "Apply to existing"
	// names its count and asks for confirmation, because it is the one action that reaches back
	// over transactions that are already in the ledger.
	//
	// The page lays out its own content region instead of using PageContentTemplate: that template
	// gives narrow screens a fixed 32rem scroll box, which suits a ledger but traps a form.
	import PreviewShell from '../PreviewShell.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Switch from '$lib/components/atoms/switch/Switch.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import {
		ignoreRules,
		lastImport,
		previewMatches,
		previewSummary,
		sourceLabels
	} from '../ignore-rules.fixture';

	type Props = {
		/** Which half a narrow screen shows; both are side by side from `lg`. */
		mobileView?: 'list' | 'detail';
		/** No rules made yet. */
		empty?: boolean;
		loading?: boolean;
		error?: string | null;
		/** The rule as typed catches nothing — the preview says so instead of staying blank. */
		noMatches?: boolean;
		/** Show the "apply to existing" confirmation. */
		applyOpen?: boolean;
		/** Lay the shell out in document flow (narrow-screen previews). */
		flow?: boolean;
	};

	let {
		mobileView = 'list',
		empty = false,
		loading = false,
		error = null,
		noMatches = false,
		applyOpen = $bindable(false),
		flow = false
	}: Props = $props();

	let selectedId = $state('rule-2');
	let enabled = $state(true);

	const rules = $derived(empty ? [] : ignoreRules);
	const selected = $derived(ignoreRules.find((rule) => rule.id === selectedId) ?? ignoreRules[1]);

	const fieldOptions = [
		{ value: 'description', label: 'Description' },
		{ value: 'note', label: 'Note' }
	];
	const directionOptions = [
		{ value: 'any', label: 'Incoming and outgoing' },
		{ value: 'in', label: 'Incoming only' },
		{ value: 'out', label: 'Outgoing only' }
	];
	// A cashflow transaction carries the bank it was imported from, not a separate account.
	const sourceOptions = [
		{ value: 'all', label: 'All banks' },
		...sourceLabels.map((label) => ({ value: label, label }))
	];

	let fieldValue = $state('description');
	let directionValue = $state('out');
	let sourceValue = $state('ING');
	let nameValue = $state('Credit card payment');
	let containsValue = $state('Credit card');

	$effect(() => {
		if (noMatches) containsValue = 'Credit card settlement 2026';
	});

	const matching = $derived(noMatches ? 0 : previewSummary.matching);
	const sample = $derived(noMatches ? [] : previewMatches);

	const directionWords = { any: 'In and out', in: 'Incoming', out: 'Outgoing' };
</script>

<PreviewShell {flow}>
	{#snippet top()}
		<TopNavbar
			title="Ignore rules"
			accountName="Lennard Claproth"
			accountEmail="lennard@example.com"
		/>
	{/snippet}

	<div class="flex min-h-full flex-col px-4 pb-6 lg:h-full lg:min-h-0 lg:px-8">
		<Panel
			variant="muted"
			shape="square"
			shadow="none"
			padding="none"
			class="flex flex-col overflow-hidden lg:min-h-0 lg:flex-1"
		>
			<div class="grid grid-cols-1 lg:min-h-0 lg:flex-1 lg:grid-cols-[23rem_1fr]">
				<!-- Index of rules -->
				<section
					class={[
						'min-h-0 flex-col border-b border-slate-200 bg-white lg:flex lg:border-r lg:border-b-0',
						mobileView === 'detail' ? 'hidden' : 'flex'
					].join(' ')}
					aria-label="Ignore rules"
				>
					<div
						class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
					>
						<h2 class="text-2xl">Rules</h2>
						<Button size="sm" shape="default" onclick={() => {}}>
							<Icon icon="heroicons:plus" />New rule
						</Button>
					</div>

					<!-- The last import reports itself here, and links to its own review page. -->
					<div
						class="shrink-0 border-b border-slate-200 bg-taupe-50 px-4 py-3"
						aria-label="Last import"
					>
						<Text as="p" size="sm" tone="muted">Last import · {lastImport.file}</Text>
						<Text as="p" size="sm" class="text-slate-800">
							{lastImport.imported} new · {lastImport.duplicates} duplicates ·
							<span class="text-slate-900">{lastImport.autoIgnored} auto-ignored</span>
						</Text>
						<a
							href="#review"
							class="mt-1 inline-block rounded-md text-sm text-sky-700 underline underline-offset-2 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
						>
							Review what was ignored
						</a>
					</div>

					<ul class="min-h-0 lg:flex-1 lg:overflow-auto">
						{#if loading}
							{#each [0, 1, 2, 3] as index (index)}
								<li class="border-b border-slate-100 px-4 py-3">
									<Skeleton class="mb-2 w-40" />
									<Skeleton class="w-56" />
								</li>
							{/each}
						{:else if error}
							<li class="px-4 py-6">
								<Alert intent="error" title="Could not load your rules">
									{error}
									<Button
										size="sm"
										variant="outline"
										intent="secondary"
										shape="default"
										class="mt-2"
									>
										Try again
									</Button>
								</Alert>
							</li>
						{:else if rules.length === 0}
							<li class="px-4 py-10 text-center">
								<Text as="p" size="sm" tone="muted">
									No ignore rules yet. Make one here, or from a transaction you keep ignoring by
									hand on Cashflow.
								</Text>
							</li>
						{:else}
							{#each rules as rule (rule.id)}
								{@const active = rule.id === selectedId}
								<li class="border-b border-slate-100">
									<button
										type="button"
										onclick={() => (selectedId = rule.id)}
										aria-current={active ? 'true' : undefined}
										class={[
											'flex w-full items-start gap-3 px-4 py-3 text-left transition-colors',
											'focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-slate-500',
											active ? 'bg-amber-50' : 'hover:bg-slate-50'
										].join(' ')}
									>
										<span class="min-w-0 flex-1">
											<span class="flex items-center gap-2">
												<span class="truncate font-heading text-lg text-slate-900">{rule.name}</span>
												{#if !rule.enabled}
													<Badge intent="neutral" variant="outline" size="sm">Off</Badge>
												{/if}
											</span>
											<span class="mt-0.5 block truncate text-sm text-slate-600">
												{rule.field === 'description' ? 'Description' : 'Note'} contains “{rule.contains}”
											</span>
											<span class="block truncate text-xs text-slate-500">
												{directionWords[rule.direction]} · {rule.source ?? 'All banks'}
											</span>
										</span>
										<span class="shrink-0 text-right">
											<span class="block text-sm text-slate-900 tabular-nums">
												{rule.ignoredTotal}
											</span>
											<span class="block text-xs text-slate-500">ignored</span>
										</span>
									</button>
								</li>
							{/each}
						{/if}
					</ul>
				</section>

				<!-- The opened rule -->
				<section
					class={[
						'min-h-0 flex-col bg-white lg:flex lg:overflow-auto',
						mobileView === 'list' ? 'hidden' : 'flex'
					].join(' ')}
					aria-label="Rule detail"
				>
					{#if empty || error}
						<!-- Nothing to open: explain what a rule is instead of showing a dead form. -->
						<div
							class="flex min-h-0 flex-1 flex-col items-center justify-center gap-3 px-6 py-16 text-center"
						>
							<Heading level="h2" size="lg" class="text-slate-900">
								{error ? 'No rule open' : 'Your first ignore rule'}
							</Heading>
							<Text as="p" size="sm" tone="muted" class="max-w-md">
								{#if error}
									Your rules did not load, so there is nothing to open. Try again on the left.
								{:else}
									A rule matches text in the description or note, in one direction, for one bank or
									all. Transactions it matches arrive ignored and stay out of your monthly totals —
									you see which rule did it, and can put any of them back.
								{/if}
							</Text>
							{#if !error}
								<Button shape="default" onclick={() => {}}>
									<Icon icon="heroicons:plus" />New rule
								</Button>
							{/if}
						</div>
					{:else}
					<div
						class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
					>
						<div class="flex min-w-0 items-center gap-2">
							<a
								href="#rules"
								class="inline-flex items-center gap-1 rounded-md text-sm text-sky-700 underline underline-offset-2 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none lg:hidden"
							>
								<Icon icon="heroicons:chevron-left" size="sm" />Rules
							</a>
							<Heading level="h2" size="xl" class="truncate text-slate-900">{selected.name}</Heading>
						</div>
						<label class="flex items-center gap-2 text-sm text-slate-700">
							<Switch bind:checked={enabled} />
							{enabled ? 'Rule is on' : 'Rule is off'}
						</label>
					</div>

					<div class="min-h-0 space-y-4 px-4 py-3 lg:flex-1 lg:px-6">
						<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
							<FormField label="Rule name" id="rule-name" class="sm:col-span-2">
								{#snippet children(ctx)}
									<Input id={ctx.id} bind:value={nameValue} />
								{/snippet}
							</FormField>

							<FormField label="Match on" id="rule-field">
								{#snippet children(ctx)}
									<Select id={ctx.id} bind:value={fieldValue} options={fieldOptions} />
								{/snippet}
							</FormField>

							<FormField label="Contains" id="rule-contains" hint="Not case sensitive">
								{#snippet children(ctx)}
									<Input id={ctx.id} bind:value={containsValue} ariaDescribedby={ctx.describedby} />
								{/snippet}
							</FormField>

							<FormField label="Direction" id="rule-direction">
								{#snippet children(ctx)}
									<Select id={ctx.id} bind:value={directionValue} options={directionOptions} />
								{/snippet}
							</FormField>

							<FormField
								label="Bank"
								id="rule-source"
								hint="Transactions carry the bank they were imported from"
							>
								{#snippet children(ctx)}
									<Select
										id={ctx.id}
										bind:value={sourceValue}
										options={sourceOptions}
										ariaDescribedby={ctx.describedby}
									/>
								{/snippet}
							</FormField>
						</div>

						<!-- Preview: a rule that is too wide has to be visible before it is saved. -->
						<section class="border-t border-slate-400 pt-2" aria-label="Matching transactions">
							<div class="mb-1 flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
								<Heading level="h3" size="sm" class="text-slate-900">
									Matches {matching} of {previewSummary.scanned.toLocaleString('en')} transactions
								</Heading>
								{#if matching > 0}
									<Text as="span" size="sm" tone="muted">Showing the 4 most recent</Text>
								{/if}
							</div>

							{#if matching === 0}
								<Alert intent="info" title="Nothing matches this rule yet">
									No transaction in your ledger contains “{containsValue}”. Save it anyway to catch
									future imports, or widen the text.
								</Alert>
							{:else}
								<ul class="text-sm">
									{#each sample as match (match.id)}
										<li class="border-b border-slate-200 py-1.5 last:border-b-0">
											<!-- Narrow: the description gets the width, the date reads as a caption. -->
											<div class="sm:hidden">
												<div class="flex items-baseline justify-between gap-3">
													<span class="min-w-0 flex-1 truncate text-slate-800">
														{match.description}
													</span>
													<span class="shrink-0">
														<Money
															amount={scaledToNumber(match.amountCents)}
															currency="EUR"
															size="sm"
														/>
													</span>
												</div>
												<span class="text-xs text-slate-500 tabular-nums">
													{formatDisplayDate(match.date.slice(0, 10))}
												</span>
											</div>

											<div class="hidden items-baseline gap-3 sm:flex">
												<span class="w-24 shrink-0 text-slate-500 tabular-nums">
													{formatDisplayDate(match.date.slice(0, 10))}
												</span>
												<span class="min-w-0 flex-1 truncate text-slate-800">
													{match.description}
												</span>
												<span class="w-28 shrink-0 text-right">
													<Money amount={scaledToNumber(match.amountCents)} currency="EUR" size="sm" />
												</span>
											</div>
										</li>
									{/each}
								</ul>

								<Alert intent="warning" title="One match looks unrelated" class="mt-2">
									“Credit card annual fee” is a real cost, not an own transfer. Narrow the text if
									you do not want it ignored.
								</Alert>
							{/if}
						</section>
					</div>

					<div
						class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3 lg:px-6"
					>
						<Text as="span" size="sm" tone="muted">Saving only affects imports from now on.</Text>
						<div class="flex flex-wrap items-center gap-2">
							<Button
								variant="outline"
								intent="secondary"
								shape="default"
								disabled={matching === 0}
								onclick={() => (applyOpen = true)}
							>
								Apply to existing ({matching})
							</Button>
							<Button shape="default">Save changes</Button>
						</div>
					</div>
					{/if}
				</section>
			</div>
		</Panel>
	</div>
</PreviewShell>

<!-- Reaching back over the ledger is the one action that changes what is already there. -->
<Dialog bind:open={applyOpen} size="md" title="Apply this rule to existing transactions">
	<div class="space-y-3 px-5 py-4">
		<Text as="p" size="md" class="text-slate-800">
			<span class="text-slate-900 tabular-nums">{previewSummary.notYetIgnored}</span> transactions
			match “{nameValue}” and are not ignored yet. They move out of your monthly totals and tag
			split.
		</Text>
		<Text as="p" size="sm" tone="muted">
			Nothing is deleted. You can show ignored rows on Cashflow and put any of them back.
		</Text>
	</div>
	{#snippet footer()}
		<div class="flex flex-wrap justify-end gap-2 border-t border-slate-200 px-5 py-3">
			<Button variant="ghost" intent="secondary" shape="default" onclick={() => (applyOpen = false)}>
				Cancel
			</Button>
			<Button shape="default">Ignore {previewSummary.notYetIgnored} transactions</Button>
		</div>
	{/snippet}
</Dialog>
