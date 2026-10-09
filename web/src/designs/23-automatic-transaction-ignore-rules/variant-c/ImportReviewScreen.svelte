<script lang="ts">
	// Variant C — "Import review". The import's own result page is the centre of the feature: the
	// counts are read first, then the auto-ignored rows grouped under the rule that caught them, so
	// a rule that is too wide is judged on its own harvest instead of row by row.
	//
	// The page lays out its own content region instead of using PageContentTemplate: that template
	// gives narrow screens a fixed 32rem scroll box, which suits a ledger but cuts a reading page.
	import PreviewShell from '../PreviewShell.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import {
		ignoreRules,
		ignoredTransactions,
		lastImport,
		ruleSummary,
		type IgnoreRule,
		type IgnoredTransaction
	} from '../ignore-rules.fixture';

	type Props = {
		/** The import finished, but no rule matched anything. */
		empty?: boolean;
		loading?: boolean;
		error?: string | null;
		/** Lay the shell out in document flow (narrow-screen previews). */
		flow?: boolean;
	};

	let { empty = false, loading = false, error = null, flow = false }: Props = $props();

	type RuleGroup = { rule: IgnoreRule; rows: IgnoredTransaction[] };

	const groups = $derived<RuleGroup[]>(
		empty
			? []
			: ignoreRules
					.filter((rule) => rule.ignoredLastImport > 0)
					.map((rule) => ({
						rule,
						rows: ignoredTransactions.filter((row) => row.ruleId === rule.id).slice(0, 2)
					}))
					.filter((group) => group.rows.length > 0)
					.slice(0, 2)
	);

	const counts = [
		{ key: 'new', label: 'New', value: lastImport.imported, note: 'added to your ledger' },
		{
			key: 'duplicates',
			label: 'Duplicates',
			value: lastImport.duplicates,
			note: 'already imported earlier'
		},
		{
			key: 'ignored',
			label: 'Auto-ignored',
			value: lastImport.autoIgnored,
			note: 'kept out of your totals'
		}
	];
</script>

