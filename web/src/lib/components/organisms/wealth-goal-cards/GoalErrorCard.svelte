<script lang="ts">
	// The standing could not be loaded. Say what failed and offer the retry; never fall back
	// to a zero percentage, which would read as a missed month.
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';

	type Props = {
		message?: string;
		onRetry?: () => void;
	};

	let {
		message = 'Your monthly standing could not be loaded, so this month is not scored yet. Your transactions and marks are unchanged.',
		onRetry
	}: Props = $props();
</script>

<NoticeBand intent="error" gutter="none" title="Could not load your monthly standing">
	{message}
	{#snippet action()}
		<Button size="sm" variant="outline" intent="secondary" onclick={onRetry}>Try again</Button>
	{/snippet}
</NoticeBand>
