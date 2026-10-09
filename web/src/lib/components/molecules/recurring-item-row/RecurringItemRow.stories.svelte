<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import RecurringItemRow from './RecurringItemRow.svelte';
	import { numberToScaled as s } from '$lib/api/money';
	import type { RecurringItem, RecurringRhythm } from '$lib/api/types';

	function item(
		name: string,
		rhythm: RecurringRhythm,
		amounts: number[],
		nextExpected: string | null,
		endedFrom: string | null = null
	): RecurringItem {
		return {
			id: name,
			name,
			direction: 'out',
			rhythm,
			lastAmountCents: s(amounts[amounts.length - 1]),
			history: amounts.map((amount, index) => ({
				date: `2026-0${4 + index}-18T00:00:00Z`,
				amountCents: s(amount)
			})),
			last_seen: '2026-09-18T00:00:00Z',
			next_expected: nextExpected,
			linked_count: amounts.length,
			ended_from: endedFrom
		};
	}

	const { Story } = defineMeta({
		title: 'Molecules/RecurringItemRow',
		component: RecurringItemRow,
		tags: ['autodocs']
	});
</script>

<Story
	name="Playground"
	args={{ item: item('Pixel Stream', 'monthly', [9, 9, 10, 10, 12, 12], '2026-10-18T00:00:00Z') }}
/>

<!-- An amount that never moved still reads as a line, not as a fall to zero. -->
<Story name="States" asChild>
	<ul class="max-w-sm divide-y divide-slate-100 border-y border-slate-200 bg-white">
		<li>
			<RecurringItemRow
				item={item('Pixel Stream', 'monthly', [9, 10, 12], '2026-10-18T00:00:00Z')}
			/>
		</li>
		<li>
			<RecurringItemRow
				item={item('Fiber Collective', 'monthly', [45, 45, 45], '2026-10-24T00:00:00Z')}
			/>
		</li>
		<li>
			<RecurringItemRow
				item={item('Anchor Insurance', 'quarterly', [90, 96], '2026-12-01T00:00:00Z')}
			/>
		</li>
		<li><RecurringItemRow item={item('Meridian Hosting', 'yearly', [120], null)} /></li>
		<li>
			<RecurringItemRow item={item('Quarterly Review', 'monthly', [6, 8], null, '2026-08')} />
		</li>
	</ul>
</Story>

<!-- A long counterparty truncates rather than pushing the amount out of view. -->
<Story name="Long name" asChild>
	<div class="max-w-sm border-y border-slate-200 bg-white">
		<RecurringItemRow
			item={item(
				'Harbour Foundation for Coastal Preservation and Maritime History',
				'monthly',
				[10, 10, 12],
				'2026-10-20T00:00:00Z'
			)}
		/>
	</div>
</Story>
