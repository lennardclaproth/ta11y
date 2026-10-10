<script lang="ts">
	// The safety net the whole feature rests on: what a rule catches, shown before it is
	// allowed to catch anything. The count above the sample is what says how wide the rule
	// is; the sample is there so a match that obviously does not belong is recognisable.
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { CashflowTransaction } from '$lib/api/types';

	type Props = {
		/** Everything in the ledger the rule matches. */
		matching: number;
		/** The size of the ledger the rule was held against. */
		scanned: number;
		/** The most recent matches, as a sample — not a page. */
		sample: CashflowTransaction[];
		/** The text the rule matches on, named in the "nothing matches" state. */
		contains?: string;
		loading?: boolean;
		error?: string | null;
		class?: string;
	};

	let {
		matching,
		scanned,
		sample,
		contains = '',
		loading = false,
		error = null,
		class: className = ''
	}: Props = $props();

	const classes = $derived(
		['border-t border-slate-400 pt-2', className].filter(Boolean).join(' ')
	);

	function day(value: string): string {
		return formatDisplayDate(value.slice(0, 10));
	}
</script>

<section class={classes} aria-label="Matching transactions" aria-busy={loading}>
	<div class="mb-1 flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
		<Heading level="h3" size="sm" class="text-slate-900">
			{#if loading}
				Checking what this rule matches…
			{:else}
				Matches {matching.toLocaleString('en')} of {scanned.toLocaleString('en')} transactions
			{/if}
		</Heading>
		{#if !loading && !error && sample.length > 0}
			<Text as="span" size="sm" tone="muted">Showing the {sample.length} most recent</Text>
		{/if}
	</div>

	{#if error}
		<Alert intent="error" title="Could not check this rule">{error}</Alert>
	{:else if loading}
		{#each [0, 1, 2] as index (index)}
			<Skeleton class="my-1.5" />
		{/each}
	{:else if matching === 0}
		<!-- Blank would read as broken. A rule that catches nothing yet is a legitimate
		     rule — it is written for next month's import — so it says so and offers the way out. -->
		<Alert intent="info" title="Nothing matches this rule yet">
			No transaction in your ledger contains “{contains}”. Save it anyway to catch future
			imports, or widen the text.
		</Alert>
	{:else}
		<ul class="text-sm">
			{#each sample as match (match.id)}
				<li class="border-b border-slate-200 py-1.5 last:border-b-0">
					<!-- Narrow: the description gets the width, the date reads as a caption. -->
					<div class="sm:hidden">
						<div class="flex items-baseline justify-between gap-3">
							<span class="min-w-0 flex-1 truncate text-slate-800">{match.description}</span>
							<span class="shrink-0">
								<Money amount={scaledToNumber(match.amountCents)} currency="EUR" size="sm" />
							</span>
						</div>
						<span class="text-xs text-slate-500 tabular-nums">{day(match.date)}</span>
					</div>

					<div class="hidden items-baseline gap-3 sm:flex">
						<span class="w-24 shrink-0 text-slate-500 tabular-nums">{day(match.date)}</span>
						<span class="min-w-0 flex-1 truncate text-slate-800">{match.description}</span>
						<span class="w-28 shrink-0 text-right">
							<Money amount={scaledToNumber(match.amountCents)} currency="EUR" size="sm" />
						</span>
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</section>
