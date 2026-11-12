<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import NoticeBandRegion from './NoticeBandRegion.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/NoticeBandRegion',
		component: NoticeBandRegion,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	import { onMount } from 'svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { toast } from '$lib/stores/toast.svelte';

	// Isolate each story from the shared singleton.
	onMount(() => {
		toast.clear();
		return () => toast.clear();
	});

	function stack() {
		toast.info('Import scheduled', { title: 'Background job' });
		toast.success('Transaction saved');
		toast.warning('Balance is low');
		toast.error('Sync failed — retrying');
	}
</script>

<!-- Each story stands in for the app shell's content area: the positioned box the stack floats in
     the top-right corner of, over the page rather than in its flow. -->

<Story name="Tones" asChild>
	<div class="relative min-h-80 bg-taupe-100">
		<NoticeBandRegion />
		<div class="flex flex-wrap gap-2 p-2">
			<Button intent="info" onclick={() => toast.info('Heads up — something happened.')}
				>Info</Button
			>
			<Button intent="success" onclick={() => toast.success('Saved successfully.')}>Success</Button>
			<Button intent="warning" onclick={() => toast.warning('Double-check this.')}>Warning</Button>
			<Button intent="error" onclick={() => toast.error('That did not work.')}>Error</Button>
		</div>
	</div>
</Story>

<!-- Four pushed, three toasts: the store holds at most three so the stack does not cover more of
     the page than it is worth. The fourth does not wait its turn — the oldest ("Import scheduled")
     is dismissed as it arrives, so no notice counts down unseen. -->
<Story name="Stacked (capped at three)" asChild>
	<div class="relative min-h-80 bg-taupe-100">
		<NoticeBandRegion />
		<div class="flex flex-wrap gap-2 p-2">
			<Button onclick={stack}>Trigger 4 notices</Button>
			<Button variant="ghost" intent="secondary" onclick={() => toast.clear()}>Clear all</Button>
		</div>
	</div>
</Story>

<!-- Push the sticky one first, then **Trigger 4 notices**: the cap counts and drops only the toasts
     that count down, so the sticky one is still there afterwards — `duration: 0` means it goes when
     someone dismisses it and not before. Hovering an auto-dismissing toast holds its countdown. -->
<Story name="Auto-dismiss vs sticky" asChild>
	<div class="relative min-h-80 bg-taupe-100">
		<NoticeBandRegion />
		<div class="flex flex-wrap gap-2 p-2">
			<Button onclick={() => toast.success('Auto-dismisses in 4.5s')}>Auto-dismiss</Button>
			<Button
				intent="warning"
				onclick={() => toast.warning('Stays until dismissed', { duration: 0, title: 'Sticky' })}
			>
				Sticky (duration 0)
			</Button>
			<Button variant="outline" intent="secondary" onclick={stack}>Trigger 4 notices</Button>
		</div>
	</div>
</Story>

<Story name="With title + message" asChild>
	<div class="relative min-h-80 bg-taupe-100">
		<NoticeBandRegion />
		<div class="flex flex-wrap gap-2 p-2">
			<Button
				intent="info"
				onclick={() =>
					toast.fromStatus('import.completed', 'Imported 128 transactions from ING.', {
						title: 'Import complete'
					})}
			>
				From status
			</Button>
		</div>
	</div>
</Story>
