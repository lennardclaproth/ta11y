<script lang="ts">
	// Initial load of the standing: the shape of the card, not a spinner and not a zero — a zero
	// here would read as "you put nothing aside".
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';

	type Props = {
		/** `month` mirrors the running-month card, `rows` the monthly-standing list. */
		shape?: 'month' | 'rows' | 'streak';
	};

	let { shape = 'month' }: Props = $props();
</script>

<div class="flex flex-col gap-3" aria-busy="true">
	<span class="sr-only">Loading your monthly standing</span>

	{#if shape === 'rows'}
		{#each [0, 1, 2, 3, 4, 5] as row (row)}
			<div class="flex items-center gap-3">
				<Skeleton width="4.5rem" />
				<Skeleton class="h-2 flex-1" variant="rect" />
				<Skeleton width="2.5rem" />
				<Skeleton width="6rem" />
			</div>
		{/each}
	{:else if shape === 'streak'}
		<Skeleton variant="rect" height="2rem" width="10rem" />
		<div class="flex gap-2">
			{#each [0, 1, 2, 3, 4, 5] as mark (mark)}
				<Skeleton variant="circle" height="2rem" width="2rem" />
			{/each}
		</div>
		<Skeleton width="80%" />
	{:else}
		<Skeleton variant="rect" height="2rem" width="14rem" />
		<Skeleton class="h-3" variant="rect" />
		<div class="flex justify-end gap-6">
			<Skeleton width="7rem" />
			<Skeleton width="7rem" />
			<Skeleton width="7rem" />
		</div>
		<Skeleton width="60%" />
	{/if}
</div>
