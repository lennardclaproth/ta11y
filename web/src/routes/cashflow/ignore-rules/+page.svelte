<script lang="ts">
	// Ignore rules live on a page of their own under Cashflow: rules are written, read and
	// applied in one place, and the preview that keeps a rule from being too wide sits
	// right under the form that writes it.
	//
	// The page lays out its own content region instead of using PageContentTemplate: that
	// template gives narrow screens a fixed 32rem scroll box, which suits a ledger but
	// traps a form.
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import IgnoreRulesDesk from '$lib/components/organisms/ignore-rules-desk/IgnoreRulesDesk.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { ApiError } from '$lib/api/client';
	import {
		applyIgnoreRule,
		createIgnoreRule,
		deleteIgnoreRule,
		listIgnoreRules,
		previewIgnoreRule,
		updateIgnoreRule
	} from '$lib/services/ignoreRules';
	import { listVendors } from '$lib/services/vendors';
	import { toast } from '$lib/stores/toast.svelte';
	import {
		draftFromRule,
		emptyDraft,
		type DeskPane,
		type IgnoreRuleDraft
	} from '$lib/components/organisms/ignore-rules-desk/ignore-rules-desk.types';
	import type { IgnoreRule, IgnoreRulePreview, IgnoreRuleRequest } from '$lib/api/types';

	/** How long typing settles before the preview is re-checked. */
	const PREVIEW_DEBOUNCE_MS = 400;
	/** Below this the rule is refused anyway, so there is nothing worth previewing. */
	const MIN_CONTAINS_LENGTH = 3;

	let rules = $state<IgnoreRule[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let draft = $state<IgnoreRuleDraft | null>(null);
	let pane = $state<DeskPane>('list');
	let saving = $state(false);
	let fieldErrors = $state<Record<string, string>>({});

	let preview = $state<IgnoreRulePreview | null>(null);
	let previewLoading = $state(false);
	let previewError = $state<string | null>(null);

	let applyOpen = $state(false);
	let applying = $state(false);

	/**
	 * The banks a rule can be limited to. A cashflow transaction carries the source it
	 * was imported from rather than an account of its own, so the choices are the cashflow
	 * vendors — the same names an import stamps on a row — rather than an invented list.
	 */
	let vendorNames = $state<string[]>([]);
	const bankOptions = $derived(
		[...new Set([...vendorNames, ...rules.map((rule) => rule.source).filter(Boolean)])]
			.sort()
			.map((source) => ({ value: source, label: source }))
	);

	async function load() {
		loading = true;
		error = null;
		try {
			rules = await listIgnoreRules();
		} catch {
			error = 'The rules did not load. Check your connection.';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void load();
		// A rule the vendor list never arrives for simply offers "all banks" plus whatever
		// the existing rules already name, so editing one never loses its own scope.
		void listVendors()
			.then((all) => {
				vendorNames = all.filter((v) => v.active && v.type === 'cashflow').map((v) => v.name);
			})
			.catch(() => undefined);
	});

	// A draft arriving as ?draft=<description> is the Cashflow page handing over a rule it
	// wants written, prefilled from the rows selected there.
	onMount(() => {
		const prefill = page.url.searchParams.get('contains');
		if (!prefill) return;
		draft = emptyDraft({ name: prefill, contains: prefill });
		pane = 'detail';
		schedulePreview(draft);
	});

	let previewTimer: ReturnType<typeof setTimeout> | null = null;

	function toRequest(value: IgnoreRuleDraft): IgnoreRuleRequest {
		return {
			name: value.name,
			match_field: value.match_field,
			contains: value.contains,
			direction: value.direction,
			source: value.source,
			enabled: value.enabled
		};
	}

	/**
	 * Re-checks what the draft catches once typing settles. Debounced rather than run per
	 * keystroke: the preview is a read over the whole ledger, and a half-typed word would
	 * report a number that is about to be wrong.
	 */
	function schedulePreview(value: IgnoreRuleDraft | null) {
		if (previewTimer) clearTimeout(previewTimer);
		if (!value || value.contains.trim().length < MIN_CONTAINS_LENGTH) {
			preview = null;
			previewError = null;
			previewLoading = false;
			return;
		}
		previewLoading = true;
		previewError = null;
		previewTimer = setTimeout(() => void runPreview(value), PREVIEW_DEBOUNCE_MS);
	}

	async function runPreview(value: IgnoreRuleDraft) {
		try {
			preview = await previewIgnoreRule(toRequest(value));
		} catch {
			preview = null;
			previewError = 'The check did not come back. Try again.';
		} finally {
			previewLoading = false;
		}
	}

	function openNew() {
		draft = emptyDraft();
		fieldErrors = {};
		preview = null;
		pane = 'detail';
	}

	function openRule(rule: IgnoreRule) {
		draft = draftFromRule(rule);
		fieldErrors = {};
		schedulePreview(draft);
	}

	function onDraftChange(value: IgnoreRuleDraft) {
		fieldErrors = {};
		schedulePreview(value);
	}

	async function save(value: IgnoreRuleDraft) {
		saving = true;
		fieldErrors = {};
		try {
			const saved = value.id
				? await updateIgnoreRule(value.id, toRequest(value))
				: await createIgnoreRule(toRequest(value));
			rules = value.id
				? rules.map((rule) => (rule.id === saved.id ? saved : rule))
				: [saved, ...rules];
			draft = draftFromRule(saved);
			toast.success(value.id ? 'Rule saved' : 'Rule created');
		} catch (cause) {
			// The API answers a refused rule per field, so the message lands next to the
			// input that caused it rather than as a banner that says "something is wrong".
			fieldErrors = fieldProblems(cause);
			if (Object.keys(fieldErrors).length === 0) toast.error('Could not save the rule');
		} finally {
			saving = false;
		}
	}

	function fieldProblems(cause: unknown): Record<string, string> {
		if (!(cause instanceof ApiError) || cause.status !== 400) return {};
		const body = cause.body;
		if (!body || typeof body !== 'object') return {};
		const problems: Record<string, string> = {};
		for (const [key, value] of Object.entries(body as Record<string, unknown>)) {
			if (typeof value === 'string') problems[key] = value;
		}
		return problems;
	}

	async function remove(value: IgnoreRuleDraft) {
		if (!value.id) return;
		saving = true;
		try {
			await deleteIgnoreRule(value.id);
			rules = rules.filter((rule) => rule.id !== value.id);
			draft = null;
			pane = 'list';
			// Nothing it ignored comes back: those rows stay as they are and are put back
			// one by one, like any other ignored row.
			toast.success('Rule deleted. Transactions it ignored stay ignored.');
		} catch {
			toast.error('Could not delete the rule');
		} finally {
			saving = false;
		}
	}

	async function confirmApply() {
		if (!draft?.id) return;
		applying = true;
		try {
			const result = await applyIgnoreRule(draft.id);
			applyOpen = false;
			toast.success(`Ignored ${result.ignored_count} transactions`);
			await load();
			const reopened = rules.find((rule) => rule.id === draft?.id);
			if (reopened) draft = draftFromRule(reopened);
			schedulePreview(draft);
		} catch {
			toast.error('Could not apply the rule');
		} finally {
			applying = false;
		}
	}

	const notYetIgnored = $derived(preview?.not_yet_ignored ?? 0);
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar />
	{/snippet}

	<div class="flex min-h-full flex-col gap-5 px-4 pb-6 lg:h-full lg:min-h-0 lg:px-8">
		<!-- The navigation bar carries destinations only, so the page says its own name here,
		     the same way PageContentTemplate does for the pages that use it. -->
		<div class="flex shrink-0 flex-wrap items-baseline justify-between gap-2 pt-1">
			<Heading level="h1" size="2xl" class="leading-none">Ignore rules</Heading>
		</div>
		<IgnoreRulesDesk
			{rules}
			bind:draft
			bind:pane
			{preview}
			{previewLoading}
			{previewError}
			{bankOptions}
			{loading}
			{error}
			{saving}
			{fieldErrors}
			onNew={openNew}
			onSelect={openRule}
			{onDraftChange}
			onSave={save}
			onDelete={remove}
			onApply={() => (applyOpen = true)}
			onRetry={() => void load()}
		>
			{#snippet aside()}
				<div class="shrink-0 border-b border-slate-200 bg-taupe-50 px-4 py-3">
					<Text as="p" size="sm" tone="muted">
						A rule takes effect on what you import from now on.
					</Text>
					<a
						href="/cashflow"
						class="mt-1 inline-block rounded-md text-sm text-sky-700 underline underline-offset-2 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
					>
						Back to Cashflow
					</a>
				</div>
			{/snippet}
		</IgnoreRulesDesk>
	</div>
</AppShellTemplate>

<!-- Reaching back over the ledger is the one action that changes what is already there,
     so it names its count and asks first. -->
<Dialog bind:open={applyOpen} size="md" title="Apply this rule to existing transactions">
	<div class="space-y-3">
		<Text as="p" size="md" class="text-slate-800">
			<span class="text-slate-900 tabular-nums">{notYetIgnored}</span> transactions match “{draft?.name ??
				''}” and are not ignored yet. They move out of your monthly totals and tag split.
		</Text>
		<Text as="p" size="sm" tone="muted">
			Nothing is deleted. You can show ignored rows on Cashflow and put any of them back.
			Transactions you already put back by hand are left alone.
		</Text>
	</div>
	{#snippet footer()}
		<Button
			variant="ghost"
			intent="secondary"
			shape="default"
			disabled={applying}
			onclick={() => (applyOpen = false)}
		>
			Cancel
		</Button>
		<Button shape="default" loading={applying} onclick={() => void confirmApply()}>
			Ignore {notYetIgnored} transactions
		</Button>
	{/snippet}
</Dialog>
