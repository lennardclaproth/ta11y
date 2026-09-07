<script lang="ts">
	import { onMount } from 'svelte';
	import { ApiError } from '$lib/api/client';
	import type { ProviderCredential } from '$lib/api/types';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import {
		listProviderCredentials,
		revealProviderApiKey,
		updateProviderCredentials
	} from '$lib/services/marketdata';
	import { toast } from '$lib/stores/toast.svelte';

	let providers = $state<ProviderCredential[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	/** Keys revealed in this session, held only in memory and cleared on reload. */
	let revealed = $state<Record<string, string>>({});
	let revealing = $state<string | null>(null);

	let editing = $state<ProviderCredential | null>(null);
	let keyInput = $state('');
	let baseUriInput = $state('');
	let saving = $state(false);
	let formError = $state('');

	async function load() {
		loading = true;
		error = null;
		try {
			providers = await listProviderCredentials();
		} catch {
			error = 'Failed to load provider credentials';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function openEditor(provider: ProviderCredential) {
		editing = provider;
		// Never prefill the key: the stored value is not readable here, and an empty
		// field makes it obvious that saving replaces whatever is configured.
		keyInput = '';
		baseUriInput = provider.base_uri ?? '';
		formError = '';
	}

	function closeEditor() {
		editing = null;
		keyInput = '';
		baseUriInput = '';
		formError = '';
	}

	async function save() {
		if (!editing) return;
		const key = keyInput.trim();
		const baseUri = baseUriInput.trim();
		const original = editing.base_uri ?? '';

		if (key === '' && baseUri === original) {
			formError = 'Enter a new API key, or change the base URL.';
			return;
		}

		saving = true;
		formError = '';
		try {
			const updated = await updateProviderCredentials(editing.id, {
				...(key === '' ? {} : { api_key: key }),
				...(baseUri === original ? {} : { base_uri: baseUri })
			});
			providers = providers.map((provider) => (provider.id === updated.id ? updated : provider));
			// A replaced key invalidates anything revealed earlier for this provider.
			delete revealed[editing.id];
			revealed = { ...revealed };
			toast.success(`${updated.name} credentials updated`);
			closeEditor();
		} catch (cause) {
			if (cause instanceof ApiError && cause.status === 409) {
				formError = 'This provider ingests uploaded files and has no credentials to configure.';
			} else if (cause instanceof ApiError && cause.status === 400) {
				formError = 'The credentials could not be saved. Check the values and try again.';
			} else {
				formError = 'Could not confirm the credentials were saved. Check your connection.';
			}
		} finally {
			saving = false;
		}
	}

	async function reveal(provider: ProviderCredential) {
		revealing = provider.id;
		try {
			const result = await revealProviderApiKey(provider.id);
			revealed = { ...revealed, [provider.id]: result.api_key };
		} catch {
			toast.error(`Could not reveal the ${provider.name} key`);
		} finally {
			revealing = null;
		}
	}

	function hide(provider: ProviderCredential) {
		delete revealed[provider.id];
		revealed = { ...revealed };
	}

	function quota(provider: ProviderCredential): string {
		if (provider.total > 0) return `${provider.remaining} of ${provider.total} left`;
		if (provider.used > 0) return `${provider.used} used`;
		return '—';
	}
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar title="Credentials" accountName="Admin" />
	{/snippet}

	<PageContentTemplate>
		<div
			class="flex shrink-0 flex-wrap items-center justify-between gap-4 border-b border-slate-200 p-4"
		>
			<div>
				<h2 class="text-2xl">External API credentials</h2>
				<p class="mt-1 max-w-prose text-sm text-slate-600">
					API keys for the market-data providers. Stored keys are never shown here — only a hint
					identifying them — so revealing one is a deliberate action.
				</p>
			</div>
			<Button variant="outline" disabled={loading} onclick={load}>Refresh</Button>
		</div>

		{#if error}
			<div role="alert" class="flex flex-wrap items-center gap-3 p-4">
				<p class="text-sm text-red-700">{error}. Check your connection and try again.</p>
				<Button variant="outline" onclick={load}>Retry loading</Button>
			</div>
		{/if}

		<DataTable
			rows={providers}
			{loading}
			getRowId={(row: ProviderCredential) => row.id}
			emptyText={error
				? 'Provider credentials are unavailable. Retry loading above.'
				: 'No providers are configured yet.'}
			columns={[
				{ key: 'name', header: 'Provider', width: 'w-40', cell: nameCell },
				{ key: 'base_uri', header: 'Base URL', cell: baseUriCell },
				{ key: 'api_key', header: 'API key', cell: keyCell },
				{
					key: 'quota',
					header: 'Requests',
					width: 'w-40',
					value: (row: ProviderCredential) => quota(row)
				},
				{ key: 'actions', header: '', width: 'w-32', align: 'right', cell: actionsCell }
			]}
		/>

		<p class="border-t border-slate-200 p-4 text-xs text-slate-500">
			Keys configured through the environment (<code>MARKETSTACK_API_KEY</code>,
			<code>ALPHA_VANTAGE_API_KEY</code>) are seeded on startup. If one is still set, restarting the
			API re-adds it alongside anything changed here.
		</p>
	</PageContentTemplate>
</AppShellTemplate>

{#snippet nameCell(row: ProviderCredential)}
	<span class="font-medium text-slate-900">{row.name}</span>
{/snippet}

{#snippet baseUriCell(row: ProviderCredential)}
	{#if row.base_uri}
		<span class="font-mono text-xs break-all text-slate-700">{row.base_uri}</span>
	{:else}
		<span class="text-slate-400">—</span>
	{/if}
{/snippet}

{#snippet keyCell(row: ProviderCredential)}
	{#if row.ingestion_mode === 'MANUAL'}
		<Badge intent="neutral" variant="soft" size="sm">Uploads only</Badge>
	{:else if !row.has_api_key}
		<Badge intent="warning" variant="soft" size="sm">Not set</Badge>
	{:else if revealed[row.id]}
		<span class="flex flex-wrap items-center gap-2">
			<span class="font-mono text-xs break-all text-slate-900">{revealed[row.id]}</span>
			<Button variant="ghost" size="sm" onclick={() => hide(row)}>Hide</Button>
		</span>
	{:else}
		<span class="flex items-center gap-2">
			<span class="font-mono text-xs text-slate-700">{row.api_key_hint}</span>
			<Button
				variant="ghost"
				size="sm"
				disabled={revealing === row.id}
				loading={revealing === row.id}
				onclick={() => reveal(row)}>Reveal</Button
			>
		</span>
	{/if}
{/snippet}

{#snippet actionsCell(row: ProviderCredential)}
	{#if row.ingestion_mode === 'MANUAL'}
		<span class="text-xs text-slate-400">No credentials</span>
	{:else}
		<Button variant="outline" size="sm" onclick={() => openEditor(row)}>
			{row.has_api_key ? 'Replace key' : 'Set key'}
		</Button>
	{/if}
{/snippet}

<Dialog
	open={editing !== null}
	title={editing ? `${editing.name} credentials` : 'Credentials'}
	dismissible={!saving}
	closeOnBackdrop={!saving}
	closeOnEscape={!saving}
	onClose={closeEditor}
>
	{#if editing}
		{@const provider = editing}
		<form
			class="space-y-4"
			novalidate
			aria-busy={saving}
			onsubmit={(event) => {
				event.preventDefault();
				void save();
			}}
		>
			<p class="text-sm text-slate-600">
				{provider.has_api_key
					? `Replaces the current key (${provider.api_key_hint}). Leave blank to keep it.`
					: 'No key is configured for this provider yet.'}
			</p>
			<fieldset disabled={saving} class="space-y-4">
				<FormField
					label="API key"
					id="provider-api-key"
					hint="Stored server-side. It is not shown again unless you reveal it."
				>
					{#snippet children(ctx)}
						<Input
							id={ctx.id}
							type="password"
							bind:value={keyInput}
							autocomplete="off"
							placeholder={provider.has_api_key ? 'Enter a new key' : 'Paste the provider API key'}
							ariaDescribedby={ctx.describedby}
						/>
					{/snippet}
				</FormField>
				<FormField label="Base URL" id="provider-base-uri">
					{#snippet children(ctx)}
						<Input
							id={ctx.id}
							bind:value={baseUriInput}
							placeholder="https://api.example.com/v2"
							ariaDescribedby={ctx.describedby}
						/>
					{/snippet}
				</FormField>
			</fieldset>
			{#if formError}<p role="alert" class="text-sm text-red-700">{formError}</p>{/if}
			<div class="flex justify-end gap-2 border-t border-slate-200 pt-4">
				<Button variant="ghost" intent="secondary" disabled={saving} onclick={closeEditor}>
					Cancel
				</Button>
				<Button type="submit" disabled={saving} loading={saving}>
					{saving ? 'Saving…' : 'Save credentials'}
				</Button>
			</div>
		</form>
	{/if}
</Dialog>
