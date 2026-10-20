<script lang="ts">
	// Asking before something is removed for good. The caller does the work and throws on a
	// refusal; the dialog stays open and says why, so "nothing happened" is never silent.
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import NoticeBand from '$lib/components/molecules/notice-band/NoticeBand.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';

	type Props = {
		open?: boolean;
		title: string;
		/** What will happen, in a sentence. */
		message: string;
		confirmLabel?: string;
		cancelLabel?: string;
		onConfirm: () => Promise<void>;
	};

	let {
		open = $bindable(false),
		title,
		message,
		confirmLabel = 'Delete',
		cancelLabel = 'Keep it',
		onConfirm
	}: Props = $props();

	let busy = $state(false);
	let error = $state<string | null>(null);

	$effect(() => {
		if (open) error = null;
	});

	async function confirm() {
		busy = true;
		error = null;
		try {
			await onConfirm();
			open = false;
		} catch (err) {
			const body = err instanceof Error && 'body' in err ? err.body : null;
			const reason = body && typeof body === 'object' ? (body as Record<string, unknown>).error : null;
			error = typeof reason === 'string' ? reason : 'It could not be deleted. Nothing was changed.';
		} finally {
			busy = false;
		}
	}
</script>

<Dialog bind:open {title} size="sm" closeOnEscape={!busy}>
	{#if error}
		<NoticeBand
			intent="error"
			surface="inset"
			gutter="dialog"
			title="Nothing was deleted"
			class="-mx-5 -mt-4 mb-4"
		>
			{error}
		</NoticeBand>
	{/if}
	<Text as="p" size="sm">{message}</Text>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" disabled={busy} onclick={() => (open = false)}>
			{cancelLabel}
		</Button>
		<Button intent="error" loading={busy} onclick={confirm}>{confirmLabel}</Button>
	{/snippet}
</Dialog>
