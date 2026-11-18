<script lang="ts" generics="T extends RankedBarDatum">
	// Amounts as bars ranked longest first: length compares better than slices, and every row
	// stays visible instead of collapsing into "+N more". With `onSelect` a row is a toggle
	// button, so the chart can drive a filter.
	import type { Snippet } from 'svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import { chartColors } from '$lib/charts/theme';
	import type { RankedBarDatum } from './charts.types';

	type Props = {
		data: T[];
		loading?: boolean;
		error?: string | null;
		/** Bar colours by rank, cycled; a datum's own `color` wins. */
		ramp?: readonly string[];
		formatValue?: (value: number) => string;
		/** Show each row's share of the total beside its amount. */
		showShare?: boolean;
		/** The selected row's key; drawn pressed. */
		selected?: string | null;
		/** Makes rows toggle buttons. Called with the row's key, or `null` when it is unselected. */
		onSelect?: (key: string | null) => void;
		/** Extra right-hand cells per row (a growth figure, a sparkline). */
		extra?: Snippet<[T]>;
		ariaLabel?: string;
		class?: string;
	};

	let {
		data,
		loading = false,
		error = null,
		ramp = [chartColors.net],
		formatValue = (value: number) => value.toLocaleString(),
		showShare = true,
		selected = null,
		onSelect,
		extra,
		ariaLabel = 'Ranked amounts',
		class: className = ''
	}: Props = $props();

	const rows = $derived([...data].sort((a, b) => b.value - a.value));
	const max = $derived(Math.max(...rows.map((r) => r.value), 0));
	const total = $derived(rows.reduce((sum, r) => sum + r.value, 0));
	const shareOf = (value: number) =>
		`${total > 0 ? ((value / total) * 100).toLocaleString('en', { maximumFractionDigits: 1 }) : 0}%`;
	const rowClass =
		'flex w-full items-center gap-3 rounded-sm px-1 py-1.5 text-left text-sm focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none';
</script>

{#snippet cells(row: T, i: number)}
	<!-- The label gives way before the figures do, so a narrow card never pushes the share
	     outside the row — and with it outside the button's clickable area. -->
	<span class="w-28 min-w-12 shrink truncate text-slate-700">{row.label}</span>
	<span class="h-2.5 min-w-6 flex-1 rounded-full bg-slate-100">
		<span
			class="block h-full rounded-full"
			style:width="{max > 0 ? (row.value / max) * 100 : 0}%"
			style:background-color={row.color ?? ramp[i % ramp.length]}
		></span>
	</span>
	<span class="w-20 shrink-0 text-right text-slate-900 tabular-nums">{formatValue(row.value)}</span>
	{#if showShare}
		<span class="w-12 shrink-0 text-right text-xs text-slate-500 tabular-nums">
			{shareOf(row.value)}
		</span>
	{/if}
	{@render extra?.(row)}
{/snippet}

<div class={['min-w-0', className].filter(Boolean).join(' ')}>
	{#if loading}
		<div class="flex flex-col gap-3">
			{#each [0, 1, 2, 3] as i (i)}
				<Skeleton variant="rect" class="h-5 w-full rounded" />
			{/each}
		</div>
	{:else if error}
		<p class="rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
			{error}
		</p>
	{:else if rows.length === 0}
		<p class="py-6 text-center text-sm text-slate-500">No data</p>
	{:else}
		<ul class="flex flex-col" aria-label={ariaLabel}>
			{#each rows as row, i (row.key)}
				<li>
					{#if onSelect}
						<button
							type="button"
							aria-pressed={selected === row.key}
							class={[rowClass, selected === row.key ? 'bg-slate-100' : 'hover:bg-slate-50'].join(
								' '
							)}
							onclick={() => onSelect(selected === row.key ? null : row.key)}
						>
							{@render cells(row, i)}
						</button>
					{:else}
						<div class={rowClass}>{@render cells(row, i)}</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
