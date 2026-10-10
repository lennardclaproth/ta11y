<script lang="ts">
	// One recurring item, stacked. A five-column ledger cannot be read on a 390px screen
	// without pushing the amount and the expected date off to the side, and those two are
	// the whole point of the page — so below `lg` the row unfolds instead of scrolling
	// sideways.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Sparkline from '$lib/components/atoms/sparkline/Sparkline.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { endedFromLabel, nextExpectedLabel, rhythmLabels } from '$lib/api/recurring';
	import type { RecurringItem } from '$lib/api/types';

	type Props = {
		item: RecurringItem;
		onOpen?: (item: RecurringItem) => void;
		class?: string;
	};

	let { item, onOpen, class: className = '' }: Props = $props();

	const amounts = $derived(item.history.map((point) => scaledToNumber(point.amountCents)));
	const firstAmount = $derived(amounts.length > 0 ? amounts[0] : 0);
	const status = $derived(
		item.ended_from
			? `Ended from ${endedFromLabel(item.ended_from)}`
			: nextExpectedLabel(item.next_expected)
	);
</script>

<button
	type="button"
	class={[
		'flex w-full flex-col gap-1 px-4 py-3 text-left transition-colors hover:bg-slate-50 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-slate-500',
		className
	]
		.filter(Boolean)
		.join(' ')}
	onclick={() => onOpen?.(item)}
>
	<span class="flex items-baseline justify-between gap-3">
		<span class="min-w-0 truncate font-medium text-slate-900">{item.name}</span>
		<Money amount={scaledToNumber(item.lastAmountCents)} currency="EUR" size="md" />
	</span>

	<span class="flex items-center justify-between gap-3">
		<span class="flex min-w-0 items-center gap-2">
			<Badge intent="neutral" variant="soft" size="sm">{rhythmLabels[item.rhythm]}</Badge>
			<span class="truncate text-xs text-slate-500">{item.linked_count} transactions</span>
		</span>

		<!-- The oldest amount, then the shape of the amounts since. Deliberately one neutral
		     tone: a rising expense is not an error and a falling one not a success. -->
		{#if amounts.length > 1}
			<span class="flex shrink-0 items-center gap-2">
				<Money
					amount={firstAmount}
					currency="EUR"
					size="sm"
					weight="normal"
					class="text-xs text-slate-500"
				/>
				<Sparkline
					data={amounts}
					tone="neutral"
					width={56}
					height={20}
					ariaLabel={`Amounts for ${item.name}, oldest to newest`}
				/>
			</span>
		{/if}
	</span>

	<span class="text-xs text-slate-700">{status}</span>
</button>
