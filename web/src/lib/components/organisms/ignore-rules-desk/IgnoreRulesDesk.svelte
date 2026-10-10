<script lang="ts">
	// The rules page: a ruled index of rules on the left, the opened rule with its live
	// preview on the right. Writing a rule and seeing what it catches belong together —
	// a rule that is too wide has to be visible here, before it is allowed to ignore
	// anything — so the editor never leaves the page the preview lives on.
	//
	// The desk is presentational: it owns the form's local state and reports every
	// decision upward. Loading, saving and applying are the page's business.
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import IgnoreRuleListItem from '$lib/components/molecules/ignore-rule-list-item/IgnoreRuleListItem.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import RuleMatchPreview from '$lib/components/molecules/rule-match-preview/RuleMatchPreview.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Switch from '$lib/components/atoms/switch/Switch.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import type { IgnoreRule, IgnoreRulePreview } from '$lib/api/types';
	import {
		matchFieldOptions,
		ruleDirectionOptions,
		type DeskPane,
		type IgnoreRuleDraft
	} from './ignore-rules-desk.types';

	type Props = {
		rules: IgnoreRule[];
		/** The rule being written. `null` closes the right half. */
		draft?: IgnoreRuleDraft | null;
		/** What the draft matches today; `null` while it has not been checked. */
		preview?: IgnoreRulePreview | null;
		previewLoading?: boolean;
		previewError?: string | null;
		/** Banks a rule can be limited to, from the sources the ledger actually holds. */
		bankOptions?: { value: string; label: string }[];
		loading?: boolean;
		error?: string | null;
		saving?: boolean;
		/** Per-field refusals from the API, keyed by field name. */
		fieldErrors?: Record<string, string>;
		/** Which half a narrow screen shows. */
		pane?: DeskPane;
		onNew?: () => void;
		onSelect?: (rule: IgnoreRule) => void;
		onDraftChange?: (draft: IgnoreRuleDraft) => void;
		onSave?: (draft: IgnoreRuleDraft) => void;
		onDelete?: (draft: IgnoreRuleDraft) => void;
		onApply?: (draft: IgnoreRuleDraft) => void;
		onRetry?: () => void;
		class?: string;
		/** Rendered above the index — the last import reporting itself. */
		aside?: import('svelte').Snippet;
	};

	let {
		rules,
		draft = $bindable(null),
		preview = null,
		previewLoading = false,
		previewError = null,
		bankOptions = [],
		loading = false,
		error = null,
		saving = false,
		fieldErrors = {},
		pane = $bindable('list'),
		onNew,
		onSelect,
		onDraftChange,
		onSave,
		onDelete,
		onApply,
		onRetry,
		class: className = '',
		aside
	}: Props = $props();

	const sourceOptions = $derived([{ value: '', label: 'All banks' }, ...bankOptions]);
	const notYetIgnored = $derived(preview?.not_yet_ignored ?? 0);
	const isNew = $derived(draft?.id === null);

	/** Reports the edited draft upward; the page re-checks the preview from it. */
	function edit(changes: Partial<IgnoreRuleDraft>) {
		if (!draft) return;
		draft = { ...draft, ...changes };
		onDraftChange?.(draft);
	}

	function select(rule: IgnoreRule) {
		pane = 'detail';
		onSelect?.(rule);
	}

	const classes = $derived(
		['flex flex-col overflow-hidden lg:min-h-0 lg:flex-1', className].filter(Boolean).join(' ')
	);
</script>