<PreviewShell {flow}>
	{#snippet top()}
		<TopNavbar title="Cashflow" accountName="Lennard Claproth" accountEmail="lennard@example.com" />
	{/snippet}

	<div class="flex min-h-full flex-col px-4 pb-6 lg:h-full lg:min-h-0 lg:px-8">
		<article class="flex flex-col bg-white lg:min-h-0 lg:flex-1">
			<header class="shrink-0 border-b border-slate-200 px-4 py-4 lg:px-8">
				<Heading level="h1" size="2xl" class="text-slate-900">Import complete</Heading>
				<Text as="p" size="sm" tone="muted" class="mt-1">
					{lastImport.file} · {lastImport.account} · {lastImport.totalRows} rows read
				</Text>

				<!-- Counts are figures, not money: tabular, labelled, never coloured by sign. -->
				<dl class="mt-4 grid grid-cols-1 gap-x-8 gap-y-3 sm:grid-cols-3">
					{#each counts as count (count.key)}
						<div class="flex flex-col gap-0.5 border-t border-slate-400 pt-2">
							<dt class="text-sm text-slate-500">{count.label}</dt>
							{#if loading}
								<Skeleton class="h-8 w-16" />
							{:else}
								<dd class="font-heading text-3xl leading-tight text-slate-900 tabular-nums">
									{count.value}
								</dd>
							{/if}
							<Text as="span" size="sm" tone="muted">{count.note}</Text>
						</div>
					{/each}
				</dl>
			</header>

			<div class="min-h-0 space-y-4 px-4 py-4 lg:flex-1 lg:overflow-auto lg:px-8">
				{#if error}
					<Alert intent="error" title="Could not load what was ignored">
						{error} The import itself finished; nothing was lost.
					</Alert>
				{:else if loading}
					{#each [0, 1] as index (index)}
						<section class="border-t border-slate-400 pt-3">
							<Skeleton class="mb-3 w-56" />
							<Skeleton class="mb-2 w-full" />
							<Skeleton class="w-full" />
						</section>
					{/each}
				{:else if groups.length === 0}
					<Alert intent="info" title="Nothing was ignored automatically">
						None of your rules matched a row in this import. Everything new is in your ledger.
					</Alert>
				{:else}
					<div class="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
						<Heading level="h2" size="lg" class="text-slate-900">
							Ignored automatically ({lastImport.autoIgnored})
						</Heading>
						<Text as="span" size="sm" tone="muted">
							Grouped by the rule that caught them · 4 rules in this import
						</Text>
					</div>

					{#each groups as group (group.rule.id)}
						<section class="border-t border-slate-400 pt-2" aria-label={group.rule.name}>
							<div class="mb-1 flex flex-wrap items-start justify-between gap-x-4 gap-y-2">
								<div class="min-w-0">
									<div class="flex flex-wrap items-center gap-2">
										<Heading level="h3" size="md" class="text-slate-900">{group.rule.name}</Heading>
										<Badge intent="neutral" variant="soft" size="sm">
											{group.rule.ignoredLastImport} rows
										</Badge>
									</div>
									<Text as="p" size="sm" tone="muted">{ruleSummary(group.rule)}</Text>
								</div>

								<div class="flex shrink-0 flex-wrap items-center gap-2">
									<Button size="sm" variant="ghost" intent="secondary" shape="default">
										Turn rule off
									</Button>
									<Button size="sm" variant="outline" intent="secondary" shape="default">
										Restore all ({group.rule.ignoredLastImport})
									</Button>
								</div>
							</div>

							<ul class="text-sm">
								{#each group.rows as row (row.id)}
									<li class="border-b border-slate-200 py-1.5 last:border-b-0">
										<!-- Narrow: description and amount lead, the rest reads as a caption. -->
										<div class="sm:hidden">
											<div class="flex items-baseline justify-between gap-3">
												<span class="min-w-0 flex-1 truncate text-slate-800">{row.description}</span>
												<span class="shrink-0">
													<Money
														amount={scaledToNumber(row.amountCents)}
														currency="EUR"
														size="sm"
													/>
												</span>
											</div>
											<div class="mt-0.5 flex items-center justify-between gap-3">
												<span class="text-xs text-slate-500">
													{formatDisplayDate(row.date.slice(0, 10))} ·
													{row.direction === 'in' ? 'Incoming' : 'Outgoing'}
												</span>
												<Button size="sm" variant="ghost" intent="secondary" shape="default">
													Restore
												</Button>
											</div>
										</div>

										<div class="hidden items-center justify-between gap-4 sm:flex">
											<span class="w-24 shrink-0 text-slate-500 tabular-nums">
												{formatDisplayDate(row.date.slice(0, 10))}
											</span>
											<span class="min-w-0 flex-1 truncate text-slate-800">{row.description}</span>
											<Badge
												intent={row.direction === 'in' ? 'success' : 'error'}
												variant="soft"
												size="sm"
											>
												{row.direction === 'in' ? 'In' : 'Out'}
											</Badge>
											<span class="w-24 shrink-0 text-right">
												<Money
													amount={scaledToNumber(row.amountCents)}
													currency="EUR"
													size="sm"
												/>
											</span>
											<Button size="sm" variant="ghost" intent="secondary" shape="default">
												Restore
											</Button>
										</div>
									</li>
								{/each}
							</ul>

							{#if group.rows.length < group.rule.ignoredLastImport}
								<Text as="p" size="sm" tone="muted" class="pt-1.5">
									Show {group.rule.ignoredLastImport - group.rows.length} more in this group
								</Text>
							{/if}
						</section>
					{/each}
				{/if}
			</div>

			<footer
				class="flex shrink-0 flex-wrap items-center justify-end gap-2 border-t border-slate-200 px-4 py-3 lg:px-8"
			>
				<Button variant="ghost" intent="secondary" shape="default">Manage ignore rules</Button>
				<Button shape="default">Back to Cashflow</Button>
			</footer>
		</article>
	</div>
</PreviewShell>
