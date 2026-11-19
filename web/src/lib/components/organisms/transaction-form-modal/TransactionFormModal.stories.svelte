<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import TransactionFormModal from './TransactionFormModal.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/TransactionFormModal',
		component: TransactionFormModal,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	let open = $state(false);
	let openError = $state(false);
	let openSaving = $state(false);
	let lastSubmit = $state('');
</script>

<Story name="Create transaction" asChild>
	<div class="space-y-2">
		<Button onclick={() => (open = true)}>New transaction</Button>
		<TransactionFormModal bind:open onSubmit={(v) => (lastSubmit = JSON.stringify(v))} />
		<p class="text-xs text-slate-500">
			Last submit: <span class="font-mono text-slate-800">{lastSubmit || '—'}</span>
		</p>
	</div>
</Story>

<!-- A refused save keeps everything that was filled in, including the chosen date. -->
<Story name="With server error" asChild>
	<div>
		<Button onclick={() => (openError = true)}>Open (error state)</Button>
		<TransactionFormModal
			bind:open={openError}
			error="A transaction with the same amount and description already exists on 14 Jul 2026."
		/>
	</div>
</Story>

<Story name="Saving" asChild>
	<div>
		<Button onclick={() => (openSaving = true)}>Open (saving)</Button>
		<TransactionFormModal bind:open={openSaving} submitting />
	</div>
</Story>
