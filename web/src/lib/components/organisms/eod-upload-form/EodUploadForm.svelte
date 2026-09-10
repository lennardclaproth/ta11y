<script lang="ts">
	import { ApiError } from '$lib/api/client';
	import type { ListingSearchRow } from '$lib/api/types';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Dropzone from '$lib/components/molecules/dropzone/Dropzone.svelte';
	import { eodCsvAccept, eodCsvColumns, eodCsvDateFormat } from './eod-upload-form.types';

	let {
		listing,
		onUpload,
		onCancel,
		uploading = $bindable(false)
	}: {
		listing: ListingSearchRow;
		/** Resolves once the API has accepted the file for processing. */
		onUpload: (file: File) => Promise<void>;
		onCancel: () => void;
		uploading?: boolean;
	} = $props();

	/** Joined here so the template needs no comparison operator inside markup. */
	const columnList = eodCsvColumns.join(', ');

	let files = $state<FileList | null>(null);
	let failure = $state('');

	const file = $derived(files?.[0] ?? null);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (uploading || !file) return;
		failure = '';
		uploading = true;
		try {
			await onUpload(file);
		} catch (cause) {
			failure = describe(cause);
		} finally {
			uploading = false;
		}
	}

	/**
	 * The API rejects an upload for four distinct reasons and they need different
	 * answers: a listing priced by an API provider will never accept a file, while a
	 * rejected file is worth retrying with a corrected one.
	 */
	function describe(cause: unknown): string {
		if (!(cause instanceof ApiError)) {
			return 'Could not confirm the upload was accepted. Check your connection and try again.';
		}
		const detail =
			cause.body && typeof cause.body === 'object'
				? ((cause.body as Record<string, string>).listing_id ??
					(cause.body as Record<string, string>).file ??
					'')
				: '';
		if (cause.status === 422) {
			if (detail.includes('not manual')) {
				return `${listing.symbol} takes its prices from ${listing.source}, so it cannot accept uploads. Only listings on a manual provider do.`;
			}
			if (detail.includes('inactive')) {
				return `${listing.symbol} is inactive, so it is not accepting price uploads.`;
			}
			return `${listing.symbol}'s provider is unavailable, so the upload was not accepted.`;
		}
		if (cause.status === 404) {
			return 'This listing no longer exists. Close this and search for it again.';
		}
		if (cause.status === 400) {
			return 'The file was rejected. Check that it is a CSV with the columns listed above.';
		}
		return 'The upload could not be accepted. Try again.';
	}
</script>

<form onsubmit={submit} novalidate class="space-y-4" aria-busy={uploading}>
	<p class="text-slate-600">
		Upload end-of-day prices for <span class="font-medium text-slate-900">{listing.symbol}</span>
		{#if listing.name}<span class="text-slate-500">— {listing.name}</span>{/if}.
	</p>

	<!-- The parser matches these headers by name and fails the whole file without them, so
	     they are stated up front rather than discovered through a rejected upload. -->
	<p class="text-xs text-slate-500">
		CSV with the columns <code class="text-slate-700">{columnList}</code>. Dates as
		<code class="text-slate-700">{eodCsvDateFormat}</code>. Each row's NAV is stored as that day's
		open, high, low and close.
	</p>

	<Dropzone
		bind:files
		accept={eodCsvAccept}
		disabled={uploading}
		label={file ? file.name : 'Drag a CSV here, or click to browse'}
		hint={file ? `${(file.size / 1024).toFixed(1)} KB` : 'One file, .csv'}
	/>

	{#if failure}<p role="alert" class="text-sm text-red-700">{failure}</p>{/if}

	<div class="flex items-center justify-end gap-2 border-t border-slate-200 pt-4">
		<Button variant="ghost" intent="secondary" disabled={uploading} onclick={onCancel}>
			Cancel
		</Button>
		<Button type="submit" disabled={uploading || !file} loading={uploading}>
			{uploading ? 'Uploading…' : 'Upload prices'}
		</Button>
	</div>
</form>