<Panel variant="muted" shape="square" shadow="none" padding="none" class={classes}>
	<div class="grid grid-cols-1 lg:min-h-0 lg:flex-1 lg:grid-cols-[23rem_1fr]">
		<section
			class={[
				'min-h-0 flex-col border-b border-slate-200 bg-white lg:flex lg:border-r lg:border-b-0',
				pane === 'detail' ? 'hidden' : 'flex'
			].join(' ')}
			aria-label="Ignore rules"
		>
			<div
				class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
			>
				<h2 class="text-2xl">Rules</h2>
				<Button size="sm" shape="default" onclick={() => onNew?.()}>
					<Icon icon="heroicons:plus" />New rule
				</Button>
			</div>

			{@render aside?.()}

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
								onclick={() => onRetry?.()}
							>
								Try again
							</Button>
						</Alert>
					</li>
				{:else if rules.length === 0}
					<li class="px-4 py-10 text-center">
						<Text as="p" size="sm" tone="muted">
							No ignore rules yet. Make one here, or from a transaction you keep ignoring by hand
							on Cashflow.
						</Text>
					</li>
				{:else}
					{#each rules as rule (rule.id)}
						<li class="border-b border-slate-100">
							<IgnoreRuleListItem {rule} active={rule.id === draft?.id} onSelect={select} />
						</li>
					{/each}
				{/if}
			</ul>
		</section>

		<section
			class={[
				'min-h-0 flex-col bg-white lg:flex lg:overflow-auto',
				pane === 'list' ? 'hidden' : 'flex'
			].join(' ')}
			aria-label="Rule detail"
		>
			{#if !draft}
				<!-- Nothing open: explain what a rule is instead of showing a dead form. -->
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
						<Button shape="default" onclick={() => onNew?.()}>
							<Icon icon="heroicons:plus" />New rule
						</Button>
					{/if}
				</div>
			{:else}
				<div
					class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
				>
					<div class="flex min-w-0 items-center gap-2">
						<Button
							size="sm"
							variant="ghost"
							intent="secondary"
							shape="default"
							class="lg:hidden"
							onclick={() => (pane = 'list')}
						>
							<Icon icon="heroicons:chevron-left" size="sm" />Rules
						</Button>
						<Heading level="h2" size="xl" class="truncate text-slate-900">
							{draft.name || 'New rule'}
						</Heading>
						{#if isNew}
							<Badge intent="info" variant="soft" size="sm">Not saved yet</Badge>
						{/if}
					</div>
					<!-- The state is a word next to the switch, not only its position, so it
					     survives a screenshot and reads the same to a screen reader. -->
					<div class="flex items-center gap-2 text-sm text-slate-700">
						<Switch
							checked={draft.enabled}
							aria-label="Rule is on"
							onchange={(event) =>
								edit({ enabled: (event.currentTarget as HTMLInputElement).checked })}
						/>
						<span>{draft.enabled ? 'Rule is on' : 'Rule is off'}</span>
					</div>
				</div>

				<div class="min-h-0 space-y-4 px-4 py-3 lg:flex-1 lg:px-6">
					<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
						<FormField
							label="Rule name"
							id="rule-name"
							class="sm:col-span-2"
							error={fieldErrors.name}
						>
							{#snippet children(ctx)}
								<Input
									id={ctx.id}
									value={draft?.name ?? ''}
									intent={ctx.invalid ? 'error' : 'default'}
									ariaDescribedby={ctx.describedby}
									oninput={(event) =>
										edit({ name: (event.currentTarget as HTMLInputElement).value })}
								/>
							{/snippet}
						</FormField>

						<FormField label="Match on" id="rule-field" error={fieldErrors.match_field}>
							{#snippet children(ctx)}
								<Select
									id={ctx.id}
									value={draft?.match_field ?? 'description'}
									options={matchFieldOptions}
									ariaDescribedby={ctx.describedby}
									onchange={(event) =>
										edit({
											match_field: (event.currentTarget as HTMLSelectElement)
												.value as IgnoreRuleDraft['match_field']
										})}
								/>
							{/snippet}
						</FormField>

						<FormField
							label="Contains"
							id="rule-contains"
							hint="Not case sensitive"
							error={fieldErrors.contains}
						>
							{#snippet children(ctx)}
								<Input
									id={ctx.id}
									value={draft?.contains ?? ''}
									intent={ctx.invalid ? 'error' : 'default'}
									ariaDescribedby={ctx.describedby}
									oninput={(event) =>
										edit({ contains: (event.currentTarget as HTMLInputElement).value })}
								/>
							{/snippet}
						</FormField>

						<FormField label="Direction" id="rule-direction" error={fieldErrors.direction}>
							{#snippet children(ctx)}
								<Select
									id={ctx.id}
									value={draft?.direction ?? ''}
									options={ruleDirectionOptions}
									ariaDescribedby={ctx.describedby}
									onchange={(event) =>
										edit({
											direction: (event.currentTarget as HTMLSelectElement)
												.value as IgnoreRuleDraft['direction']
										})}
								/>
							{/snippet}
						</FormField>

						<!-- A cashflow transaction carries the bank it was imported from, not an
						     account of its own, so that is what a rule can honestly scope on. -->
						<FormField
							label="Bank"
							id="rule-source"
							hint="Transactions carry the bank they were imported from"
						>
							{#snippet children(ctx)}
								<Select
									id={ctx.id}
									value={draft?.source ?? ''}
									options={sourceOptions}
									ariaDescribedby={ctx.describedby}
									onchange={(event) =>
										edit({ source: (event.currentTarget as HTMLSelectElement).value })}
								/>
							{/snippet}
						</FormField>
					</div>

					<RuleMatchPreview
						matching={preview?.matching ?? 0}
						scanned={preview?.scanned ?? 0}
						sample={preview?.sample ?? []}
						contains={draft.contains}
						loading={previewLoading}
						error={previewError}
					/>
				</div>

				<div
					class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3 lg:px-6"
				>
					<Text as="span" size="sm" tone="muted">Saving only affects imports from now on.</Text>
					<div class="flex flex-wrap items-center gap-2">
						{#if !isNew}
							<Button
								variant="ghost"
								intent="secondary"
								shape="default"
								disabled={saving}
								onclick={() => draft && onDelete?.(draft)}
							>
								Delete
							</Button>
						{/if}
						<Button
							variant="outline"
							intent="secondary"
							shape="default"
							disabled={saving || isNew || notYetIgnored === 0}
							onclick={() => draft && onApply?.(draft)}
						>
							Apply to existing ({notYetIgnored})
						</Button>
						<Button
							shape="default"
							loading={saving}
							onclick={() => draft && onSave?.(draft)}
						>
							{isNew ? 'Save rule' : 'Save changes'}
						</Button>
					</div>
				</div>
			{/if}
		</section>
	</div>
</Panel>
