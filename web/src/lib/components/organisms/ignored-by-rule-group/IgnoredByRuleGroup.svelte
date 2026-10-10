<script lang="ts">
	// What one rule caught in one import. Grouping the rows under the rule that caught them
	// is what lets a rule be judged on its own harvest instead of row by row: a rule that is
	// too wide shows up as a group full of things that do not belong, and can be switched
	// off or emptied from its own header.
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import { ruleSummary } from '$lib/api/ignoreRules';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { CashflowTransaction, IgnoredRuleGroup } from '$lib/api/types';

	type Props = {
		group: IgnoredRuleGroup;
		/** A mutation on this group is in flight; its actions short-circuit until it lands. */
		busy?: boolean;
		/** Show the rest of the group. Absent when every row is already on screen. */
		onShowMore?: (group: IgnoredRuleGroup) => void;
		onRestore?: (transaction: CashflowTransaction) => void;
		onRestoreAll?: (group: IgnoredRuleGroup) => void;
		onTurnOff?: (group: IgnoredRuleGroup) => void;
		class?: string;
	};

	let {
		group,
		busy = false,
		onShowMore,
		onRestore,
		onRestoreAll,
		onTurnOff,
		class: className = ''
	}: Props = $props();

	const classes = $derived(
		['border-t border-slate-400 pt-2', className].filter(Boolean).join(' ')
	);

	const hidden = $derived(Math.max(0, group.total - group.transactions.length));

	function day(value: string): string {
		return formatDisplayDate(value.slice(0, 10));
	}

	/** A row put back by hand is finished with: no rule touches it again, so it offers no action. */
	function restored(row: CashflowTransaction): boolean {
		return !row.ignored;
	}
</script>

{#snippet rowAction(row: CashflowTransaction)}
	{#if restored(row)}
		<span class="text-xs text-emerald-700">Restored · rules leave it alone</span>
	{:else}
		<Button
			size="sm"
			variant="ghost"
			intent="secondary"
			shape="default"
			disabled={busy}
			onclick={() => onRestore?.(row)}
		>
			Restore
		</Button>
	{/if}
{/snippet}

<section class={classes} aria-label={group.rule.name} aria-busy={busy}>
	<div class="mb-1 flex flex-wrap items-start justify-between gap-x-4 gap-y-2">
		<div class="min-w-0">
			<div class="flex flex-wrap items-center gap-2">
				<Heading level="h3" size="md" class="text-slate-900">{group.rule.name}</Heading>
				<Badge intent="neutral" variant="soft" size="sm">
					{group.total}
					{group.total === 1 ? 'row' : 'rows'}
				</Badge>
				{#if !group.rule.enabled}
					<Badge intent="neutral" variant="outline" size="sm">Off</Badge>
				{/if}
			</div>
			<Text as="p" size="sm" tone="muted">{ruleSummary(group.rule)}</Text>
		</div>

		<div class="flex shrink-0 flex-wrap items-center gap-2">
			{#if group.rule.enabled}
				<!-- Switching a rule off and emptying its harvest stay two actions: turning it
				     off says "not again", restoring says "not these". -->
				<Button
					size="sm"
					variant="ghost"
					intent="secondary"
					shape="default"
					disabled={busy}
					onclick={() => onTurnOff?.(group)}
				>
					Turn rule off
				</Button>
			{/if}
			<Button
				size="sm"
				variant="outline"
				intent="secondary"
				shape="default"
				disabled={busy}
				onclick={() => onRestoreAll?.(group)}
			>
				Restore all ({group.total})
			</Button>
		</div>
	</div>

	<ul class="text-sm">
		{#each group.transactions as row (row.id)}
			<li class="border-b border-slate-200 py-1.5 last:border-b-0">
				<!-- Narrow: description and amount lead, the rest reads as a caption. -->
				<div class="sm:hidden">
					<div class="flex items-baseline justify-between gap-3">
						<span class="min-w-0 flex-1 truncate text-slate-800">{row.description}</span>
						<span class="shrink-0">
							<Money amount={scaledToNumber(row.amountCents)} currency="EUR" size="sm" />
						</span>
					</div>
					<div class="mt-0.5 flex flex-wrap items-center justify-between gap-x-3">
						<span class="text-xs whitespace-nowrap text-slate-500">
							{day(row.date)} · {row.direction === 'in' ? 'Incoming' : 'Outgoing'}
						</span>
						{@render rowAction(row)}
					</div>
				</div>

				<div class="hidden items-center gap-4 sm:flex">
					<span class="w-24 shrink-0 text-slate-500 tabular-nums">{day(row.date)}</span>
					<span class="min-w-0 flex-1 truncate text-slate-800">{row.description}</span>
					<span class="w-20 shrink-0">
						<Badge intent={row.direction === 'in' ? 'success' : 'error'} variant="soft" size="sm">
							{row.direction === 'in' ? 'In' : 'Out'}
						</Badge>
					</span>
					<span class="w-28 shrink-0 text-right">
						<Money amount={scaledToNumber(row.amountCents)} currency="EUR" size="sm" />
					</span>
					<span class="flex w-44 shrink-0 justify-end">{@render rowAction(row)}</span>
				</div>
			</li>
		{/each}
	</ul>

	{#if hidden > 0}
		<div class="pt-1.5">
			<Button
				size="sm"
				variant="ghost"
				intent="secondary"
				shape="default"
				disabled={busy}
				onclick={() => onShowMore?.(group)}
			>
				Show {hidden} more in this group
			</Button>
		</div>
	{/if}
</section>
