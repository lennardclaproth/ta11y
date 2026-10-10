import type { ImportResult } from '$lib/api/types';

/** Where the dialog is in its two-step flow. */
export const importPhases = ['form', 'processing', 'result'] as const;
export type ImportPhase = (typeof importPhases)[number];

/** The two destinations one upload fills, in the order the trail walks them. */
export const importDestinations = ['Cashflow', 'Portfolio'] as const;
export type ImportDestination = (typeof importDestinations)[number];

/** One destination's outcome, paired with the import record it was read from. */
export interface DestinationOutcome {
	destination: ImportDestination;
	result: ImportResult | null;
}

/** The rows the result shows per destination, in reading order. */
export const importCountRows = [
	{ key: 'total_rows', label: 'Rows read' },
	{ key: 'imported', label: 'New' },
	{ key: 'duplicates', label: 'Already imported' },
	{ key: 'failed', label: 'Failed' }
] as const;

export type ImportCountKey = (typeof importCountRows)[number]['key'];

/** Status word per destination. Colour never stands alone. */
export const destinationStatusLabels = {
	completed: 'Imported',
	partial: 'Partly imported',
	failed: 'Failed'
} as const;

export type DestinationStatus = keyof typeof destinationStatusLabels;

/**
 * The message shown when the parser refused the file. The API classifies the refusal
 * (`reason: "file_not_recognised"`); the wording of the answer belongs here.
 */
export const wrongFileMessage =
	'This is not the DEGIRO Account statement export. Export "Account statement" from DEGIRO and upload that CSV.';
