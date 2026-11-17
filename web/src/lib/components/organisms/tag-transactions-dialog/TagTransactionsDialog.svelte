<script lang="ts">
	// Tagging a selection from Cashflow: find one of your tags (or start a new one) on the left,
	// read what it changes on the right. A transaction carries one tag, so picking one replaces
	// what each row has now — which is why the change is shown per row before it is saved.
	import { untrack } from 'svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { scaledToNumber } from '$lib/api/money';
	import type { CashflowTag, CashflowTransaction } from '$lib/api/types';
	import { tagPreviewLimit, type TagTransactionsValue } from './tag-transactions-dialog.types';

	type Props = {
		open?: boolean;
		/** The transactions the user selected on Cashflow. */
		selection: CashflowTransaction[];
		/** Every tag the account has used, most used first. */
		tags: CashflowTag[];
		/** The tag list is still being read. */
		tagsLoading?: boolean;
		saving?: boolean;
		error?: string | null;
		onSubmit?: (value: TagTransactionsValue) => void;
	};

	let {
		open = $bindable(false),
		selection,
		tags,
		tagsLoading = false,
		saving = false,
		error = null,
		onSubmit
	}: Props = $props();

	let query = $state('');
	/** The chosen tag; `''` takes the tag off, `null` is nothing chosen yet. */
	let picked = $state<string | null>(null);

	// Every opening starts empty: a tag picked for the last selection says nothing about this one.
	$effect(() => {
		if (!open) return;
		untrack(() => {
			query = '';
			picked = null;
		});
	});

	const needle = $derived(query.trim().toLowerCase());
	const matches = $derived(tags.filter((entry) => entry.tag.toLowerCase().includes(needle)));
	// Tags compare without case, as the ledger's tag filter does, so typing "Groceries" picks the
	// existing "groceries" rather than starting a second one.
	const existing = $derived(tags.find((entry) => entry.tag.toLowerCase() === needle));
	const canCreate = $derived(needle !== '' && !existing);

	const changes = $derived(
		picked === null ? 0 : selection.filter((row) => row.tag.trim() !== picked).length
	);
	const previewRows = $derived(selection.slice(0, tagPreviewLimit));
	const hiddenRows = $derived(selection.length - previewRows.length);

	const canSubmit = $derived(picked !== null && changes > 0 && !saving);
	const submitLabel = $derived(
		picked === null
			? 'Choose a tag'
			: picked === ''
				? `Remove tag from ${changes}`
				: `Tag ${changes} as “${picked}”`
	);

	function pick(tag: string) {
		picked = tag;
	}

	function countLabel(count: number): string {
		return `${count} ${count === 1 ? 'transaction' : 'transactions'}`;
	}

	function optionClasses(selected: boolean): string {
		return [
			'flex min-h-11 w-full items-center gap-2 border-b border-slate-200 px-4 text-left text-sm transition-colors focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-slate-500 disabled:cursor-not-allowed',
			selected ? 'bg-amber-100 text-slate-900' : 'text-slate-800 hover:bg-white'
		].join(' ');
	}

	function submit() {
		if (!canSubmit || picked === null) return;
		onSubmit?.({ tag: picked });
	}
</script>

<Dialog
	bind:open
	title={`Tag ${selection.length} ${selection.length === 1 ? 'transaction' : 'transactions'}`}
	size="2xl"
	closeOnEscape={!saving}
