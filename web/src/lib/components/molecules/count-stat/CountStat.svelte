<script lang="ts">
	// A labelled figure under a rule: how many rows were new, duplicate, or auto-ignored.
	// It is deliberately not StatCard — that one renders money, and a count is a different
	// kind of number: never coloured by sign, never given a currency, always tabular so a
	// row of them lines up.
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import type { CountStatSize } from './count-stat.types';
	import { countStatSkeletonClasses, countStatValueClasses } from './count-stat.variants';

	type Props = {
		label: string;
		value: number;
		/** Short line under the figure saying what it counts. */
		note?: string;
		size?: CountStatSize;
		/** The figure is not known yet; a placeholder stands in rather than a zero. */
		loading?: boolean;
		class?: string;
	};

	let { label, value, note, size = 'lg', loading = false, class: className = '' }: Props = $props();

	const classes = $derived(
		['flex flex-col gap-0.5 border-t border-slate-400 pt-2', className].filter(Boolean).join(' ')
	);
</script>

<div class={classes}>
	<dt class="text-sm text-slate-500">{label}</dt>
	{#if loading}
		<!-- A zero would be a claim; while the import is still running there is no number. -->
		<dd><Skeleton variant="rect" class={countStatSkeletonClasses[size]} /></dd>
	{:else}
		<dd class={countStatValueClasses[size]}>{value.toLocaleString('en')}</dd>
	{/if}
	{#if note}
		<Text as="span" size="sm" tone="muted">{note}</Text>
	{/if}
</div>
