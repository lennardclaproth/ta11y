<script lang="ts">
	// Whether a month has anything marked as income and anything marked as going to wealth.
	// Those two marks are what make a month's score mean something: without income the share
	// has nothing to divide by, and without a contribution it can only be zero. A missing mark
	// links to the month's unmarked money in the direction that could carry it.
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { monthLabel } from '$lib/api/wealthgoal';
	import type { MonthStanding } from '$lib/api/types';

	type Props = {
		month: MonthStanding;
		/** Open the month's unmarked transactions in the direction that could carry `purpose`. */
		onOpenMissing?: (month: MonthStanding, purpose: 'income' | 'wealth') => void;
		class?: string;
	};

	let { month, onOpenMissing, class: className = '' }: Props = $props();

	const label = $derived(monthLabel(month.month));
	const marks = $derived([
		{ purpose: 'income' as const, text: 'Income', marked: month.income_cents > 0 },
		{ purpose: 'wealth' as const, text: 'Wealth', marked: month.contributed_cents > 0 }
	]);
</script>

<ul
	aria-label="Purposes marked in {label}"
	class={['flex items-center gap-3 text-xs', className].filter(Boolean).join(' ')}
>
	{#each marks as mark (mark.purpose)}
		<li>
			{#if mark.marked}
				<span class="inline-flex items-center gap-1 text-emerald-700">
					<Icon icon="heroicons:check-circle" size="sm" />
					{mark.text}
					<span class="sr-only">marked</span>
				</span>
			{:else}
				<button
					type="button"
					class="inline-flex items-center gap-1 rounded-sm text-amber-700 underline decoration-amber-300 underline-offset-2 hover:text-amber-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
					onclick={() => onOpenMissing?.(month, mark.purpose)}
				>
					<Icon icon="heroicons:minus-circle" size="sm" />
					{mark.text}
					<span class="sr-only">not marked yet, show the transactions that could be</span>
				</button>
			{/if}
		</li>
	{/each}
</ul>
