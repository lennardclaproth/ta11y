import { ApiError, apiDownload, apiUpload, saveFile } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type { BackupDataset, BackupResult } from '$lib/api/types';
import { backupFixture, backupPreviews } from '$lib/data/fixtures/backup';
import { clone, delay } from './_mock';

function today(): string {
	return new Date().toISOString().slice(0, 10);
}

/** `GET /backup/{dataset}/export` — downloads the dataset's file. */
export async function exportBackup(dataset: BackupDataset): Promise<void> {
	const fallbackName = `ta11y-${dataset}-${today()}.json`;
	if (useMocks) {
		await delay();
		const blob = new Blob([JSON.stringify(backupFixture(dataset), null, 2)], {
			type: 'application/json'
		});
		saveFile(blob, fallbackName);
		return;
	}
	const { blob, filename } = await apiDownload(`/backup/${dataset}/export`, fallbackName);
	saveFile(blob, filename);
}

/**
 * `POST /backup/{dataset}/import` (multipart). A file naming vendors or listings this ta11y
 * lacks comes back as a 409 carrying the same counts and what is missing; that is returned
 * as a result, so callers read `missing` rather than catching it. Any other refusal throws.
 */
export async function importBackup(
	dataset: BackupDataset,
	file: File,
	dryRun: boolean
): Promise<BackupResult> {
	if (useMocks) {
		await delay();
		let kind: unknown;
		try {
			kind = (JSON.parse(await file.text()) as { kind?: unknown }).kind;
		} catch {
			throw new ApiError(422, 'Not a readable backup', { file: 'not a readable ta11y backup' });
		}
		if (kind !== dataset) {
			throw new ApiError(422, 'Wrong kind of file', {
				file: `backup: wrong kind of file: holds "${String(kind)}", expected "${dataset}"`
			});
		}
		return { ...clone(backupPreviews[dataset]), dry_run: dryRun };
	}

	const form = new FormData();
	form.append('file', file);
	form.append('dry_run', String(dryRun));
	try {
		return await apiUpload<BackupResult>(`/backup/${dataset}/import`, form);
	} catch (error) {
		if (error instanceof ApiError && error.status === 409) {
			return error.body as BackupResult;
		}
		throw error;
	}
}
