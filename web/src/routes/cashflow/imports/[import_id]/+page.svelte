<script lang="ts">
	// What one import did: the three counts first, then the rows its rules ignored, grouped
	// under the rule that caught them. Judging a rule on its own harvest is what makes a
	// rule that is too wide visible — and it is the one moment where undoing it is cheap.
	//
	// The page lays out its own content region instead of using PageContentTemplate: that
	// template gives narrow screens a fixed 32rem scroll box, which suits a ledger but cuts
	// a reading page.
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import CountStat from '$lib/components/molecules/count-stat/CountStat.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import IgnoredByRuleGroup from '$lib/components/organisms/ignored-by-rule-group/IgnoredByRuleGroup.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Spinner from '$lib/components/atoms/spinner/Spinner.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { getImport } from '$lib/services/importer';
	import { getImportIgnored, updateIgnoreRule } from '$lib/services/ignoreRules';
	import {
		ignoreCashflowTransactionsByFilter,
		ignoreCashflowTransactionsBySelection,
		listCashflowTransactions
	} from '$lib/services/cashflow';
	import { ruleToRequest } from '$lib/api/ignoreRules';
	import { toast } from '$lib/stores/toast.svelte';
	import type { CashflowTransaction, IgnoredRuleGroup, ImportResult } from '$lib/api/types';

	/** How often an import still being processed is re-read. */
	const POLL_MS = 1500;
	/** Stop polling rather than spin forever if an import never reaches a terminal state. */
	const POLL_TIMEOUT_MS = 120_000;

	const importId = $derived(page.params.import_id ?? '');

	let result = $state<ImportResult | null>(null);
	let groups = $state<IgnoredRuleGroup[]>([]);
	let loading = $state(true);
	let groupsLoading = $state(true);
	let error = $state<string | null>(null);
	let groupsError = $state<string | null>(null);
	let busyRuleId = $state<string | null>(null);

	const running = $derived(
		result === null || result.status === 'pending' || result.status === 'processing'
	);
	const failed = $derived(result?.status === 'failed');

	const counts = $derived([
		{
			key: 'new',
			label: 'New',
			value: result?.imported ?? 0,
			note: 'added to your ledger'
		},
		{
			key: 'duplicates',
			label: 'Duplicates',
			value: result?.duplicates ?? 0,
			note: 'already imported earlier'
		},
		{
			key: 'ignored',
			label: 'Auto-ignored',
			value: result?.auto_ignored ?? 0,
			note: 'kept out of your totals'
		}
	]);

	const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

	/**
	 * Whether this page is gone. The poll below outlives a click on "Back to Cashflow" at
	 * the bottom of this very page, and would keep asking for up to two minutes — and
	 * twice over if the review is opened again.
	 */
	let gone = false;

	/**
	 * Follows the import to the end. Processing is detached from the upload, so the page
	 * fills in when the import reports it is done rather than claiming a figure it cannot
	 * know yet.
	 */
	async function load() {
		loading = true;
		error = null;
		const deadline = Date.now() + POLL_TIMEOUT_MS;
		try {
			while (!gone && Date.now() < deadline) {
				const current = await getImport(importId);
				if (gone) return;
				result = current;
				if (current.status === 'completed' || current.status === 'failed') break;
				await sleep(POLL_MS);
			}
		} catch {
			if (gone) return;
			error = 'This import could not be read. It may belong to another account.';
		} finally {
			if (!gone) loading = false;
		}
		if (gone) return;
		if (result?.status === 'completed') await loadGroups();
		else groupsLoading = false;
	}

	async function loadGroups() {
		groupsLoading = true;
		groupsError = null;
		try {
			const loaded = await getImportIgnored(importId);
			if (gone) return;
			groups = loaded;
		} catch {
			if (gone) return;
			groupsError = 'What this import ignored could not be read.';
		} finally {
			if (!gone) groupsLoading = false;
		}
	}

	onMount(() => {
		void load();
		return () => {
			gone = true;
		};
	});

	/** Replaces one group's rows in place, so a restore does not reload the whole page. */
	function patchGroup(ruleId: string, transactions: CashflowTransaction[]) {
		groups = groups.map((group) =>
			group.rule.id === ruleId ? { ...group, transactions } : group
		);
	}

	async function restore(group: IgnoredRuleGroup, row: CashflowTransaction) {
		busyRuleId = group.rule.id;
		try {
			await ignoreCashflowTransactionsBySelection({ ignored: false, ids: [row.id] });
			patchGroup(
				group.rule.id,
				group.transactions.map((tx) => (tx.id === row.id ? { ...tx, ignored: false } : tx))
			);
			toast.success('Put back in your ledger. Rules leave it alone from now on.');
		} catch {
			toast.error('Could not restore the transaction');
		} finally {
			busyRuleId = null;
		}
	}

	async function restoreAll(group: IgnoredRuleGroup) {
		busyRuleId = group.rule.id;
		try {
			await ignoreCashflowTransactionsByFilter({
				ignored: false,
				filters: { import_id: importId, ignored_by_rule: group.rule.id }
			});
			patchGroup(
				group.rule.id,
				group.transactions.map((tx) => ({ ...tx, ignored: false }))
			);
			toast.success(`Put ${group.total} transactions back in your ledger`);
		} catch {
			toast.error('Could not restore this group');
		} finally {
			busyRuleId = null;
		}
	}

	/**
	 * Switching a rule off and emptying its harvest stay two actions. Turning it off says
	 * "not again"; what it already ignored keeps its state and is restored separately.
	 */
	async function turnOff(group: IgnoredRuleGroup) {
		busyRuleId = group.rule.id;
		try {
			const updated = await updateIgnoreRule(group.rule.id, {
				...ruleToRequest(group.rule),
				enabled: false
			});
			groups = groups.map((entry) =>
				entry.rule.id === group.rule.id ? { ...entry, rule: updated } : entry
			);
			toast.success('Rule switched off. What it already ignored stays ignored.');
		} catch {
			toast.error('Could not switch the rule off');
		} finally {
			busyRuleId = null;
		}
	}

	async function showMore(group: IgnoredRuleGroup) {
		busyRuleId = group.rule.id;
		try {
			const rest = await listCashflowTransactions({
				import_id: importId,
				ignored_by_rule: group.rule.id,
				limit: 200,
				sort_by: 'date',
				sort_order: 'desc'
			});
			patchGroup(group.rule.id, rest.data);
		} catch {
			toast.error('Could not load the rest of this group');
		} finally {
			busyRuleId = null;
		}
	}
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar />
	{/snippet}

	<div class="flex min-h-full flex-col px-4 pb-6 lg:h-full lg:min-h-0 lg:px-8">
		<article class="flex flex-col bg-white lg:min-h-0 lg:flex-1">
			<header class="shrink-0 border-b border-slate-200 px-4 py-4 lg:px-8">
				<div class="flex flex-wrap items-center gap-3">
					<Heading level="h1" size="2xl" class="text-slate-900">
						{#if failed}Import failed{:else if running}Importing…{:else}Import complete{/if}
					</Heading>
					<!-- A word, not only a spinner: the state has to survive reduced motion. -->
					{#if running && !error}
						<Badge intent="info" variant="soft" size="sm">Still running</Badge>
						<Spinner size="sm" />
					{/if}
				</div>
				<Text as="p" size="sm" tone="muted" class="mt-1">
					{#if error}
						Nothing could be read about this import.
					{:else if running}
						Reading the file
					{:else}
						{(result?.total_rows ?? 0).toLocaleString('en')} rows read
					{/if}
				</Text>

				<dl class="mt-4 grid grid-cols-1 gap-x-8 gap-y-3 sm:grid-cols-3">
					{#each counts as count (count.key)}
						<CountStat
							label={count.label}
							value={count.value}
							note={count.note}
							loading={running && !error}
						/>
					{/each}
				</dl>
			</header>

			<div class="min-h-0 space-y-4 px-4 py-4 lg:flex-1 lg:overflow-auto lg:px-8">
				{#if error}
					<Alert intent="error" title="Could not read this import">{error}</Alert>
				{:else if failed}
					<Alert intent="warning" title="This import did not finish">
						{result?.reason === 'file_not_recognised'
							? 'The file is not an export this bank produces. Upload the CSV your bank exported.'
							: 'Something went wrong while reading the file. The counts above are what landed before it stopped.'}
					</Alert>
				{:else if running}
					<Text as="p" size="sm" tone="muted">
						This page fills in as soon as the import reports it is done. You can leave it open.
					</Text>
					{#each [0, 1] as index (index)}
						<section class="border-t border-slate-400 pt-3">
							<Skeleton class="mb-3 w-56" />
							<Skeleton class="mb-2 w-full" />
							<Skeleton class="w-full" />
						</section>
					{/each}
				{:else if groupsError}
					<Alert intent="error" title="Could not load what was ignored">
						{groupsError} The import itself finished; nothing was lost.
						<Button
							size="sm"
							variant="outline"
							intent="secondary"
							shape="default"
							class="mt-2"
							onclick={() => void loadGroups()}
						>
							Try again
						</Button>
					</Alert>
				{:else if groupsLoading}
					<Skeleton class="w-56" />
				{:else if groups.length === 0}
					<Alert intent="info" title="Nothing was ignored automatically">
						None of your rules matched a row in this import. Everything new is in your ledger.
					</Alert>
				{:else}
					<div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
						<Heading level="h2" size="lg" class="text-slate-900">
							Ignored automatically ({result?.auto_ignored ?? 0})
						</Heading>
						<Text as="span" size="sm" tone="muted">
							Grouped by the rule that caught them · {groups.length}
							{groups.length === 1 ? 'rule' : 'rules'} in this import
						</Text>
					</div>

					{#each groups as group (group.rule.id)}
						<IgnoredByRuleGroup
							{group}
							busy={busyRuleId === group.rule.id}
							onRestore={(row) => void restore(group, row)}
							onRestoreAll={() => void restoreAll(group)}
							onTurnOff={() => void turnOff(group)}
							onShowMore={() => void showMore(group)}
						/>
					{/each}
				{/if}
			</div>

			<footer
				class="flex shrink-0 flex-wrap items-center justify-end gap-2 border-t border-slate-200 px-4 py-3 lg:px-8"
			>
				<Button
					variant="ghost"
					intent="secondary"
					shape="default"
					onclick={() => void goto('/cashflow/ignore-rules')}
				>
					Manage ignore rules
				</Button>
				<Button shape="default" onclick={() => void goto('/cashflow')}>Back to Cashflow</Button>
			</footer>
		</article>
	</div>
</AppShellTemplate>
