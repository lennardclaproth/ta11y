<script lang="ts">
	// Renaming something: one field, prefilled with the current name. The caller does the saving
	// and throws on a refusal, so the dialog stays open with the reason under the field.
	import { untrack } from 'svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';

	type Props = {
		open?: boolean;
		title: string;
		label?: string;
		/** The current name, put in the field on every open. */
		value: string;
		/** Saves the new name; a thrown error keeps the dialog open and is shown. */
		onSave: (name: string) => Promise<void>;
	};

	let { open = $bindable(false), title, label = 'Name', value, onSave }: Props = $props();

	let name = $state('');
	let saving = $state(false);
	let error = $state<string | null>(null);

	$effect(() => {
		if (open) {
			untrack(() => {
				name = value;
				error = null;
			});
		}
	});

	async function save() {
		const trimmed = name.trim();
		if (trimmed === '') {
			error = 'Enter a name.';
			return;
		}
		saving = true;
		error = null;
		try {
			await onSave(trimmed);
			open = false;
		} catch (err) {
			const body = err instanceof Error && 'body' in err ? err.body : null;
			const reason =
				body && typeof body === 'object'
					? ((body as Record<string, unknown>).name ?? (body as Record<string, unknown>).error)
					: null;
			error = typeof reason === 'string' ? reason : 'The name could not be saved. Try again.';
		} finally {
			saving = false;
		}
	}
</script>

<!-- Closing while the request is on its way would save the new name behind the user's back,
     or write the refusal into a dialog nobody sees — so every way out is inert while it runs. -->
<Dialog
	bind:open
	{title}
	size="sm"
	closeOnEscape={!saving}
	closeOnBackdrop={!saving}
	dismissible={!saving}
>
	<FormField {label} id="name-dialog-field" error={error ?? undefined}>
		{#snippet children(ctx)}
			<Input
				id={ctx.id}
				bind:value={name}
				ariaDescribedby={ctx.describedby}
				intent={ctx.invalid ? 'error' : 'default'}
			/>
		{/snippet}
	</FormField>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" disabled={saving} onclick={() => (open = false)}>
			Cancel
		</Button>
		<Button intent="success" loading={saving} onclick={save}>Save</Button>
	{/snippet}
</Dialog>
