<script lang="ts">
	// The standing could not be loaded. Say what failed and offer the retry; never fall back
	// to a zero percentage, which would read as a missed month.
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
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

<div class="flex flex-col items-start gap-3">
	<Alert intent="error" title="Could not load your monthly standing">{message}</Alert>
	<Button size="sm" variant="outline" onclick={onRetry}>Try again</Button>
</div>
