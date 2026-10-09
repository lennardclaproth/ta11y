<script lang="ts">
	// Every finished month, one ruled line each. On a phone four fit the card without making
	// the whole rail taller than the card you are reading; the rest are one tap away.
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import MonthStandingRow from '$lib/components/molecules/month-standing-row/MonthStandingRow.svelte';
	import type { MonthStanding } from '$lib/api/types';

	type Props = {
		/** Finished months, newest first. */
		months: MonthStanding[];
		onOpenUnassigned?: (month: MonthStanding) => void;
	};

	let { months, onOpenUnassigned }: Props = $props();

	const compactLimit = 4;
	let showAllOnPhone = $state(false);

	const hidden = $derived(Math.max(0, months.length - compactLimit));
</script>

{#if months.length === 0}
	<Text size="sm" tone="muted">
		No finished month yet. The month you are in is scored once it is over.
	</Text>
{:else}
	{#each months as month, index (month.month)}
		<div class={index >= compactLimit && !showAllOnPhone ? 'hidden sm:block' : ''}>
			<MonthStandingRow {month} divider={index > 0} {onOpenUnassigned} />
		</div>
	{/each}
	{#if hidden > 0 && !showAllOnPhone}
		<span class="sm:hidden">
			<Button
				size="sm"
				variant="ghost"
				intent="secondary"
				class="mt-1 px-0"
				onclick={() => (showAllOnPhone = true)}
			>
				Show {hidden} earlier {hidden === 1 ? 'month' : 'months'}
			</Button>
		</span>
	{/if}
{/if}
