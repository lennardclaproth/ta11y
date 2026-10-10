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

<Story name="Tones" asChild>
	<NoticeBandRegion />
	<div class="flex flex-wrap gap-2 p-2">
		<Button intent="info" onclick={() => toast.info('Heads up — something happened.')}>Info</Button>
		<Button intent="success" onclick={() => toast.success('Saved successfully.')}>Success</Button>
		<Button intent="warning" onclick={() => toast.warning('Double-check this.')}>Warning</Button>
		<Button intent="error" onclick={() => toast.error('That did not work.')}>Error</Button>
	</div>
</Story>

<!-- Four pushed, three bands: the store holds at most three so the page below does not jump by an
     arbitrary height. The fourth does not wait its turn — the oldest ("Import scheduled") is
     dismissed as it arrives, so no notice counts down unseen. -->
<Story name="Stacked (capped at three)" asChild>
	<NoticeBandRegion />
	<div class="flex flex-wrap gap-2 p-2">
		<Button onclick={stack}>Trigger 4 notices</Button>
		<Button variant="ghost" intent="secondary" onclick={() => toast.clear()}>Clear all</Button>
	</div>
</Story>

<Story name="Auto-dismiss vs sticky" asChild>
	<NoticeBandRegion />
	<div class="flex flex-wrap gap-2 p-2">
		<Button onclick={() => toast.success('Auto-dismisses in 4.5s')}>Auto-dismiss</Button>
		<Button
			intent="warning"
			onclick={() => toast.warning('Stays until dismissed', { duration: 0, title: 'Sticky' })}
		>
			Sticky (duration 0)
		</Button>
	</div>
</Story>

<Story name="With title + message" asChild>
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
</Story>
