/** What happened to one row when a selection was adopted into the listings. */
export type AdoptOutcomeStatus = 'added' | 'exists' | 'failed';

export interface AdoptOutcome {
	symbol: string;
	status: AdoptOutcomeStatus;
	/** Present for failures, so the row can explain itself. */
	message?: string;
}