>
	<div class="-mx-5 -my-4 flex flex-col md:flex-row">
		<!-- The index: your tags, searchable, plus starting a new one and taking the tag off. -->
		<div
			class="flex flex-col border-b border-slate-200 bg-taupe-50 md:w-72 md:shrink-0 md:border-r md:border-b-0"
		>
			<div class="flex flex-col gap-1.5 px-4 pt-4 pb-3">
				<Text as="span" size="sm" weight="medium">Find or create a tag</Text>
				<SearchInput
					bind:value={query}
					debounceMs={0}
					ariaLabel="Find or create a tag"
					placeholder="groceries, rent, …"
					disabled={saving}
				/>
			</div>

			<div
				role="listbox"
				aria-label="Your tags"
				class="max-h-48 overflow-y-auto border-t border-slate-400 md:max-h-80"
			>
				{#if canCreate}
					<button
						type="button"
						role="option"
						aria-selected={picked === query.trim()}
						class={optionClasses(picked === query.trim())}
						disabled={saving}
						onclick={() => pick(query.trim())}
					>
						<Icon icon="heroicons:plus" size="sm" />
						<span class="min-w-0 truncate text-sky-700">New tag “{query.trim()}”</span>
					</button>
				{/if}

				{#if tagsLoading}
					<div class="flex flex-col gap-3 px-4 py-3">
						<Skeleton width="70%" />
						<Skeleton width="55%" />
						<Skeleton width="62%" />
					</div>
				{:else}
					{#each matches as entry (entry.tag)}
						<button
							type="button"
							role="option"
							aria-selected={picked === entry.tag}
							class={optionClasses(picked === entry.tag)}
							disabled={saving}
							onclick={() => pick(entry.tag)}
						>
							<span class="min-w-0 flex-1 truncate">{entry.tag}</span>
							<span class="text-xs text-slate-500 tabular-nums" title={countLabel(entry.count)}>
								{entry.count}
							</span>
						</button>
					{/each}
					{#if tags.length === 0 && !canCreate}
						<p class="px-4 py-3 text-sm text-slate-500">
							No tags yet. Type one above to start your first.
						</p>
					{:else if matches.length === 0 && needle !== '' && !canCreate}
						<p class="px-4 py-3 text-sm text-slate-500">No tag contains “{query.trim()}”.</p>
					{/if}
				{/if}

				<button
					type="button"
					role="option"
					aria-selected={picked === ''}
					class={optionClasses(picked === '')}
					disabled={saving}
					onclick={() => pick('')}
				>
					<span class="text-slate-600 italic">No tag — remove it</span>
				</button>
			</div>
		</div>

		<!-- What the pick changes, row by row, so a tag that gets replaced is seen before it is. -->
		<div class="flex min-w-0 flex-1 flex-col gap-2 px-5 pt-4 pb-5">
			<div class="flex items-baseline justify-between gap-3">
				<h3 class="font-heading text-xl text-slate-900">What changes</h3>
				<Text as="span" size="xs" tone="muted">
					{picked === null ? 'Pick a tag on the left' : `${changes} of ${selection.length} change`}
				</Text>
			</div>

			<div class="max-h-72 overflow-auto">
				<table class="w-full border-collapse text-sm">
					<thead class="sticky top-0 bg-white">
						<tr class="border-b border-slate-400 text-xs font-semibold text-slate-700">
							<th scope="col" class="py-2 pr-3 text-left">Date</th>
							<th scope="col" class="py-2 pr-3 text-left">Description</th>
							<th scope="col" class="py-2 pr-3 text-right">Amount</th>
							<th scope="col" class="py-2 pr-3 text-left">Tag now</th>
							<th scope="col" class="py-2 text-left">After</th>
						</tr>
					</thead>
					<tbody>
						{#each previewRows as row (row.id)}
							{@const now = row.tag.trim()}
							<tr class="border-b border-slate-100">
								<td class="py-2 pr-3 whitespace-nowrap text-slate-700">
									{formatDisplayDate(row.date.slice(0, 10))}
								</td>
								<td class="max-w-48 truncate py-2 pr-3 text-slate-700">{row.description}</td>
								<td class="py-2 pr-3 text-right">
									<Money amount={scaledToNumber(row.amountCents)} currency="EUR" size="sm" />
								</td>
								<td class="py-2 pr-3 text-slate-500">{now || 'Untagged'}</td>
								<td class="py-2">
									{#if picked === null}
										<span class="text-slate-400" aria-label="Not chosen yet">—</span>
									{:else if now === picked}
										<span class="text-xs text-slate-500">unchanged</span>
									{:else if picked === ''}
										<Badge intent="neutral" variant="soft" size="sm">Untagged</Badge>
									{:else}
										<Badge intent="primary" variant="solid" size="sm">{picked}</Badge>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
			{#if hiddenRows > 0}
				<Text size="xs" tone="muted">
					And {hiddenRows} more, changed the same way.
				</Text>
			{/if}

			{#if error}
				<NoticeBand intent="error" surface="inset" gutter="dialog" class="-mx-5">{error}</NoticeBand
				>
			{/if}

			<Text size="xs" tone="muted">
				One tag per transaction: a new tag replaces the one it has. The tag split on Cashflow
				follows; your monthly totals do not change.
			</Text>
		</div>
	</div>

	{#snippet footer()}
		<Button
			variant="ghost"
			intent="secondary"
			shape="default"
			disabled={saving}
			onclick={() => (open = false)}
		>
			Cancel
		</Button>
		<Button shape="default" loading={saving} disabled={!canSubmit} onclick={submit}>
			{submitLabel}
		</Button>
	{/snippet}
</Dialog>
