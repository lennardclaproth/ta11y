<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Drawer from './Drawer.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/Drawer',
		component: Drawer,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	let open = $state(false);
	let wideOpen = $state(false);
</script>

<Story name="Right slide-in" asChild>
	<div>
		<Button onclick={() => (open = true)}>Open drawer</Button>
		<Drawer bind:open title="Details">
			<p class="text-sm text-slate-700">
				A right-hand slide-in panel. Press <kbd>Esc</kbd>, click the backdrop, or the close button
				to dismiss it.
			</p>
			{#snippet footer()}
				<Button variant="ghost" intent="secondary" onclick={() => (open = false)}>Close</Button>
				<Button onclick={() => (open = false)}>Save</Button>
			{/snippet}
		</Drawer>
	</div>
</Story>

<!-- Edge to edge: no body padding, for a table that fills the panel. -->
<Story name="Edge to edge" asChild>
	<div>
		<Button onclick={() => (wideOpen = true)}>Open full table</Button>
		<Drawer bind:open={wideOpen} title="Transactions" width="max-w-[96vw]" padded={false}>
			<div class="flex flex-1 items-center justify-center bg-taupe-50 text-sm text-slate-500">
				A table fills this panel from edge to edge.
			</div>
		</Drawer>
	</div>
</Story>
