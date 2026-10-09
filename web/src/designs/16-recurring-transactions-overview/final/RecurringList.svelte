<script lang="ts">
	// The same items, stacked. A five-column ledger cannot be read on a 390px screen without
	// pushing the amount and the date off to the side, and those two are the whole point of the
	// page — so below `lg` the row unfolds instead of scrolling sideways.
	// Proposal: molecule `recurring-item-row`, paired with the table in the organism.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import AmountTrend from './AmountTrend.svelte';
	import { rhythmLabels, type RecurringItem } from '../recurring.fixture';
	import { nextExpectedLabel } from '../recurring.format';

	type Props = {
		rows: RecurringItem[];
		loading?: boolean;
		error?: string | null;
		emptyText: string;
		onOpen?: (row: RecurringItem) => void;
		class?: string;
	};

	let { rows, loading = false, error = null, emptyText, onOpen, class: className = '' }: Props =
		$props();
</script>

<div class={['min-h-0 flex-1 overflow-y-auto', className].filter(Boolean).join(' ')}>
	{#if loading}
		{#each [0, 1, 2, 3, 4] as index (index)}
			<div class="flex flex-col gap-2 border-b border-slate-100 px-4 py-3">
				<Skeleton width="60%" />
				<Skeleton width="40%" />
			</div>
		{/each}
	{:else if error}
		<div class="px-4 py-8">
			<div
				class="mx-auto max-w-md rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-center text-sm text-red-700"
			>
				{error}
			</div>
		</div>
	{:else if rows.length === 0}
		<p class="px-6 py-10 text-center text-sm text-slate-500">{emptyText}</p>
	{:else}
		<ul>
			{#each rows as row (row.id)}
				<li class="border-b border-slate-100">
					<button
						type="button"
						class="flex w-full flex-col gap-1 px-4 py-3 text-left transition-colors hover:bg-slate-50 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-slate-500"
						onclick={() => onOpen?.(row)}
					>
						<span class="flex items-baseline justify-between gap-3">
							<span class="min-w-0 truncate font-medium text-slate-900">{row.name}</span>
							<Money amount={row.amount} currency="EUR" size="md" />
						</span>

						<span class="flex items-center justify-between gap-3">
							<span class="flex min-w-0 items-center gap-2">
								<Badge intent="neutral" variant="soft" size="sm">{rhythmLabels[row.rhythm]}</Badge>
								<span class="truncate text-xs text-slate-500">{row.linked} transactions</span>
							</span>
							<span class="flex shrink-0 items-center gap-2">
								<Money
									amount={row.history[0]}
									currency="EUR"
									size="sm"
									weight="normal"
									class="text-xs text-slate-500"
								/>
								<AmountTrend
									data={row.history}
									width={56}
									height={20}
									ariaLabel={`Amounts for ${row.name}, oldest to newest`}
								/>
							</span>
						</span>

						<span class="text-xs text-slate-700">
							{row.endedFrom ? `Ended from ${row.endedFrom}` : nextExpectedLabel(row.nextExpected)}
						</span>
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
