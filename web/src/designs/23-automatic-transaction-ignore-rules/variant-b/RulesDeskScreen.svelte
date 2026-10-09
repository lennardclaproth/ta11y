<script lang="ts">
	// Variant B — "Rules desk". Ignore rules get their own page under Cashflow: a ruled index of
	// rules on the left, the opened rule with its live preview on the right. The preview is the
	// safety net — you see what a rule would catch before it is allowed to catch anything.
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
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import {
		accountLabels,
		ignoreRules,
		lastImport,
		previewMatches,
		previewSummary
	} from '../ignore-rules.fixture';

	type Props = {
		/** Which half a narrow screen shows; both are side by side from `lg`. */
		mobileView?: 'list' | 'detail';
		/** No rules made yet. */
		empty?: boolean;
		loading?: boolean;
		error?: string | null;
		/** Lay the shell out in document flow (narrow-screen previews). */
		flow?: boolean;
	};

	let {
		mobileView = 'list',
		empty = false,
		loading = false,
		error = null,
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
	const accountOptions = [
		{ value: 'all', label: 'All accounts' },
		...accountLabels.map((label) => ({ value: label, label }))
	];

	let fieldValue = $state('description');
	let containsValue = $state('Credit card');
	let directionValue = $state('out');
	let accountValue = $state('ING — Current');
	let nameValue = $state('Credit card payment');

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
								<Alert intent="error" title="Could not load your rules">{error}</Alert>
							</li>
						{:else if rules.length === 0}
							<li class="px-4 py-10 text-center">
								<Text as="p" size="sm" tone="muted">
									No ignore rules yet. Make one from a transaction you keep ignoring by hand.
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
												{directionWords[rule.direction]} · {rule.account ?? 'All accounts'}
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

					<div class="min-h-0 space-y-5 px-4 py-4 lg:flex-1 lg:px-6">
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

							<FormField label="Account" id="rule-account">
								{#snippet children(ctx)}
									<Select id={ctx.id} bind:value={accountValue} options={accountOptions} />
								{/snippet}
							</FormField>
						</div>

						<!-- Preview: a rule that is too wide has to be visible before it is saved. -->
						<section class="border-t border-slate-400 pt-3" aria-label="Matching transactions">
							<div class="mb-2 flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
								<Heading level="h3" size="sm" class="text-slate-900">
									Matches {previewSummary.matching} of {previewSummary.scanned.toLocaleString('en')} transactions
								</Heading>
								<Text as="span" size="sm" tone="muted">Showing the 4 most recent</Text>
							</div>

							<ul class="text-sm">
								{#each previewMatches as match (match.id)}
									<li
										class="flex items-baseline justify-between gap-3 border-b border-slate-200 py-2 last:border-b-0"
									>
										<span class="w-24 shrink-0 text-slate-500 tabular-nums">
											{formatDisplayDate(match.date.slice(0, 10))}
										</span>
										<span class="min-w-0 flex-1 truncate text-slate-800">{match.description}</span>
										<span class="shrink-0">
											<Money amount={scaledToNumber(match.amountCents)} currency="EUR" size="sm" />
										</span>
									</li>
								{/each}
							</ul>

							<Alert intent="warning" title="One match looks unrelated" class="mt-3">
								“Credit card annual fee” is a real cost, not an own transfer. Narrow the text if you
								do not want it ignored.
							</Alert>
						</section>
					</div>

					<div
						class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3 lg:px-6"
					>
						<Text as="span" size="sm" tone="muted">Saving only affects imports from now on.</Text>
						<div class="flex flex-wrap items-center gap-2">
							<Button variant="outline" intent="secondary" shape="default">
								Apply to existing ({previewSummary.notYetIgnored})
							</Button>
							<Button shape="default">Save changes</Button>
						</div>
					</div>
				</section>
			</div>
		</Panel>
	</div>
</PreviewShell>
