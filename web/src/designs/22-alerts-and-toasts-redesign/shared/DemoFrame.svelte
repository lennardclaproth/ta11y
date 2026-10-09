<script lang="ts">
	/**
	 * Static stand-in for the Cashflow page (masthead + ledger toolbar + records), used only to
	 * show a notification in its real surroundings. Not a component proposal.
	 */
	import type { Snippet } from 'svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import { mockRows } from './mock-data';

	type Props = {
		/** In-flow band directly under the masthead (variant C). */
		banner?: Snippet;
		/** Persistent notice inside the content panel, above the records. */
		notice?: Snippet;
		/** Floating layer, positioned inside the frame instead of the viewport. */
		overlay?: Snippet;
		class?: string;
	};

	let { banner, notice, overlay, class: className = '' }: Props = $props();
</script>

<div
	class={['relative overflow-hidden border border-slate-300 bg-taupe-100 text-slate-800', className]
		.filter(Boolean)
		.join(' ')}
>
	<header class="px-4 pt-3 pb-4 lg:px-6">
		<div class="flex items-center justify-between gap-4 border-t-2 border-b border-slate-800 py-2">
			<span class="font-heading text-4xl leading-none tracking-tight text-slate-900">ta11y</span>
			<nav class="hidden items-center gap-6 md:flex" aria-hidden="true">
				<span class="border-b-2 border-amber-500 py-3 text-sm font-semibold text-slate-950"
					>Cashflow</span
				>
				<span class="border-b-2 border-transparent py-3 text-sm text-slate-700">Assets</span>
				<span class="border-b-2 border-transparent py-3 text-sm text-slate-700">Portfolio</span>
			</nav>
		</div>
		<div class="flex flex-wrap items-center justify-between gap-3 pt-4">
			<h2 class="font-heading text-4xl leading-none tracking-tight text-slate-900">Cashflow</h2>
			<span
				class="h-10 rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-500"
				>Search description…</span
			>
		</div>
	</header>

	{#if banner}
		<div class="px-4 pb-4 lg:px-6">{@render banner()}</div>
	{/if}

	<div class="px-4 pb-5 lg:px-6">
		<div class="flex flex-col bg-taupe-50">
			<div
				class="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200 px-4 py-3"
			>
				<h3 class="font-heading text-2xl text-slate-900">Transactions</h3>
				<span
					class="inline-flex h-10 items-center rounded-md border border-amber-200 bg-slate-600 px-3 text-sm text-amber-200"
					>+ Add transaction</span
				>
			</div>

			{#if notice}
				<div class="border-b border-slate-200 px-4 py-3">{@render notice()}</div>
			{/if}

			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-slate-200 text-xs tracking-wider text-slate-500 uppercase">
						<th scope="col" class="px-4 py-2 text-left font-medium">Date</th>
						<th scope="col" class="px-4 py-2 text-left font-medium">Description</th>
						<th scope="col" class="hidden px-4 py-2 text-left font-medium sm:table-cell">Tag</th>
						<th scope="col" class="px-4 py-2 text-right font-medium">Amount</th>
					</tr>
				</thead>
				<tbody>
					{#each mockRows as row (row.description)}
						<tr class="border-b border-slate-200/70">
							<td class="px-4 py-2.5 whitespace-nowrap text-slate-500">{row.date}</td>
							<td class="px-4 py-2.5 text-slate-800">{row.description}</td>
							<td class="hidden px-4 py-2.5 text-slate-500 sm:table-cell">{row.tag}</td>
							<td class="px-4 py-2.5 text-right">
								<Money amount={row.amount} size="sm" colored />
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>

	{#if overlay}
		{@render overlay()}
	{/if}
</div>
